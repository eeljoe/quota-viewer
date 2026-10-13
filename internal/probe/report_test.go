package probe

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"quota-viewer/internal/config"
	"quota-viewer/internal/fetcher"
)

// ---- 夹具 ----

// fakeFetcher 是「按 id 造记录」的注入点,用来替代真去打上游的 fetcher。
type fakeFetcher struct{ r fetcher.QuotaResult }

func (f fakeFetcher) Fetch() fetcher.QuotaResult { return f.r }

// fakeDef 造注册表条目;kind 留空即模拟 kimi/xfyun/mimo/ollama/opencode-go
// 「fetcher 自己不设 Kind」的现状。
func fakeDef(id, kind string, r fetcher.QuotaResult) fetcher.ProviderDef {
	return fetcher.ProviderDef{
		ID:   id,
		Abbr: "X",
		Kind: kind,
		Build: func(map[string]string) fetcher.Fetcher {
			return fakeFetcher{r: r}
		},
	}
}

func fakeLookup(defs map[string]fetcher.ProviderDef) Lookup {
	return func(id string) (fetcher.ProviderDef, bool) {
		d, ok := defs[id]
		return d, ok
	}
}

// testConfig 造配置:ids 全部 enabled,顺序即配置顺序。
func testConfig(extended bool, ids ...string) *config.Config {
	cfg := &config.Config{ExtendedMode: extended}
	for _, id := range ids {
		cfg.Providers = append(cfg.Providers, config.ProviderConfig{ID: id, Enabled: true})
	}
	return cfg
}

func findLine(t *testing.T, rep Report, id string) ProviderLine {
	t.Helper()
	for _, l := range rep.Providers {
		if l.ID == id {
			return l
		}
	}
	t.Fatalf("providers 里没有 id=%q 的记录:%+v", id, rep.Providers)
	return ProviderLine{}
}

func mustEncode(t *testing.T, rep Report) []byte {
	t.Helper()
	data, err := EncodeLine(rep)
	if err != nil {
		t.Fatalf("EncodeLine 失败:%v", err)
	}
	return data
}

// ---- AC-6:归一四句 + 顺序 ----

// TestBuildReport_BalanceKind_PercentFromBudget —— A1 / A2 的靶子。
// ApplyBudget 是余额型 percent 的唯一产出处;删掉它或把 Kind 归一挪到它之后,
// percent 都恒为 0,本条必红。
func TestBuildReport_BalanceKind_PercentFromBudget(t *testing.T) {
	defs := map[string]fetcher.ProviderDef{
		"balance-x": fakeDef("balance-x", fetcher.KindBalance, fetcher.QuotaResult{
			Kind: fetcher.KindBalance, Balance: 150, Currency: "CNY",
		}),
	}
	cfg := testConfig(false, "balance-x")
	cfg.Providers[0].Budget = 300

	rep := BuildReport(testNow(), cfg, fakeLookup(defs))
	line := findLine(t, rep, "balance-x")

	if line.Kind != fetcher.KindBalance {
		t.Errorf("kind: got %q, want %q", line.Kind, fetcher.KindBalance)
	}
	if line.Percent == nil {
		t.Fatal("余额型 percent 为 null;ApplyBudget 未生效")
	}
	if *line.Percent != 50 {
		t.Errorf("percent: got %v, want 50(预算 300 扣掉余额 150)", *line.Percent)
	}
	if line.RemainingPercent == nil || *line.RemainingPercent != 50 {
		t.Errorf("remaining_percent: got %v, want 50", line.RemainingPercent)
	}
	if line.Balance == nil || *line.Balance != 150 {
		t.Errorf("balance: got %v, want 150", line.Balance)
	}
	if line.Currency == nil || *line.Currency != "CNY" {
		t.Errorf("currency: got %v, want CNY", line.Currency)
	}
}

