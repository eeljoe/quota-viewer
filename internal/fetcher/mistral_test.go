package fetcher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mistralFetcher 起一个返回固定 body/状态码的假服务,并把 baseURL 指过去。
// hit 记录命中次数(用于断言"空凭证时一个请求都不发")。
func mistralFetcher(t *testing.T, status int, body string) (*MistralFetcher, *int) {
	t.Helper()
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path != mistralBudgetPath {
			t.Errorf("路径: got %q, want %q", r.URL.Path, mistralBudgetPath)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	f := NewMistralFetcher("sk-mistral-test")
	f.baseURL = server.URL
	return f, &hits
}

// liveBody 是 2026-11-06 真机实测的响应原文(EDU 计划,尚未消耗)。
const liveBody = `{
  "vibe_budget": {"usage_percentage": 0.00460872, "initial_budget": 150.0, "currency": "USD",
                  "reset_at": "2026-12-01T00:00:00Z", "payg_enabled": false},
  "api_budget": {"usage_percentage": 0.0, "initial_budget": 15.0, "currency": "USD",
                 "reset_at": "2026-12-01T00:00:00Z", "payg_enabled": false},
  "usage_percentage": 0.0, "initial_budget": 15.0, "currency": "USD",
  "reset_at": "2026-12-01T00:00:00Z"
}`

// AS-1 凭证为空:报错、保留三元组、一个请求都不发。
func TestMistral_EmptyKey_ReturnsErrorWithoutRequest(t *testing.T) {
	f, hits := mistralFetcher(t, 200, liveBody)
	f.apiKey = ""

	result := f.Fetch()
	if result.Error != "未配置 Mistral API Key" {
		t.Errorf("Error: got %q", result.Error)
	}
	if result.Platform != "Mistral" {
		t.Errorf("Platform: got %q, want Mistral", result.Platform)
	}
	if result.Kind != KindUsage {
		t.Errorf("Kind: got %q, want %q", result.Kind, KindUsage)
	}
	if result.LastUpdated.IsZero() {
		t.Error("LastUpdated 不得为零值")
	}
	if *hits != 0 {
		t.Errorf("空凭证不得发请求,got %d hits", *hits)
	}
}

// AS-2 请求形状:x-api-key + Accept,单次 GET。
func TestMistral_RequestShape(t *testing.T) {
	var gotKey, gotAccept, gotMethod string
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		gotKey = r.Header.Get("x-api-key")
		gotAccept = r.Header.Get("Accept")
		gotMethod = r.Method
		w.Write([]byte(liveBody))
	}))
	defer server.Close()

	f := NewMistralFetcher("sk-mistral-test")
	f.baseURL = server.URL
	f.Fetch()

	if gotMethod != "GET" {
		t.Errorf("method: got %q, want GET", gotMethod)
	}
	if gotKey != "sk-mistral-test" {
		t.Errorf("x-api-key: got %q", gotKey)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept: got %q", gotAccept)
	}
	if hits != 1 {
		t.Errorf("应恰好一次请求,got %d", hits)
	}
}

// AS-3/AS-4 200 正常:Vibe 驱动百分比与重置,Remaining 两张预算都出现。
func TestMistral_OK_VibeDrivesPercent(t *testing.T) {
	f, _ := mistralFetcher(t, 200, liveBody)
	result := f.Fetch()

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 0.00460872 {
		t.Errorf("Percent: got %v, want 0.00460872 (Vibe 预算驱动)", result.Percent)
	}
	if result.Used != result.Percent {
		t.Errorf("Used 应等于 Percent, got %v / %v", result.Used, result.Percent)
	}
	if result.Total != 100 {
		t.Errorf("Total: got %v, want 100", result.Total)
	}
	if result.ResetAt != "2026-12-01T00:00:00Z" {
		t.Errorf("ResetAt: got %q", result.ResetAt)
	}
	if result.ResetLabel != "Vibe 额度" {
		t.Errorf("ResetLabel: got %q, want Vibe 额度", result.ResetLabel)
	}
	// Vibe $0.0069 → $0.01(两位小数);API $0 → $0(整数不带小数)
	if !strings.Contains(result.Remaining, "Vibe $0.01/$150") {
		t.Errorf("Remaining 缺 Vibe 段: %q", result.Remaining)
	}
	if !strings.Contains(result.Remaining, "API $0/$15") {
		t.Errorf("Remaining 缺 API 段: %q", result.Remaining)
	}
}

