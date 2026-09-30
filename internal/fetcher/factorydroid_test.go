package fetcher

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFactoryDroid_EmptyKey_ReturnsError 验证空 API Key(且无 .env)返回错误。
func TestFactoryDroid_EmptyKey_ReturnsError(t *testing.T) {
	f := NewFactoryDroidFetcher("")
	f.readEnvFile = false // 禁用 .env 读取,保证离线可测
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty api key")
	}
	if result.Platform != "Factory Droid" {
		t.Errorf("expected platform 'Factory Droid', got '%s'", result.Platform)
	}
}

// TestFactoryDroid_EmptyKey_ReadsEnvFile 验证空 key 时自动读取 ~/.factory/.env 的
// FACTORY_API_KEY(支持 export 前缀与行尾注释,Droid CLI 同款格式)。
func TestFactoryDroid_EmptyKey_ReadsEnvFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("USERPROFILE", tmp)
	envDir := filepath.Join(tmp, ".factory")
	if err := os.MkdirAll(envDir, 0755); err != nil {
		t.Fatal(err)
	}
	envContent := "export FACTORY_API_KEY='fk_secret' # Droid key\n"
	if err := os.WriteFile(filepath.Join(envDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(validFactoryLimitsJSON))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotAuth != "Bearer fk_secret" {
		t.Errorf("expected Bearer fk_secret, got '%s'", gotAuth)
	}
}

// TestFactoryDroid_TokenRateLimits_FormatsResult 验证新账单模型(token-rate limits)解析:
// 5h 窗口驱动主展示,Percent 取全部窗口最紧张的,预付余额写入 Remaining。
func TestFactoryDroid_TokenRateLimits_FormatsResult(t *testing.T) {
	var gotClientHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClientHeader = r.Header.Get("x-factory-client")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(validFactoryLimitsJSON))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotClientHeader != "web-app" {
		t.Errorf("expected x-factory-client 'web-app', got '%s'", gotClientHeader)
	}
	// 5h 20% / 周 45% / 月 12% → Percent = 45(最紧张窗口)
	if result.Percent != 45 {
		t.Errorf("expected Percent=45, got %f", result.Percent)
	}
	if result.Total != 100 || result.Used != 20 {
		t.Errorf("expected Used=20/Total=100 (5h 窗口驱动主展示), got %f/%f", result.Used, result.Total)
	}
	if !strings.Contains(result.Remaining, "5小时 20.0% 已用") {
		t.Errorf("expected 5-hour window leading Remaining (Ollama convention), got '%s'", result.Remaining)
	}
	if !strings.Contains(result.Remaining, "周 45.0% 已用") || !strings.Contains(result.Remaining, "Extra $3.20") {
		t.Errorf("unexpected Remaining: '%s'", result.Remaining)
	}
	// Core 池全 0 且无预付余额时不产生展示噪音(真实账户常见形态)
	if strings.Contains(result.Remaining, "Core") || strings.Contains(result.Remaining, "Extra $0.00") {
		t.Errorf("zero-value pools should stay silent, got '%s'", result.Remaining)
	}
	// ResetAt 来自 5h 窗口的 secondsRemaining(3600s)
	if reset, err := time.Parse(time.RFC3339, result.ResetAt); err != nil {
		t.Errorf("expected parseable ResetAt, got '%s'", result.ResetAt)
	} else if diff := time.Until(reset); diff < 55*time.Minute || diff > 65*time.Minute {
		t.Errorf("expected ResetAt ~1h from now, got '%s'", result.ResetAt)
	}
}

// TestFactoryDroid_WeeklyExhausted_DrivesPercent 验证周/月窗口耗尽时球色必须告警:
// Percent 取全部窗口中更紧张的一个(5h 0% / 周 100% 不应恒绿)。
func TestFactoryDroid_WeeklyExhausted_DrivesPercent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "usesTokenRateLimitsBilling": true,
  "limits": {"standard": {"fiveHour": {"usedPercent": 0}, "weekly": {"usedPercent": 100}}}
}`))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 100 {
		t.Errorf("expected Percent=100 (weekly exhausted drives ball color), got %f", result.Percent)
	}
}

// TestFactoryDroid_CoreWindow_Alerts 验证 Core 池(独立计费窗口)耗尽同样驱动告警。
func TestFactoryDroid_CoreWindow_Alerts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "usesTokenRateLimitsBilling": true,
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 10}},
    "core": {"monthly": {"usedPercent": 100}}
  }
}`))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 100 {
		t.Errorf("expected Percent=100 (core exhausted drives ball color), got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "Core") {
		t.Errorf("expected Core info in Remaining, got '%s'", result.Remaining)
	}
}

// TestFactoryDroid_Legacy_Fallback 验证旧账单模型兜底:billing/limits 未启用
// token-rate 计费时,回退 /api/organization/subscription/usage。
func TestFactoryDroid_Legacy_Fallback(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/billing/limits":
			_, _ = w.Write([]byte(`{"usesTokenRateLimitsBilling": false}`))
		case "/api/organization/subscription/usage":
			gotQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{
  "usage": {
    "endDate": 1789344000000,
    "standard": {"userTokens": 250, "totalAllowance": 1000, "usedRatio": 0},
    "premium": {"userTokens": 50, "totalAllowance": 100, "usedRatio": 0.5}
  }
}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotQuery == "" || !strings.Contains(gotQuery, "useCache=true") {
		t.Errorf("expected useCache=true on legacy route, got '%s'", gotQuery)
	}
	// Standard: ratio=0 但 userTokens=250/1000(脏 ratio 不可信)→ 25%;Premium 50% → Percent=50
	if result.Percent != 50 {
		t.Errorf("expected Percent=50, got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "Standard 25.0%") || !strings.Contains(result.Remaining, "Premium 50.0%") {
		t.Errorf("unexpected Remaining: '%s'", result.Remaining)
	}
}

// TestFactoryDroid_401_ReturnsInvalidKey 验证 401 返回 Key 失效错误。
func TestFactoryDroid_401_ReturnsInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "无效") {
		t.Errorf("expected invalid key error, got '%s'", result.Error)
	}
}

// TestFactoryDroid_429_ReturnsRateLimited 验证 429 透出限流错误。
func TestFactoryDroid_429_ReturnsRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "429") {
		t.Errorf("expected rate limit error, got '%s'", result.Error)
	}
}

// TestFactoryDroid_BadJSON_ReturnsError 验证两路响应都异常时明确报错。
func TestFactoryDroid_BadJSON_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "无法解析") {
		t.Errorf("expected parse error, got '%s'", result.Error)
	}
}

// validFactoryLimitsJSON 是真实 /api/billing/limits 返回结构的代表
// (token-rate limits 计费:5h/周/月窗口 + 预付 Extra Usage 余额)。
const validFactoryLimitsJSON = `{
  "usesTokenRateLimitsBilling": true,
  "limits": {
    "standard": {
      "fiveHour": {"usedPercent": 20, "secondsRemaining": 3600},
      "weekly": {"usedPercent": 45},
      "monthly": {"usedPercent": 12}
    }
  },
  "extraUsageBalanceCents": 320
}`
