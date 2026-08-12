package fetcher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestOllamaFetcher_EmptyCookie_ReturnsError 验证空 Cookie 返回错误。
func TestOllamaFetcher_EmptyCookie_ReturnsError(t *testing.T) {
	f := NewOllamaFetcher("")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty cookie")
	}
	if result.Platform != "Ollama" {
		t.Errorf("expected platform 'Ollama', got '%s'", result.Platform)
	}
}

// TestOllamaFetcher_302_ReturnsCookieExpired 验证 302 重定向返回 Cookie 过期错误。
func TestOllamaFetcher_302_ReturnsCookieExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://ollama.com/signin")
		w.WriteHeader(302)
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "Cookie") {
		t.Errorf("expected Cookie error, got '%s'", result.Error)
	}
}

// TestOllamaFetcher_401_ReturnsInvalidCredentials 验证 401 返回凭据无效。
func TestOllamaFetcher_401_ReturnsInvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "凭据无效") {
		t.Errorf("expected '凭据无效', got '%s'", result.Error)
	}
}

// TestOllamaFetcher_HTTP500_ReturnsStatusCode 验证 500 返回通用 HTTP 错误。
func TestOllamaFetcher_HTTP500_ReturnsStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "HTTP 500") {
		t.Errorf("expected 'HTTP 500' in error, got '%s'", result.Error)
	}
}

// TestOllamaFetcher_ValidHTML_ParsesBothWindows 验证 HTML 解析:5h + 周窗口。
// 这是用户截图的真实页面结构(Session 10% / Weekly 41.6%)。
func TestOllamaFetcher_ValidHTML_ParsesBothWindows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 Cookie 头
		if !strings.Contains(r.Header.Get("Cookie"), "wos-session=test") {
			t.Errorf("expected Cookie with wos-session, got '%s'", r.Header.Get("Cookie"))
		}
		w.Header().Set("Content-Type", "text/html")
		html := `<!DOCTYPE html>
<html>
<head><title>Settings - Ollama</title></head>
<body>
<h1>Cloud usage</h1>
<span class="badge">Free</span>

<section>
  <h2>Session usage</h2>
  <div class="usage-bar">
    <span>10% used</span>
  </div>
  <span data-time="2026-08-12T12:00:00Z">Resets in 5 hours</span>
</section>

<section>
  <h2>Weekly usage</h2>
  <div class="usage-bar">
    <span>41.6% used</span>
  </div>
  <span data-time="2026-08-16T00:00:00Z">Resets in 4 days</span>
</section>
</body>
</html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	// 主展示 = session 5h 窗口
	if result.Percent != 10 {
		t.Errorf("expected Percent=10 (session), got %f", result.Percent)
	}
	if result.Used != 10 {
		t.Errorf("expected Used=10, got %f", result.Used)
	}
	if result.Total != 100 {
		t.Errorf("expected Total=100, got %f", result.Total)
	}
	// Remaining 应同时含 5 小时和周窗口
	if !strings.Contains(result.Remaining, "5小时 10.0% 已用") {
		t.Errorf("expected 5-hour usage in Remaining, got '%s'", result.Remaining)
	}
	if !strings.Contains(result.Remaining, "周 41.6% 已用") {
		t.Errorf("expected weekly usage in Remaining, got '%s'", result.Remaining)
	}
	// ResetAt = session 重置(更紧迫)
	if result.ResetAt != "2026-08-12T12:00:00Z" {
		t.Errorf("expected ResetAt='2026-08-12T12:00:00Z', got '%s'", result.ResetAt)
	}
}

// TestOllamaFetcher_NoDataTime_LeavesResetUnknown 验证没有 data-time 时仍能展示用量,重置时间留空。
func TestOllamaFetcher_NoDataTime_LeavesResetUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<html><body>
<h1>Cloud usage <span>Free</span></h1>
<h2>Session usage</h2>
<span>5% used</span>
<span>Resets in 3 hours</span>
<h2>Weekly usage</h2>
<span>12% used</span>
<span>Resets in 2 days</span>
</body></html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 5 {
		t.Errorf("expected Percent=5, got %f", result.Percent)
	}
	if result.ResetAt != "" {
		t.Errorf("expected empty ResetAt without data-time, got '%s'", result.ResetAt)
	}
}

// TestOllamaFetcher_OnlySession_NoWeekly 验证只有 session 段(周窗口缺失)也能正确展示。
func TestOllamaFetcher_OnlySession_NoWeekly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<html><body>
<h1>Cloud usage <span>Free</span></h1>
<h2>Session usage</h2>
<span>25% used</span>
<span data-time="2026-08-12T10:00:00Z">Resets in 5 hours</span>
<!-- Weekly 段缺失 -->
</body></html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 25 {
		t.Errorf("expected Percent=25, got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "5小时 25.0% 已用") {
		t.Errorf("expected 5-hour usage in Remaining, got '%s'", result.Remaining)
	}
	// Remaining 不应包含 "周"
	if strings.Contains(result.Remaining, "周") {
		t.Errorf("Remaining should not contain '周' when weekly missing, got '%s'", result.Remaining)
	}
}

// TestOllamaFetcher_PageStructureChanged_ReturnsError 验证页面结构变化(无 Cloud usage 标题)报错。
func TestOllamaFetcher_PageStructureChanged_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>Welcome</h1><p>Some other page</p></body></html>`))
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "页面结构") {
		t.Errorf("expected page structure error, got '%s'", result.Error)
	}
}

