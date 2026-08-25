package fetcher

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommandCode_EmptyKey_ReturnsError 验证空 API Key(且无 auth.json)返回错误。
func TestCommandCode_EmptyKey_ReturnsError(t *testing.T) {
	f := NewCommandCodeFetcher("")
	f.readAuthFile = false // 禁用 auth.json 读取,保证离线可测
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty api key")
	}
	if result.Platform != "Command Code" {
		t.Errorf("expected platform 'Command Code', got '%s'", result.Platform)
	}
}

// TestCommandCode_EmptyKey_ReadsAuthFile 验证空 key 时自动读取 ~/.commandcode/auth.json。
func TestCommandCode_EmptyKey_ReadsAuthFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("USERPROFILE", tmp)
	if err := writeAuthFile(t, tmp, `{"apiKey":"user_secret"}`); err != nil {
		t.Fatal(err)
	}

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/alpha/whoami":
			_, _ = w.Write([]byte(`{"success":true,"user":{"userName":"eeljoe"},"org":null}`))
		case "/alpha/billing/credits":
			_, _ = w.Write([]byte(validCreditsJSON))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	f := NewCommandCodeFetcher("")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotAuth != "Bearer user_secret" {
		t.Errorf("expected Bearer user_secret, got '%s'", gotAuth)
	}
}

// TestCommandCode_401_ReturnsInvalidKey 验证 401 返回 Key 失效错误。
func TestCommandCode_401_ReturnsInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewCommandCodeFetcher("user_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "API Key 无效") {
		t.Errorf("expected invalid key error, got '%s'", result.Error)
	}
}

// TestCommandCode_HTTP500_ReturnsStatusCode 验证 500 返回通用 HTTP 错误。
func TestCommandCode_HTTP500_ReturnsStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	f := NewCommandCodeFetcher("user_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "HTTP 500") {
		t.Errorf("expected 'HTTP 500' in error, got '%s'", result.Error)
	}
}

// TestCommandCode_ValidResp_FormatsResult 验证正常 JSON 解析:
// 5 小时窗口驱动主展示,周窗口与余额写入 Remaining,ResetAt 转 UTC ISO。
func TestCommandCode_ValidResp_FormatsResult(t *testing.T) {
	var gotCredsQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/alpha/whoami":
			_, _ = w.Write([]byte(`{"success":true,"user":{"userName":"eeljoe"},"org":null}`))
		case "/alpha/billing/credits":
			gotCredsQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(validCreditsJSON))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	f := NewCommandCodeFetcher("user_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent < 21.2 || result.Percent > 21.3 { // 0.6385129295 / 3 * 100
		t.Errorf("expected Percent~21.28, got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "5小时") ||
		!strings.Contains(result.Remaining, "周") ||
		!strings.Contains(result.Remaining, "余额 $9.36") {
		t.Errorf("unexpected Remaining: '%s'", result.Remaining)
	}
	if result.ResetAt != "2026-08-25T07:45:12Z" {
		t.Errorf("expected ResetAt='2026-08-25T07:45:12Z', got '%s'", result.ResetAt)
	}
	// 个人用户 org=null → credits 不带 orgId 参数
	if gotCredsQuery != "" {
		t.Errorf("expected no query for personal user, got '%s'", gotCredsQuery)
	}
}

// TestCommandCode_OrgUser_SendsOrgID 验证团队用户(whoami 返回 org)会把 orgId 传给 credits。
func TestCommandCode_OrgUser_SendsOrgID(t *testing.T) {
	var gotCredsQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/alpha/whoami":
			_, _ = w.Write([]byte(`{"success":true,"user":{"userName":"eeljoe"},"org":{"id":"org_123"}}`))
		case "/alpha/billing/credits":
			gotCredsQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(validCreditsJSON))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	f := NewCommandCodeFetcher("user_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotCredsQuery != "orgId=org_123" {
		t.Errorf("expected orgId=org_123, got '%s'", gotCredsQuery)
	}
}

// TestCommandCode_MissingWindows_ReturnsError 验证 credits 响应缺窗口数据时明确报错。
func TestCommandCode_MissingWindows_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/alpha/whoami":
			_, _ = w.Write([]byte(`{"success":true,"org":null}`))
		case "/alpha/billing/credits":
			_, _ = w.Write([]byte(`{"credits":{"monthlyCredits":9.36}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	f := NewCommandCodeFetcher("user_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "响应结构") {
		t.Errorf("expected structure error, got '%s'", result.Error)
	}
}

// validCreditsJSON 是真实 /alpha/billing/credits 返回结构的代表(含 5h/周窗口)。
const validCreditsJSON = `{
  "credits": {
    "belowThreshold": false,
    "creditThreshold": 0,
    "monthlyCredits": 9.3614870705,
    "purchasedCredits": 0,
    "freeCredits": 0
  },
  "windowLimits": {
    "limited": true,
    "exceeded": null,
    "fiveHour": {"used": 0.6385129295, "cap": 3, "exceeded": false, "resetAt": 1787643912049},
    "weekly": {"used": 0.6385129295, "cap": 6, "exceeded": false, "resetAt": 1788230712049}
  }
}`

// writeAuthFile 写入 ~/.commandcode/auth.json(模拟官方 CLI 登录后的状态)。
func writeAuthFile(t *testing.T, dir, content string) error {
	t.Helper()
	authDir := filepath.Join(dir, ".commandcode")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(authDir, "auth.json"), []byte(content), 0644)
}