// AS-3 驱动窗口必须取 Vibe 而不是 API:给两张不同的百分比,断言用的是 Vibe 那个。
func TestMistral_VibeDrivesEvenWhenAPIBigger(t *testing.T) {
	body := `{
	  "vibe_budget": {"usage_percentage": 12.5, "initial_budget": 150.0, "currency": "USD", "reset_at": "2026-12-01T00:00:00Z"},
	  "api_budget":  {"usage_percentage": 90.0, "initial_budget": 15.0,  "currency": "USD", "reset_at": "2026-11-20T00:00:00Z"}
	}`
	f, _ := mistralFetcher(t, 200, body)
	result := f.Fetch()

	if result.Percent != 12.5 {
		t.Errorf("Percent: got %v, want 12.5(Vibe)", result.Percent)
	}
	if result.ResetAt != "2026-12-01T00:00:00Z" {
		t.Errorf("ResetAt 必须跟 Vibe 走, got %q", result.ResetAt)
	}
	if result.ResetLabel != "Vibe 额度" {
		t.Errorf("ResetLabel: got %q", result.ResetLabel)
	}
}

// AS-5 Vibe 缺席 ⇒ 退回 API 驱动,标签与重置时间都换。
func TestMistral_VibeMissing_FallsBackToAPI(t *testing.T) {
	body := `{
	  "api_budget": {"usage_percentage": 40.0, "initial_budget": 15.0, "currency": "USD", "reset_at": "2026-12-01T00:00:00Z"}
	}`
	f, _ := mistralFetcher(t, 200, body)
	result := f.Fetch()

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 40.0 {
		t.Errorf("Percent: got %v, want 40(API 兜底)", result.Percent)
	}
	if result.ResetLabel != "API 额度" {
		t.Errorf("ResetLabel: got %q, want API 额度", result.ResetLabel)
	}
	if strings.Contains(result.Remaining, "Vibe") {
		t.Errorf("Vibe 缺席时不得出现 Vibe 段: %q", result.Remaining)
	}
	if !strings.Contains(result.Remaining, "API $6/$15") {
		t.Errorf("API 段金额错: %q", result.Remaining)
	}
}

// AS-5 变体:Vibe 对象存在但没有 usage_percentage(零值省略)也算不可用。
func TestMistral_VibeWithoutPercentage_FallsBackToAPI(t *testing.T) {
	body := `{
	  "vibe_budget": {"initial_budget": 150.0, "currency": "USD"},
	  "api_budget":  {"usage_percentage": 7.5, "initial_budget": 15.0, "currency": "USD"}
	}`
	f, _ := mistralFetcher(t, 200, body)
	result := f.Fetch()

	if result.Percent != 7.5 {
		t.Errorf("Percent: got %v, want 7.5(API 兜底)", result.Percent)
	}
	if result.ResetLabel != "API 额度" {
		t.Errorf("ResetLabel: got %q", result.ResetLabel)
	}
}

// AS-6 两张都不可用 ⇒ 结构错误,不返回 0% 假数据。
func TestMistral_NoBudgets_ReturnsError(t *testing.T) {
	for _, body := range []string{
		`{}`,
		`{"vibe_budget": null, "api_budget": null}`,
		`{"vibe_budget": {"initial_budget": 150.0}, "api_budget": {"initial_budget": 15.0}}`,
	} {
		f, _ := mistralFetcher(t, 200, body)
		result := f.Fetch()
		if result.Error != "响应中未找到预算数据" {
			t.Errorf("body %s: Error got %q", body, result.Error)
		}
		if result.Percent != 0 || result.Total != 0 {
			t.Errorf("body %s: 不得返回假数据, got percent=%v total=%v", body, result.Percent, result.Total)
		}
	}
}

