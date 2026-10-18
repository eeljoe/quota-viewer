// Package probe 提供无界面探针的 JSON 契约与入口逻辑：
// 抓一轮额度 → 归一 → 一行合法 JSON(见 .scratch/json-mode/spec.md 第四节)。
package probe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"quota-viewer/internal/config"
	"quota-viewer/internal/fetcher"
)

// schemaVersion 是顶层 schema 的固定值。
//
// 2 起:记录从"一家 Provider 一条"改为"一个监控对象一条",并新增 slot / sub_label
// 两个键(纯增量,按名取键的旧消费方不受影响)。
const schemaVersion = 2

// reportTimeFormat 是 generated_at 的格式(RFC3339 带本地偏移)。
const reportTimeFormat = time.RFC3339

// cstZone 是解释无时区串用的固定 +08:00 时区。
// 中国无夏令时,与 LoadLocation("Asia/Shanghai") 等价,且不依赖 tzdata 文件。
var cstZone = time.FixedZone("CST", 8*3600)

// reportTimeLayout 是无时区上游串的格式(xfyun.go:131 的写入形态)。
const reportTimeLayout = "2006-01-02 15:04:05"

// Report 是探针输出的一行 JSON 的顶层结构(成功与失败同形状)。
type Report struct {
	Schema       int            `json:"schema"`
	GeneratedAt  string         `json:"generated_at"`
	ExtendedMode bool           `json:"extended_mode"`
	EnabledLimit int            `json:"enabled_limit"` // 悬浮球可显示的监控对象上限(普通 9 / 扩展 16)
	EnabledCount int            `json:"enabled_count"`
	Providers    []ProviderLine `json:"providers"`
	Error        string         `json:"error"`
}

// ProviderLine 是每个监控对象的记录(10 个键始终存在)。
// 数值/字符串字段用指针,否则失败记录只能是 0 / "" 而不是 null。
//
// Slot / SubLabel 自 schema 2 起加入:一格 = 一组凭证 × 一张套餐卡,
// 多 Key / 多套餐时同一家会有多条记录,靠 slot 区分(否则消费方无从分辨哪条是哪把 Key)。
type ProviderLine struct {
	ID               string   `json:"id"`
	Slot             string   `json:"slot"`
	SubLabel         string   `json:"sub_label"`
	Kind             string   `json:"kind"`
	Percent          *float64 `json:"percent"`
	RemainingPercent *float64 `json:"remaining_percent"`
	ResetAt          *string  `json:"reset_at"`
	Balance          *float64 `json:"balance"`
	Currency         *string  `json:"currency"`
	Error            string   `json:"error"`
}

// Lookup 按 id 取注册表条目(生产用 fetcher.Get;测试注入夹具)。
type Lookup func(id string) (fetcher.ProviderDef, bool)

// BuildReport 按配置抓一轮额度并构造契约对象。
// get 为 nil 时用 fetcher.Get。
//
// 记录粒度 = 监控对象 = 一组凭证 × 一张套餐卡(与悬浮球一格一一对应),
// 抓完再按 enabled_limit 截断,顺序 = Provider 注册表顺序 → 凭证组顺序 → 套餐卡顺序。
func BuildReport(now time.Time, cfg *config.Config, get Lookup) Report {
	if get == nil {
		get = fetcher.Get
	}
	rep := Report{
		Schema:       schemaVersion,
		GeneratedAt:  now.Format(reportTimeFormat),
		ExtendedMode: cfg.ExtendedMode,
		EnabledLimit: config.MaxObjects(cfg.ExtendedMode),
		Providers:    make([]ProviderLine, 0),
	}

	// 「已启用」= 与悬浮球一致的钳制后集合(app.go fetchAll 同构)。
	enabled := make([]config.ProviderConfig, 0, len(cfg.Providers))
	for _, p := range cfg.Providers {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}

	// 每家展开成若干"凭证组"任务;一组凭证可能产出多张套餐卡 ⇒ 多条记录。
	type job struct {
		p      config.ProviderConfig
		creds  map[string]string
		credNo int
	}
	var jobs []job
	for _, p := range enabled {
		for i, creds := range p.CredSets() {
			jobs = append(jobs, job{p: p, creds: creds, credNo: i})
		}
	}

	perJob := make([][]ProviderLine, len(jobs))
	var wg sync.WaitGroup
	for i := range jobs {
		wg.Add(1)
		go func(i int, j job) {
			defer wg.Done()
			def, ok := get(j.p.ID)
			if !ok {
				// app.go fetchAll 同口径的未知平台记录。
				perJob[i] = []ProviderLine{{
					ID: j.p.ID, Slot: SlotOf(j.p.ID, j.credNo, 0, 1), Kind: fetcher.KindUsage, Error: "未知平台",
				}}
				return
			}
			perJob[i] = buildLines(now, j.p, def, j.creds, j.credNo)
		}(i, jobs[i])
	}
	wg.Wait()

	lines := make([]ProviderLine, 0, len(perJob))
	for _, ls := range perJob {
		lines = append(lines, ls...)
	}
	if len(lines) > rep.EnabledLimit {
		lines = lines[:rep.EnabledLimit]
	}
	rep.Providers = lines
	rep.EnabledCount = len(lines)
	return rep
}

