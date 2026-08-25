package fetcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestOpenCodeGoFetcher_EmptySession_ReturnsError 验证空 sessionToken 返回错误。
func TestOpenCodeGoFetcher_EmptySession_ReturnsError(t *testing.T) {
	f := NewOpenCodeGoFetcher("ws-123", "")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty session token")
	}
	if result.Platform != "OpenCode Go" {
		t.Errorf("expected platform 'OpenCode Go', got '%s'", result.Platform)
	}
}

// TestOpenCodeGoFetcher_302_ReturnsSessionExpired 验证 302 跳转返回"会话已过期"。
func TestOpenCodeGoFetcher_302_ReturnsSessionExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://opencode.ai/login")
		w.WriteHeader(302)
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-123", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "会话已过期") {
		t.Errorf("expected '会话已过期', got '%s'", result.Error)
	}
}

// TestOpenCodeGoFetcher_303_ReturnsSessionExpired 验证 303 跳转也返回"会话已过期"。
func TestOpenCodeGoFetcher_303_ReturnsSessionExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://opencode.ai/login")
		w.WriteHeader(303)
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-123", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "会话已过期") {
		t.Errorf("expected '会话已过期', got '%s'", result.Error)
	}
}

// TestOpenCodeGoFetcher_401_ReturnsInvalidCredentials 验证 401 返回"凭据无效"。
func TestOpenCodeGoFetcher_401_ReturnsInvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-123", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "凭据无效") {
		t.Errorf("expected '凭据无效', got '%s'", result.Error)
	}
}

// TestOpenCodeGoFetcher_403_ReturnsInvalidCredentials 验证 403 返回"凭据无效"。
func TestOpenCodeGoFetcher_403_ReturnsInvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-123", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "凭据无效") {
		t.Errorf("expected '凭据无效', got '%s'", result.Error)
	}
}

// TestOpenCodeGoFetcher_404_ReturnsWorkspaceNotFound 验证 404 返回"Workspace 不存在"。
func TestOpenCodeGoFetcher_404_ReturnsWorkspaceNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-999", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "Workspace 不存在") {
		t.Errorf("expected 'Workspace 不存在', got '%s'", result.Error)
	}
}

