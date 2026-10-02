package fetcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// minimaxRemainsJSON 构造 coding_plan/remains 成功响应(remaining_percent 为"剩余%"语义)。
func minimaxRemainsJSON(intervalRemaining, weeklyRemaining *float64, weeklyTotal float64) string {
	return minimaxRemainsJSONTotal(intervalRemaining, weeklyRemaining, weeklyTotal, 5000000)
}

// minimaxRemainsJSONTotal 同上,但可自定义 5h 窗口总额(0 用于测未 provision 的退化路径)。
func minimaxRemainsJSONTotal(intervalRemaining, weeklyRemaining *float64, weeklyTotal, intervalTotal float64) string {
	m := fmt.Sprintf(`{
		"base_resp": {"status_code": 0, "status_msg": "success"},
		"model_remains": [{
			"model_name": "MiniMax-M3",
			"current_interval_total_count": %s,
			"current_interval_remaining_percent": %s,
			"remains_time": 3600000,
			"current_weekly_total_count": %s,
			"current_weekly_remaining_percent": %s,
			"weekly_remains_time": 86400000
		}]
	}`, fnumJSON(intervalTotal), fptrJSON(intervalRemaining), fnumJSON(weeklyTotal), fptrJSON(weeklyRemaining))
	return m
}

func fptrJSON(v *float64) string {
	if v == nil {
		return "null"
	}
	return fmt.Sprintf("%v", *v)
}

func fnumJSON(v float64) string {
	if v == 0 {
		return "0"
	}
	return fmt.Sprintf("%v", v)
}

func fptr(v float64) *float64 { return &v }

// TestMiniMax_EmptyKey_ReturnsError 验证空 Key 返回错误。
func TestMiniMax_EmptyKey_ReturnsError(t *testing.T) {
	f := NewMiniMaxFetcher("")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty api key")
	}
	if result.Platform != "MiniMax" {
		t.Errorf("expected platform 'MiniMax', got '%s'", result.Platform)
	}
}

// TestMiniMax_Success_FormatsResult 验证正常解析:剩余% 反转为已用%,Percent 取
// 5h 与周窗口更紧张者,Used/Total 用 5h 窗口绝对值(与 Kimi 展示风格对齐)。
func TestMiniMax_Success_FormatsResult(t *testing.T) {
	var gotAuth, gotReferer string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotReferer = r.Header.Get("Referer")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(minimaxRemainsJSON(fptr(90), fptr(70), 35000000)))
	}))
	defer server.Close()

	f := NewMiniMaxFetcher("sk-cp_test")
	f.baseURLs = []string{server.URL}
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotAuth != "Bearer sk-cp_test" {
		t.Errorf("expected Bearer sk-cp_test, got '%s'", gotAuth)
	}
	if gotReferer == "" {
		t.Error("expected Referer header (官方 web 端同款)")
	}
	// 5h 剩 90% → 已用 10%;周剩 70% → 已用 30%;Percent 取更紧张的 30
	if result.Percent != 30 {
		t.Errorf("expected Percent=30, got %f", result.Percent)
	}
	if result.Used != 500000 || result.Total != 5000000 {
		t.Errorf("expected Used=500000/Total=5000000, got %f/%f", result.Used, result.Total)
	}
	if result.Remaining != "500,000 / 5,000,000 (5小时) · 周 30.0% 已用" {
		t.Errorf("unexpected Remaining: '%s'", result.Remaining)
	}
	// ResetAt 来自 5h 窗口的 remains_time(3600000ms)
	if reset, err := time.Parse(time.RFC3339, result.ResetAt); err != nil {
		t.Errorf("expected parseable ResetAt, got '%s'", result.ResetAt)
	} else if diff := time.Until(reset); diff < 55*time.Minute || diff > 65*time.Minute {
		t.Errorf("expected ResetAt ~1h from now, got '%s'", result.ResetAt)
	}
}

// TestMiniMax_RemainingPercent_SemanticsGuard 回归守护:字段是"剩余%",不能当"已用%"
// 直接展示(参照 minimax-status 1.2.5 的语义反转事故)。
func TestMiniMax_RemainingPercent_SemanticsGuard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(minimaxRemainsJSON(fptr(100), fptr(100), 35000000)))
	}))
	defer server.Close()

	f := NewMiniMaxFetcher("sk-cp_test")
	f.baseURLs = []string{server.URL}
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 0 {
		t.Errorf("剩余 100%% 应得 Percent=0(未用), got %f", result.Percent)
	}
	if !strings.HasPrefix(result.Remaining, "0 / 5,000,000") {
		t.Errorf("剩余 100%% 应得已用 0, got '%s'", result.Remaining)
	}
}