// SlotOf 拼监控对象的稳定标识:<provider>/<凭证组号>[/p<套餐卡号>],与 app.go 同口径。
func SlotOf(providerID string, credNo int, planIdx int, planCount int) string {
	slot := fmt.Sprintf("%s/%d", providerID, credNo)
	if planCount > 1 {
		slot = fmt.Sprintf("%s/%d/p%d", providerID, credNo, planIdx)
	}
	return slot
}

// buildLines 抓一组凭证,按返回的每张套餐卡产出一条记录;Fetcher 返回单结果时只有一条。
func buildLines(now time.Time, p config.ProviderConfig, def fetcher.ProviderDef, creds map[string]string, credNo int) []ProviderLine {
	results := fetcher.FetchAll(def.Build(creds))
	lines := make([]ProviderLine, 0, len(results))
	for i, r := range results {
		// 四句 + 顺序照抄 app.go fetchAll:Kind 必须在 ApplyBudget 之前归一
		// (ApplyBudget 靠 Kind 判断是不是余额型,它是余额型 percent 的唯一产出处)。
		r.ID = def.ID
		if r.Kind == "" {
			// 空串按 usage 归一;注册表声明了 balance 时以注册表为准
			// ——10 家里 5 家连 Kind 都不设,注册表是同源口径的落点。
			r.Kind = def.Kind
			if r.Kind == "" {
				r.Kind = fetcher.KindUsage
			}
		}
		// 缩写/后缀/slot 消歧与 app.go 同口径(共用 SettleIdentity)。
		fetcher.SettleIdentity(&r, def, credNo, i, len(results), len(p.CredSets()))

		fetcher.ApplyBudget(&r, p.Budget)
		lines = append(lines, lineOf(now, r))
	}
	return lines
}

// lineOf 把一条归一后的结果转成契约记录。
func lineOf(now time.Time, r fetcher.QuotaResult) ProviderLine {
	line := ProviderLine{
		ID: r.ID, Slot: r.Slot, SubLabel: r.SubLabel, Kind: r.Kind, Error: r.Error,
	}
	if r.Error != "" {
		// 失败记录:数值/可空字段全 null(不是 0 / "")。
		return line
	}
	percent := r.Percent
	remaining := 100 - r.Percent
	if remaining < 0 {
		remaining = 0
	}
	line.Percent = &percent
	line.RemainingPercent = &remaining
	line.ResetAt = normalizeResetAt(r.ResetAt, now)
	if r.Kind == fetcher.KindBalance {
		balance, currency := r.Balance, r.Currency
		line.Balance = &balance
		line.Currency = &currency
	}
	return line
}

// normalizeResetAt 把上游 ResetAt 归一为 UTC 秒精度 RFC3339;
// 缺失 / 空串 / 无法解析 / 解析出的时刻不在未来(!After(now))一律 nil。
// 无时区串按 +08:00 解释(今天只有 xfyun)。
func normalizeResetAt(raw string, now time.Time) *string {
	if raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		parsed, err = time.ParseInLocation(reportTimeLayout, raw, cstZone)
		if err != nil {
			return nil
		}
	}
	if !parsed.After(now) {
		return nil
	}
	out := parsed.UTC().Format(time.RFC3339)
	return &out
}