// TestBuildReport_EmptyKind_NormalizedToUsage —— A6 的前半。
// kimi/xfyun/mimo/ollama/opencode-go 五家连 Kind 都不设,必须归一成 "usage"。
func TestBuildReport_EmptyKind_NormalizedToUsage(t *testing.T) {
	ids := []string{"kimi", "xfyun", "mimo", "ollama", "opencode-go"}
	defs := map[string]fetcher.ProviderDef{}
	for _, id := range ids {
		defs[id] = fakeDef(id, "", fetcher.QuotaResult{Percent: 10})
	}

	rep := BuildReport(testNow(), testConfig(true, ids...), fakeLookup(defs))
	for _, id := range ids {
		if got := findLine(t, rep, id).Kind; got != fetcher.KindUsage {
			t.Errorf("%s 的 kind: got %q, want %q", id, got, fetcher.KindUsage)
		}
	}
	if rep.EnabledCount != 5 {
		t.Errorf("enabled_count: got %d, want 5", rep.EnabledCount)
	}
}

// ---- 顶层形状 / AC-5 ----

// TestBuildReport_EnvelopeShape_Exactly7Keys 顶层恒定 7 个键、永不舍略,且值钉死。
func TestBuildReport_EnvelopeShape_Exactly7Keys(t *testing.T) {
	defs := map[string]fetcher.ProviderDef{"kimi": fakeDef("kimi", "", fetcher.QuotaResult{Percent: 26})}
	cfg := testConfig(true, "kimi", "deepseek")
	cfg.Providers[1].Enabled = false

	data := string(mustEncode(t, BuildReport(testNow(), cfg, fakeLookup(defs))))

	want := `{"schema":1,"generated_at":"2026-10-13T10:45:00+08:00","extended_mode":true,` +
		`"enabled_limit":9,"enabled_count":1,"providers":[` +
		`{"id":"kimi","kind":"usage","percent":26,"remaining_percent":74,` +
		`"reset_at":null,"balance":null,"currency":null,"error":""}],"error":""}`
	if data != want {
		t.Errorf("顶层 JSON 与契约不符:\n got %s\nwant %s", data, want)
	}
}

// TestBuildReport_OnlyEnabledProviders 未启用的 provider 不出现(AC-2)。
func TestBuildReport_OnlyEnabledProviders(t *testing.T) {
	defs := map[string]fetcher.ProviderDef{
		"kimi":     fakeDef("kimi", "", fetcher.QuotaResult{Percent: 1}),
		"deepseek": fakeDef("deepseek", fetcher.KindBalance, fetcher.QuotaResult{Balance: 1}),
	}
	cfg := testConfig(true, "kimi", "deepseek")
	cfg.Providers[1].Enabled = false

	rep := BuildReport(testNow(), cfg, fakeLookup(defs))
	if rep.EnabledCount != 1 || len(rep.Providers) != 1 || rep.Providers[0].ID != "kimi" {
		t.Fatalf("只应输出已启用的 kimi:%+v", rep.Providers)
	}
}

// TestBuildReport_TruncatesToEnabledLimit 条数按 EnabledLimit 截断(AC-2 / AC-5)。
func TestBuildReport_TruncatesToEnabledLimit(t *testing.T) {
	ids := []string{"kimi", "xfyun", "opencode-go", "mimo", "ollama"}
	defs := map[string]fetcher.ProviderDef{}
	for _, id := range ids {
		defs[id] = fakeDef(id, "", fetcher.QuotaResult{Percent: 5})
	}

	rep := BuildReport(testNow(), testConfig(false, ids...), fakeLookup(defs))
	if rep.EnabledLimit != 3 {
		t.Errorf("enabled_limit: got %d, want 3", rep.EnabledLimit)
	}
	if len(rep.Providers) != 3 || rep.EnabledCount != 3 {
		t.Fatalf("应截断到 3 条:count=%d providers=%d", rep.EnabledCount, len(rep.Providers))
	}
	got := []string{rep.Providers[0].ID, rep.Providers[1].ID, rep.Providers[2].ID}
	want := []string{"kimi", "xfyun", "opencode-go"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("顺序应 = 配置顺序: got %v, want %v", got, want)
			break
		}
	}
}