// TestOpenCodeGoFetcher_SSRData_PicksHighestPercent 验证 SSR 数据解析正确,选择最高 usagePercent 的窗口。
func TestOpenCodeGoFetcher_SSRData_PicksHighestPercent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 Cookie 和 User-Agent
		cookie := r.Header.Get("Cookie")
		if !strings.Contains(cookie, "auth=tok_abc") {
			t.Errorf("expected Cookie containing 'auth=tok_abc', got '%s'", cookie)
		}
		ua := r.Header.Get("User-Agent")
		if !strings.Contains(ua, "Mozilla") {
			t.Errorf("expected User-Agent containing 'Mozilla', got '%s'", ua)
		}
		// 验证路径包含 workspaceID
		if !strings.Contains(r.URL.Path, "ws-456") {
			t.Errorf("expected path containing 'ws-456', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/html")
		html := `<!DOCTYPE html>
<html>
<head><title>Usage</title></head>
<body>
<script>
  var data = {
    rollingUsage:$R[10]={usagePercent:7,resetInSec:18000},
    weeklyUsage:$R[11]={usagePercent:2,resetInSec:540000},
    monthlyUsage:$R[12]={usagePercent:16,resetInSec:2480000}
  };
</script>
</body>
</html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-456", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 16 {
		t.Errorf("expected Percent=16 (monthly highest), got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "月窗口") {
		t.Errorf("expected '月窗口' in Remaining, got '%s'", result.Remaining)
	}
	if result.Total != 100 {
		t.Errorf("expected Total=100, got %f", result.Total)
	}
	// ResetAt 应为将来时间
	_, err := time.Parse(time.RFC3339, result.ResetAt)
	if err != nil {
		t.Errorf("expected valid ISO time in ResetAt, got '%s': %v", result.ResetAt, err)
	}
}

// TestOpenCodeGoFetcher_SSRData_TiebreakerRollingWins 验证相同百分比时 rolling > weekly > monthly。
func TestOpenCodeGoFetcher_SSRData_TiebreakerRollingWins(t *testing.T) {
	// rolling=10%, weekly=10%, monthly=5% → rolling wins
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<script>
  rollingUsage:$R[1]={usagePercent:10,resetInSec:18000}
  weeklyUsage:$R[2]={usagePercent:10,resetInSec:540000}
  monthlyUsage:$R[3]={usagePercent:5,resetInSec:2480000}
</script>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-1", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 10 {
		t.Errorf("expected Percent=10, got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "5小时窗口") {
		t.Errorf("expected '5小时窗口' in Remaining (rolling wins), got '%s'", result.Remaining)
	}
}

// TestOpenCodeGoFetcher_SSRData_ReversedFieldOrder 验证 usagePercent 和 resetInSec 顺序调换也能解析。
func TestOpenCodeGoFetcher_SSRData_ReversedFieldOrder(t *testing.T) {
	// resetInSec 在前, usagePercent 在后
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<script>
  rollingUsage:$R[1]={resetInSec:18000,usagePercent:25}
</script>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-1", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 25 {
		t.Errorf("expected Percent=25, got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "5小时窗口") {
		t.Errorf("expected '5小时窗口' in Remaining, got '%s'", result.Remaining)
	}
}

// TestOpenCodeGoFetcher_SSRData_DecimalPercent 验证 usagePercent 浮点值(OpenCode 页面实际格式,如 3.1)能正确解析。
func TestOpenCodeGoFetcher_SSRData_DecimalPercent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<script>
  rollingUsage:$R[34]={status:"ok",resetInSec:11630,usagePercent:3.1,usage:36862966,limit:12000000},
  weeklyUsage:$R[35]={status:"ok",resetInSec:467844,usagePercent:43,usage:1289220543,limit:3000000000},
  monthlyUsage:$R[36]={status:"ok",resetInSec:1275137,usagePercent:38.3,usage:2300879983,limit:6000000000}
</script>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-1", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	// 最高 = weekly 43(> rolling 3.1 > monthly 38.3)
	if result.Percent != 43 {
		t.Errorf("expected Percent=43 (weekly highest), got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "周窗口") {
		t.Errorf("expected '周窗口' in Remaining, got '%s'", result.Remaining)
	}
}

// TestOpenCodeGoFetcher_SSRData_DecimalOnlyHighest 验证浮点窗口可被选中为最高值(rolling 3.1 高于 weekly 2)。
func TestOpenCodeGoFetcher_SSRData_DecimalOnlyHighest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<script>
  rollingUsage:$R[1]={usagePercent:3.1,resetInSec:18000}
  weeklyUsage:$R[2]={usagePercent:2,resetInSec:540000}
</script>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-1", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 3.1 {
		t.Errorf("expected Percent=3.1 (rolling highest), got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "5小时窗口") {
		t.Errorf("expected '5小时窗口' in Remaining, got '%s'", result.Remaining)
	}
}

// TestOpenCodeGoFetcher_DataSlot_Fallback 验证 SSR 匹配失败时回退到 data-slot HTML 解析。
func TestOpenCodeGoFetcher_DataSlot_Fallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<!DOCTYPE html>
<html>
<body>
<div data-slot="usage-item"><span data-slot="usage-label">Rolling Usage</span><span data-slot="usage-value"><!--$-->7<!--/-->%</span></div>
<div data-slot="usage-item"><span data-slot="usage-label">Weekly Usage</span><span data-slot="usage-value"><!--$-->2<!--/-->%</span></div>
<div data-slot="usage-item"><span data-slot="usage-label">Monthly Usage</span><span data-slot="usage-value"><!--$-->16<!--/-->%</span></div>
</body>
</html>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-789", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 16 {
		t.Errorf("expected Percent=16 (monthly highest), got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "月窗口") {
		t.Errorf("expected '月窗口' in Remaining, got '%s', error=%s", result.Remaining, result.Error)
	}
}

// TestOpenCodeGoFetcher_NoData_ReturnsPageStructureChanged 验证无数据返回"页面结构已变化"。
func TestOpenCodeGoFetcher_NoData_ReturnsPageStructureChanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>Welcome to OpenCode</body></html>`))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-000", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "页面结构已变化") {
		t.Errorf("expected '页面结构已变化', got '%s'", result.Error)
	}
}

// TestOpenCodeGoFetcher_SSRData_AllZeroPercent_ReturnsNoValidWindows 验证所有窗口 usagePercent=0 时返回"未找到有效配额窗口"。
func TestOpenCodeGoFetcher_SSRData_AllZeroPercent_ReturnsNoValidWindows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<script>
  rollingUsage:$R[10]={usagePercent:0,resetInSec:18000}
  weeklyUsage:$R[11]={usagePercent:0,resetInSec:540000}
  monthlyUsage:$R[12]={usagePercent:0,resetInSec:2480000}
</script>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-1", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "未找到有效配额窗口") {
		t.Errorf("expected '未找到有效配额窗口', got '%s'", result.Error)
	}
}

// TestOpenCodeGoFetcher_SSRData_ZeroPercent_NotPicked 验证 usagePercent=0 的窗口不被纳入选择(避免无意义窗口)。
func TestOpenCodeGoFetcher_SSRData_ZeroPercent_NotPicked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<script>
  rollingUsage:$R[10]={usagePercent:0,resetInSec:18000}
  weeklyUsage:$R[11]={usagePercent:5,resetInSec:540000}
</script>`
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-1", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 5 {
		t.Errorf("expected Percent=5 (weekly, since rolling=0 is skipped), got %f", result.Percent)
	}
}

// TestOpenCodeGoFetcher_HTTP500_ReturnsStatusCode 验证 500 返回通用错误。
func TestOpenCodeGoFetcher_HTTP500_ReturnsStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	f := NewOpenCodeGoFetcher("ws-123", "tok_abc")
	f.baseURL = server.URL
	result := f.Fetch()
	expected := fmt.Sprintf("HTTP %d", 500)
	if result.Error != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Error)
	}
}