// TestOllamaFetcher_LoggedInButNoQuotaData 验证 settings 页面缺少 5 小时窗口时返回明确错误。
func TestOllamaFetcher_LoggedInButNoQuotaData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>Cloud Usage <span>Pro</span></h1><p>No usage yet</p></body></html>`))
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "5 小时窗口") {
		t.Errorf("expected missing 5-hour window error, got '%s'", result.Error)
	}
}

// TestOllamaFetcher_SignedOutHTML_ReturnsCookieExpired 验证登录页 HTML 被识别为 Cookie 失效。
func TestOllamaFetcher_SignedOutHTML_ReturnsCookieExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>Sign in to Ollama</h1><form action="/auth/signin"><input type="email"><input type="password"></form></body></html>`))
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "Cookie 已过期") {
		t.Errorf("expected expired Cookie error, got '%s'", result.Error)
	}
}

// TestOllamaFetcher_ZeroPercent_IsValid 验证刚重置后的 0% 不是解析失败。
func TestOllamaFetcher_ZeroPercent_IsValid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>
<span>Session usage</span><span>0% used</span><div data-time="2026-08-12T12:00:00Z">Resets in 5 hours</div>
<span>Weekly usage</span><span>41.6% used</span><div data-time="2026-08-16T00:00:00Z">Resets in 4 days</div>
</body></html>`))
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 0 {
		t.Errorf("expected Percent=0, got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "周 41.6%") {
		t.Errorf("expected weekly usage in Remaining, got '%s'", result.Remaining)
	}
}

// TestOllamaFetcher_WidthFallback 验证页面只保留进度条 width 时仍可解析 5 小时窗口。
func TestOllamaFetcher_WidthFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>
<span>Session usage</span><div style="width: 8.5%"></div><div data-time="2026-08-12T12:00:00Z">Resets in 5 hours</div>
<span>Weekly usage</span><div style="width: 20%"></div>
</body></html>`))
	}))
	defer server.Close()

	f := NewOllamaFetcher("test-token")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 8.5 {
		t.Errorf("expected Percent=8.5, got %f", result.Percent)
	}
}

// TestOllamaFetcher_HighSessionPercent 验证 5h 窗口接近 100%(用户告警场景)。
func TestOllamaFetcher_HighSessionPercent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<html><body>
<h1>Cloud usage <span>Free</span></h1>
<h2>Session usage</h2>
<span>95.5% USED</span>
<span data-time="2026-08-12T10:00:00Z">Resets in 1 hour</span>
<h2>Weekly usage</h2>
<span>20% used</span>
<span data-time="2026-08-15T00:00:00Z">Resets in 3 days</span>
</body></html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOllamaFetcher("wos-session=test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	// 5h 95.5% 应该是主展示(警示色)
	if result.Percent != 95.5 {
		t.Errorf("expected Percent=95.5, got %f", result.Percent)
	}
}
