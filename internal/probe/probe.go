// Package probe 提供无界面探针的 JSON 契约与入口逻辑：
// 抓一轮额度 → 归一 → 一行合法 JSON(见 .scratch/json-mode/spec.md 第四节)。
package probe

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"quota-viewer/internal/config"
	"quota-viewer/internal/fetcher"
)

// schemaVersion 是顶层 schema 的固定值。
const schemaVersion = 1

// reportTimeFormat 是 generated_at 的格式(RFC3339 带本地偏移)。
const reportTimeFormat = time.RFC3339

// cstZone 是解释无时区串用的固定 +08:00 时区。
// 中国无夏令时,与 LoadLocation("Asia/Shanghai") 等价,且不依赖 tzdata 文件。
var cstZone = time.FixedZone("CST", 8*3600)

// reportTimeLayout 是无时区上游串的格式(xfyun.go:131 的写入形态)。
const reportTimeLayout = "2006-01-02 15:04:05"

// Report 是探针输出的一行 JSON 的顶层结构(成功与失败同形状)。
type Report struct {
	Schema        int            `json:"schema"`
	GeneratedAt   string         `json:"generated_at"`
	ExtendedMode  bool           `json:"extended_mode"`
	EnabledLimit  int            `json:"enabled_limit"`
	EnabledCount  int            `json:"enabled_count"`
	Providers     []ProviderLine `json:"providers"`
	Error         string         `json:"error"`
}

// ProviderLine 是每家 Provider 的记录(8 个键始终存在)。
// 数值/字符串字段用指针,否则失败记录只能是 0 / "" 而不是 null。
type ProviderLine struct {
	ID               string   `json:"id"`
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
func BuildReport(now time.Time, cfg *config.Config, get Lookup) Report {
	if get == nil {
		get = fetcher.Get
	}
	rep := Report{
		Schema:       schemaVersion,
		GeneratedAt:  now.Format(reportTimeFormat),
		ExtendedMode: cfg.ExtendedMode,
		EnabledLimit: config.EnabledLimit(cfg.ExtendedMode),
		Providers:    make([]ProviderLine, 0),
	}

	// 「已启用」= 与悬浮球一致的钳制后集合(app.go:409-418 同构)。
	enabled := make([]config.ProviderConfig, 0, rep.EnabledLimit)
	for _, p := range cfg.Providers {
		if !p.Enabled {
			continue
		}
		enabled = append(enabled, p)
		if len(enabled) == rep.EnabledLimit {
			break
		}
	}

	lines := make([]ProviderLine, len(enabled))
	var wg sync.WaitGroup
	for i, p := range enabled {
		def, ok := get(p.ID)
		if !ok {
			// app.go:425 同口径的未知平台记录。
			lines[i] = ProviderLine{ID: p.ID, Kind: fetcher.KindUsage, Error: "未知平台"}
			continue
		}
		wg.Add(1)
		go func(i int, p config.ProviderConfig, def fetcher.ProviderDef) {
			defer wg.Done()
			lines[i] = buildLine(now, p, def)
		}(i, p, def)
	}
	wg.Wait()

	rep.Providers = lines
	rep.EnabledCount = len(lines)
	return rep
}

// buildLine 抓取单家并按 app.go:431-437 的四句与顺序归一后转成契约记录。
func buildLine(now time.Time, p config.ProviderConfig, def fetcher.ProviderDef) ProviderLine {
	r := def.Build(p.Creds).Fetch()
	// 四句 + 顺序照抄 app.go:432-437:Kind 必须在 ApplyBudget 之前归一
	// (ApplyBudget 靠 Kind 判断是不是余额型,它是余额型 percent 的唯一产出处)。
	r.ID = def.ID
	r.Abbr = def.Abbr
	if r.Kind == "" {
		// 空串按 usage 归一(app.go:434-435);注册表声明了 balance 时以注册表为准
		// ——10 家里 5 家连 Kind 都不设,注册表是同源口径的落点。
		r.Kind = def.Kind
		if r.Kind == "" {
			r.Kind = fetcher.KindUsage
		}
	}
	fetcher.ApplyBudget(&r, p.Budget)

	line := ProviderLine{ID: r.ID, Kind: r.Kind, Error: r.Error}
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
		EnabledLimit: config.EnabledLimit(def.ExtendedMode),
		Providers:    make([]ProviderLine, 0),
		Error:        msg,
	}
}

// NowLocal 返回本次抓取时刻(generated_at 用,带本地偏移)。
func NowLocal() time.Time {
	return time.Now()
}

// Run 是 CLI 的入口逻辑(可测):参数不合法 → usage 打到 stderr 并返回 2;
// --json → 一行 JSON 打到 wout 并返回 0(参数/配置失败也返回 0,失败写在 error 字段里)。
func Run(args []string, wout, werr io.Writer, lookup Lookup) int {
	wantJSON := false
	for _, a := range args {
		switch a {
		case "--json", "-json":
			wantJSON = true
		default:
			writeUsage(werr, "未知参数: "+a)
			return 2
		}
	}
	if !wantJSON {
		writeUsage(werr, "缺少 --json 参数")
		return 2
	}

	now := NowLocal()
	cfg, lookupFn, err := loadConfig(lookup)
	if err != nil {
		WriteReport(wout, FailReport(now, err.Error()))
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