// TestBuildReport_EveryRecordHas8Keys 每条记录都含全部 8 个键(AC-2,缺键即红)。
func TestBuildReport_EveryRecordHas8Keys(t *testing.T) {
	defs := map[string]fetcher.ProviderDef{
		"kimi": fakeDef("kimi", "", fetcher.QuotaResult{Percent: 26}),
		"bad":  fakeDef("bad", "", fetcher.QuotaResult{Error: "boom"}),
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(mustEncode(t, BuildReport(testNow(), testConfig(false, "kimi", "bad"), fakeLookup(defs))), &m); err != nil {
		t.Fatalf("顶层不是合法 JSON 对象:%v", err)
	}
	if len(m) != 7 {
		t.Errorf("顶层键数: got %d, want 7(%v)", len(m), m)
	}

	var arr []map[string]json.RawMessage
	if err := json.Unmarshal(m["providers"], &arr); err != nil {
		t.Fatalf("providers 不是数组:%v", err)
	}
	if len(arr) != 2 {
		t.Fatalf("providers 条数: got %d, want 2", len(arr))
	}
	for _, want := range []string{"id", "kind", "percent", "remaining_percent", "reset_at", "balance", "currency", "error"} {
		for i, rec := range arr {
			if _, ok := rec[want]; !ok {
				t.Errorf("第 %d 条记录缺键 %q(实际键:%v)", i, want, rec)
			}
		}
		if len(arr[0]) != 8 {
			t.Errorf("每条记录应恰好 8 个键")
		}
	}
}

// ---- AC-3 / R2:失败记录形状 ----

// TestBuildReport_FailedProvider_NullScalars —— A4 的靶子。
// percent 用 float64 而非指针时,失败记录会出 0 而不是 null,本条必红。
func TestBuildReport_FailedProvider_NullScalars(t *testing.T) {
	defs := map[string]fetcher.ProviderDef{
		"bad":  fakeDef("bad", "", fetcher.QuotaResult{Percent: 42, Balance: 7, Currency: "CNY", ResetAt: "2026-10-20T08:00:00Z", Error: "上游 401"}),
		"good": fakeDef("good", "", fetcher.QuotaResult{Percent: 26, ResetAt: "2026-10-20T08:00:00Z"}),
	}

	rep := BuildReport(testNow(), testConfig(false, "bad", "good"), fakeLookup(defs))
	bad := findLine(t, rep, "bad")

	if bad.Error == "" {
		t.Error("失败记录 error 必须非空")
	}
	if bad.Kind != fetcher.KindUsage {
		t.Errorf("失败记录 kind 仍应有值: got %q", bad.Kind)
	}
	if bad.Percent != nil || bad.RemainingPercent != nil || bad.Balance != nil || bad.Currency != nil || bad.ResetAt != nil {
		t.Errorf("失败记录五个可空字段必须全 null:%+v", bad)
	}

	// 其他各家照常完整。
	good := findLine(t, rep, "good")
	if good.Error != "" || good.Percent == nil || *good.Percent != 26 {
		t.Errorf("成功的记录被拖垮了:%+v", good)
	}
	if good.ResetAt == nil || *good.ResetAt != "2026-10-20T08:00:00Z" {
		t.Errorf("成功记录 reset_at: got %v", good.ResetAt)
	}
}

// TestBuildReport_UnknownProviderID_ErrorRecord 未知 id 产出「未知平台」错误记录
// (app.go:425 同口径),不 panic、不中断其他家。
func TestBuildReport_UnknownProviderID_ErrorRecord(t *testing.T) {
	defs := map[string]fetcher.ProviderDef{"kimi": fakeDef("kimi", "", fetcher.QuotaResult{Percent: 1})}

	rep := BuildReport(testNow(), testConfig(false, "nope", "kimi"), fakeLookup(defs))
	if got := findLine(t, rep, "nope").Error; got != "未知平台" {
		t.Errorf("未知 id 的 error: got %q, want %q", got, "未知平台")
	}
	if got := findLine(t, rep, "kimi").Error; got != "" {
		t.Errorf("其他家不应受影响: got %q", got)
	}
}

// ---- AC-7:R1/R3 失败信封 ----

// TestFailReport_EmptyProvidersIsArrayNotNull —— A3 的靶子。
// providers 用 nil 切片时 JSON 出 null,本条必红。
func TestFailReport_EmptyProvidersIsArrayNotNull(t *testing.T) {
	data := string(mustEncode(t, FailReport(testNow(), "APPDATA 环境变量缺失")))

	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		t.Fatalf("失败信封不是合法 JSON:%v", err)
	}
	if len(m) != 7 {
		t.Errorf("失败信封顶层键数: got %d, want 7(%v)", len(m), m)
	}
	if got := string(m["providers"]); got != "[]" {
		t.Errorf("providers: got %s, want [](nil 切片会序列化成 null)", got)
	}

	var got struct {
		Schema       int    `json:"schema"`
		ExtendedMode bool   `json:"extended_mode"`
		EnabledLimit int    `json:"enabled_limit"`
		EnabledCount int    `json:"enabled_count"`
		Error        string `json:"error"`
	}
	if err := json.Unmarshal([]byte(data), &got); err != nil {
		t.Fatalf("失败信封字段解析失败:%v", err)
	}
	if got.Schema != 1 || got.ExtendedMode || got.EnabledLimit != 3 || got.EnabledCount != 0 {
		t.Errorf("失败信封顶层值不符 R3: %+v", got)
	}
	if got.Error == "" {
		t.Error("失败信封 error 必须带文本")
	}
}