// TestMiniMax_WeeklyUnlimited_SkipsWeekly 验证"无周限"判定:周总额=0 且未返回
// 周剩余百分比时,周窗口不参与 Percent,展示"周无限制"。
func TestMiniMax_WeeklyUnlimited_SkipsWeekly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(minimaxRemainsJSON(fptr(90), nil, 0)))
	}))
	defer server.Close()

	f := NewMiniMaxFetcher("sk-cp_test")
	f.baseURLs = []string{server.URL}
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 10 {
		t.Errorf("expected Percent=10 (仅 5h), got %f", result.Percent)
	}
	if result.Remaining != "500,000 / 5,000,000 (5小时) · 周无限制" {
		t.Errorf("unexpected Remaining: '%s'", result.Remaining)
	}
}

// TestMiniMax_TotalZero_FallsBackToPercentStyle 验证总量为 0(Plus 年付未 provision
// 的已知问题)时退化为纯百分比展示,不产生 NaN。
func TestMiniMax_TotalZero_FallsBackToPercentStyle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(minimaxRemainsJSONTotal(fptr(90), fptr(70), 0, 0)))
	}))
	defer server.Close()

	f := NewMiniMaxFetcher("sk-cp_test")
	f.baseURLs = []string{server.URL}
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 30 {
		t.Errorf("expected Percent=30, got %f", result.Percent)
	}
	// Used/Total 以 5h 主窗口为准(Percent 才取最紧张窗口,与 Factory Droid 契约一致)
	if result.Total != 100 || result.Used != 10 {
		t.Errorf("expected Used=10/Total=100, got %f/%f", result.Used, result.Total)
	}
	if result.Remaining != "5小时 10.0% 已用 · 周 30.0% 已用" {
		t.Errorf("unexpected Remaining: '%s'", result.Remaining)
	}
}

// TestMiniMax_AuthErrorCode_HintsKeyType 验证业务鉴权错误(1004)给出 Key 类型/区域提示。
func TestMiniMax_AuthErrorCode_HintsKeyType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"base_resp":{"status_code":1004,"status_msg":"login fail"}}`))
	}))
	defer server.Close()

	// 首个主机鉴权失败,第二个主机网络不通,全部失败后报首个错误
	f := NewMiniMaxFetcher("sk-cp_test")
	f.baseURLs = []string{server.URL, "http://127.0.0.1:1"}
	result := f.Fetch()
	if result.Error == "" {
		t.Fatal("expected error for auth failure")
	}
	if !strings.Contains(result.Error, "1004") {
		t.Errorf("expected status code in error, got '%s'", result.Error)
	}
}

// TestMiniMax_FirstHost401_TriesNextHost 验证域名链回退:第一个主机 401 时自动尝试
// 下一个主机,成功则正常解析。
func TestMiniMax_FirstHost401_TriesNextHost(t *testing.T) {
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer first.Close()

	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(minimaxRemainsJSON(fptr(80), fptr(60), 35000000)))
	}))
	defer second.Close()

	f := NewMiniMaxFetcher("sk-cp_test")
	f.baseURLs = []string{first.URL, second.URL}
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	// 5h 剩 80% → 已用 20%;周剩 60% → 已用 40%
	if result.Percent != 40 {
		t.Errorf("expected Percent=40, got %f", result.Percent)
	}
}

// TestMiniMax_AllHostsFail_ReportsError 验证域名链全部失败时返回错误。
func TestMiniMax_AllHostsFail_ReportsError(t *testing.T) {
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer first.Close()
	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer second.Close()

	f := NewMiniMaxFetcher("sk-cp_test")
	f.baseURLs = []string{first.URL, second.URL}
	result := f.Fetch()
	if result.Error == "" {
		t.Fatal("expected error when all hosts fail")
	}
}

// TestMiniMax_EmptyModelRemains_ReturnsError 验证空 model_remains 报结构变化错误。
func TestMiniMax_EmptyModelRemains_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"base_resp":{"status_code":0},"model_remains":[]}`))
	}))
	defer server.Close()

	f := NewMiniMaxFetcher("sk-cp_test")
	f.baseURLs = []string{server.URL}
	result := f.Fetch()
	if result.Error == "" {
		t.Fatal("expected error for empty model_remains")
	}
}