// AS-7 401/403 与非 200 的文案分流。
func TestMistral_AuthAndHTTPErrors(t *testing.T) {
	cases := []struct {
		status int
		want   string
	}{
		{401, "API Key 无效或已过期"},
		{403, "API Key 无效或已过期"},
		{500, "HTTP 500"},
		{404, "HTTP 404"},
	}
	for _, c := range cases {
		f, _ := mistralFetcher(t, c.status, `{}`)
		result := f.Fetch()
		if result.Error != c.want {
			t.Errorf("status %d: Error got %q, want %q", c.status, result.Error, c.want)
		}
	}
}

// AS-8 非法 JSON。
func TestMistral_BadJSON_ReturnsParseError(t *testing.T) {
	f, _ := mistralFetcher(t, 200, `not json`)
	result := f.Fetch()
	if !strings.HasPrefix(result.Error, "解析响应失败") {
		t.Errorf("Error: got %q, want 解析响应失败 前缀", result.Error)
	}
}

// AS-9 越界百分比夹取到 [0,100]。
func TestMistral_PercentClamping(t *testing.T) {
	cases := map[string]float64{
		`137.5`: 100,
		`-3.0`:  0,
		`100.0`: 100,
	}
	for pct, want := range cases {
		body := `{"vibe_budget": {"usage_percentage": ` + pct + `, "initial_budget": 150.0, "currency": "USD"}}`
		f, _ := mistralFetcher(t, 200, body)
		result := f.Fetch()
		if result.Percent != want {
			t.Errorf("usage_percentage %s: Percent got %v, want %v", pct, result.Percent, want)
		}
	}
}

// AS-10 预算金额缺失/为 0 ⇒ 该侧退百分比展示,不出现 "/0"。
func TestMistral_ZeroBudget_ShowsPercentOnly(t *testing.T) {
	// 刻意避开 3.25 这类半分边界:那考的是共享 formatter 的取整方向,不是本票的契约。
	body := `{
	  "vibe_budget": {"usage_percentage": 3.2, "initial_budget": 0, "currency": "USD"},
	  "api_budget":  {"usage_percentage": 0.0, "currency": "USD"}
	}`
	f, _ := mistralFetcher(t, 200, body)
	result := f.Fetch()

	if strings.Contains(result.Remaining, "/0") {
		t.Errorf("预算为 0 时不得出现 /0: %q", result.Remaining)
	}
	if !strings.Contains(result.Remaining, "Vibe 3.2% 已用") {
		t.Errorf("Vibe 段应退化为百分比: %q", result.Remaining)
	}
	if !strings.Contains(result.Remaining, "API 0% 已用") {
		t.Errorf("API 段应退化为百分比: %q", result.Remaining)
	}
}

// 纯函数:金额格式化(整数不带小数、两位小数、未知币种、EUR 符号、负零)。
func TestMistralMoney(t *testing.T) {
	cases := []struct {
		v        float64
		currency string
		want     string
	}{
		{150, "USD", "$150"},
		{0, "USD", "$0"},
		{0.0069, "USD", "$0.01"},
		{12.345, "USD", "$12.35"},
		{150, "EUR", "€150"},
		{150, "CNY", "¥150"},
		{150, "GBP", "150"},
		{-0.001, "USD", "$0"},
	}
	for _, c := range cases {
		if got := mistralMoney(c.v, c.currency); got != c.want {
			t.Errorf("mistralMoney(%v, %q): got %q, want %q", c.v, c.currency, got, c.want)
		}
	}
}

// 纯函数:两张都不可用时 Remaining 是空串(由 applyBudget 提前报错,这里只锁纯函数契约)。
func TestMistralRemaining_EmptyWhenNoUsableBudget(t *testing.T) {
	if got := mistralRemaining(nil, nil); got != "" {
		t.Errorf("mistralRemaining(nil,nil): got %q, want 空串", got)
	}
}

// 网络失败:不可达主机必须变成 Error 而不是 panic。
func TestMistral_UnreachableHost_ReturnsError(t *testing.T) {
	f := NewMistralFetcher("sk-mistral-test")
	f.baseURL = "http://127.0.0.1:1"
	result := f.Fetch()
	if result.Error == "" {
		t.Error("不可达主机必须返回 Error")
	}
	if !strings.HasPrefix(result.Error, "请求失败") {
		t.Errorf("Error: got %q, want 请求失败 前缀", result.Error)
	}
}