// ---- AC-1 / AC-7:入口 Run ----

// TestRun_InvalidArgs_UsageAndExit2 R1:不带参数 / 未知参数 → usage + 退出 2。
func TestRun_InvalidArgs_UsageAndExit2(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"不带参数", nil},
		{"未知参数", []string{"--nope"}},
	}
	for _, c := range cases {
		var out, errBuf strings.Builder
		code := Run(c.args, &out, &errBuf, nil)
		if code != 2 {
			t.Errorf("%s: 退出码 got %d, want 2", c.name, code)
		}
		if out.String() != "" {
			t.Errorf("%s: 不应写 stdout, got %q", c.name, out.String())
		}
		if !strings.Contains(errBuf.String(), "--json") {
			t.Errorf("%s: stderr 应含 usage, got %q", c.name, errBuf.String())
		}
	}
}

// TestRun_JSONArgs_OneLineExit0 --json(含 -json 单横线形)出单行 JSON 且退出 0。
func TestRun_JSONArgs_OneLineExit0(t *testing.T) {
	defs := map[string]fetcher.ProviderDef{"kimi": fakeDef("kimi", "", fetcher.QuotaResult{Percent: 26})}
	for _, flag := range []string{"--json", "-json"} {
		var out, errBuf strings.Builder
		if code := Run([]string{flag}, &out, &errBuf, fakeLookup(defs)); code != 0 {
			t.Fatalf("%s: 退出码 got %d, want 0(stderr=%q)", flag, code, errBuf.String())
		}
		got := out.String()
		if strings.Count(got, "\n") != 0 {
			t.Errorf("%s: stdout 应恰好一行、内部无换行, got %q", flag, got)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(got), &m); err != nil {
			t.Fatalf("%s: stdout 不是合法 JSON:%v", flag, err)
		}
	}
}