// DefaultLookup 是生产用的查表函数。
func DefaultLookup(id string) (fetcher.ProviderDef, bool) {
	return fetcher.Get(id)
}

// EncodeLine 把契约对象序列化成一行 JSON(内部无换行)。
func EncodeLine(rep Report) ([]byte, error) {
	return json.Marshal(rep)
}

// WriteReport 序列化并整块写入 w;写失败返回 false(调用方可据此置退出码)。
func WriteReport(w io.Writer, rep Report) bool {
	data, err := EncodeLine(rep)
	if err != nil {
		return false
	}
	if _, err := w.Write(data); err != nil {
		return false
	}
	return true
}

// FailReport 构造失败信封:providers 是初始化的空切片(序列化成 [] 而非 null),
// 顶层字段齐全(R3:extended_mode / enabled_limit 用 config.Default() 的值)。
func FailReport(now time.Time, msg string) Report {
	def := config.Default()
	return Report{
		Schema:       schemaVersion,
		GeneratedAt:  now.Format(reportTimeFormat),
		ExtendedMode: def.ExtendedMode,
		EnabledLimit: config.MaxObjects(def.ExtendedMode),
		Providers:    make([]ProviderLine, 0),
		Error:        msg,
	}
}

// NowLocal 返回本次抓取时刻(generated_at 用,带本地偏移)。
func NowLocal() time.Time {
	return time.Now()
}

// Run 是 CLI 的入口逻辑(可测):
//   - 带 --json 时:一行 JSON 打到 wout 并返回 0——参数失败与配置读取失败都出
//     失败信封(spec 约束 8:参数或配置读取失败时仍输出一行同形状的合法 JSON
//     并退出 0),失败同时写 error 字段与 stderr(spec §二「诊断旁路」);
//   - 不带 --json(无参 / 未知参数):stderr 打 usage、stdout 一个字节都不写,返回 2;
//   - 写 stdout 失败(成功路径与失败信封路径一视同仁)返回 1——空 stdout + exit 0
//     是 spec 约束 8 明令禁止的形状,那是唯一的非 0 例外。
func Run(args []string, wout, werr io.Writer, lookup Lookup) int {
	wantJSON := false
	unknown := ""
	for _, a := range args {
		if a == "--json" || a == "-json" {
			wantJSON = true
			continue
		}
		if unknown == "" {
			unknown = "未知参数: " + a
		}
		// `--json=1` 这类「取值形」本身是参数失败(本开关是裸开关),但它同时声明了
		// 「这次是按 JSON 模式调用」⇒ 下面照旧走失败信封,而不是 usage + 2。
		if strings.HasPrefix(a, "--json=") || strings.HasPrefix(a, "-json=") {
			wantJSON = true
		}
	}
	if !wantJSON {
		msg := "缺少 --json 参数"
		if unknown != "" {
			msg = unknown
		}
		writeUsage(werr, msg)
		return 2
	}

	now := NowLocal()
	cfg, lookupFn, err := loadConfig(lookup)
	if err != nil || unknown != "" {
		msg := unknown
		if err != nil {
			msg = err.Error()
		}
		// 顶层失败的诊断旁路:信封之外再往 stderr 写一行(spec §二)。
		io.WriteString(werr, "quota-viewer-cli: "+msg+"\n")
		if !WriteReport(wout, FailReport(now, msg)) {
			return 1
		}
		return 0
	}

	rep := BuildReport(now, cfg, lookupFn)
	if !WriteReport(wout, rep) {
		return 1
	}
	return 0
}

// loadConfig 读取配置;lookup 非 nil 时跳过配置读取(测试用,避免碰真实配置)。
// 自己先查一次 APPDATA:config.Load() 在 APPDATA 缺失时会吞掉错误返回默认配置
// (config.go:115-118,err == nil),不预检就会输出「成功形状」的空结果。
func loadConfig(lookup Lookup) (*config.Config, Lookup, error) {
	if lookup != nil {
		return config.Default(), lookup, nil
	}
	if os.Getenv("APPDATA") == "" {
		return nil, nil, errors.New("APPDATA 环境变量缺失,无法定位配置目录")
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	return cfg, DefaultLookup, nil
}

func writeUsage(w io.Writer, msg string) {
	io.WriteString(w, "usage: quota-viewer-cli.exe --json\n"+msg+"\n")
}