// TestRun_NoAppData_FailureEnvelope —— A7 的靶子。
// 去掉 APPDATA 预检后 config.Load() 会吞错返回默认配置,error 变空串,本条必红。
func TestRun_NoAppData_FailureEnvelope(t *testing.T) {
	t.Setenv("APPDATA", "")

	var out, errBuf strings.Builder
	if code := Run([]string{"--json"}, &out, &errBuf, nil); code != 0 {
		t.Fatalf("配置读取失败也必须退出 0: got %d", code)
	}

	var got struct {
		Providers []json.RawMessage `json:"providers"`
		Error     string            `json:"error"`
		Schema    int               `json:"schema"`
	}
	if err := json.Unmarshal([]byte(out.String()), &got); err != nil {
		t.Fatalf("空 stdout 或非法 JSON:%q(%v)", out.String(), err)
	}
	if got.Error == "" {
		t.Error("APPDATA 缺失时 error 必须带文本,不能是成功形状")
	}
	if got.Providers == nil || len(got.Providers) != 0 {
		t.Errorf("providers 必须是初始化的空切片: got %v", got.Providers)
	}
	if got.Schema != 1 {
		t.Errorf("失败信封 schema: got %d, want 1", got.Schema)
	}
	if strings.Count(out.String(), "\n") != 0 {
		t.Errorf("stdout 应恰好一行:%q", out.String())
	}
}

// TestRun_JSONWithRealLookup_BuildsFromConfig --json 走真配置+真注册表时,
// providers 条数 = 钳制后已启用集合、且每条都在注册表里。
func TestRun_JSONWithRealLookup_BuildsFromConfig(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir()) // 隔离:用空目录的默认配置,不碰真实配置
	cfg := config.Default()
	rep := BuildReport(testNow(), cfg, DefaultLookup)

	if rep.EnabledCount != 3 {
		t.Fatalf("默认配置应启用 3 家: got %d", rep.EnabledCount)
	}
	for _, l := range rep.Providers {
		if _, ok := fetcher.Get(l.ID); !ok {
			t.Errorf("记录 id=%q 不在注册表里", l.ID)
		}
	}
	if rep.GeneratedAt != "2026-10-13T10:45:00+08:00" {
		t.Errorf("generated_at: got %q", rep.GeneratedAt)
	}
}

// ---- AC-5:generated_at 用本地偏移 ----

// TestNowLocal_CarriesLocalOffset generated_at 必须是 RFC3339 带本地偏移,不是 UTC Z。
func TestNowLocal_CarriesLocalOffset(t *testing.T) {
	now := NowLocal()
	if _, err := time.Parse(time.RFC3339, now.Format(reportTimeFormat)); err != nil {
		t.Fatalf("NowLocal 无法按 RFC3339 格式化:%v", err)
	}
	fixed := time.Date(2026, 10, 13, 10, 45, 0, 0, time.FixedZone("CST", 8*3600))
	if got := FailReport(fixed, "").GeneratedAt; got != "2026-10-13T10:45:00+08:00" {
		t.Errorf("带偏移的时刻应原样保留偏移: got %q", got)
	}
}

// ---- normalizeResetAt:无时区串按 +08:00(AC 契约 ②) ----

// TestNormalizeResetAt_NoZoneVsUTC_Differ8Hours 同一条无时区串,按 +08:00 与按 UTC
// 解释必须差 8 小时(守住 A6 的判据方向:按 UTC 解释会得到 …10:07:39Z)。
func TestNormalizeResetAt_NoZoneVsUTC_Differ8Hours(t *testing.T) {
	got := normalizeResetAt("2026-10-16 10:07:39", testNow())
	if got == nil {
		t.Fatal("无时区串应解析出时刻")
	}
	if *got == "2026-10-16T10:07:39Z" {
		t.Fatal("无时区串被按 UTC 解释了,应按 +08:00")
	}
	if *got != "2026-10-16T02:07:39Z" {
		t.Errorf("got %q, want %q", *got, "2026-10-16T02:07:39Z")
	}
}
