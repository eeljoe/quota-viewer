package fetcher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestXfyunFetcher_EmptyCookie_ReturnsError(t *testing.T) {
	f := NewXfyunFetcher("", "")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty cookie")
	}
	if result.Platform != "讯飞星辰" {
		t.Errorf("expected platform '讯飞星辰', got '%s'", result.Platform)
	}
}

func TestXfyunFetcher_401_ReturnsCookieExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 Cookie header
		if r.Header.Get("Cookie") != "test=cookie" {
			t.Errorf("expected cookie header, got '%s'", r.Header.Get("Cookie"))
		}
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "Cookie 已过期,请更新" {
		t.Errorf("expected 'Cookie 已过期', got '%s'", result.Error)
	}
}

func TestXfyunFetcher_302_ReturnsCookieExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://maas.xfyun.cn/login")
		w.WriteHeader(302)
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "Cookie 已过期,请更新" {
		t.Errorf("expected 'Cookie 已过期', got '%s'", result.Error)
	}
}

func TestXfyunFetcher_HTMLResponse_ParsesUsage(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<body>
	<div class="package-name">高效版-包季</div>
	<div class="usage">
		<span class="usage-used">1,017</span>
		<span class="usage-separator">/</span>
		<span class="usage-total">6,000</span>
		<span class="usage-unit"> 次</span>
		<div class="progress-bar" style="width: 0.1695"></div>
	</div>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 Accept header 是 text/html 而非 application/json
		accept := r.Header.Get("Accept")
		if accept == "application/json" {
			t.Error("expected Accept to be text/html, not application/json")
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Used != 1017 {
		t.Errorf("expected Used=1017, got %f", result.Used)
	}
	if result.Total != 6000 {
		t.Errorf("expected Total=6000, got %f", result.Total)
	}
	// 1017/6000 = 16.95%
	if result.Percent < 16.9 || result.Percent > 17.0 {
		t.Errorf("expected ~16.95%%, got %f%%", result.Percent)
	}
	// 套餐名应该出现在 Remaining 中
	if !strings.Contains(result.Remaining, "高效版-包季") {
		t.Errorf("expected plan name in Remaining, got '%s'", result.Remaining)
	}
	// 单位应该出现在 Remaining 中
	if !strings.Contains(result.Remaining, "次") {
		t.Errorf("expected unit '次' in Remaining, got '%s'", result.Remaining)
	}
}

func TestXfyunFetcher_HTMLMultipleRows_PicksFirst(t *testing.T) {
	// 页面可能有多行(5小时/周/总),第一组是 5 小时窗口
	html := `<!DOCTYPE html>
<html>
<body>
	<div class="usage-row">
		<span class="usage-used">500</span>
		<span class="usage-total">1,000</span>
		<span class="usage-unit"> 次</span>
	</div>
	<div class="usage-row">
		<span class="usage-used">2,000</span>
		<span class="usage-total">10,000</span>
		<span class="usage-unit"> 次</span>
	</div>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	// 第一组 500/1000 应被解析(usage-used/usage-total 首次匹配)
	if result.Used != 500 {
		t.Errorf("expected Used=500, got %f", result.Used)
	}
	if result.Total != 1000 {
		t.Errorf("expected Total=1000, got %f", result.Total)
	}
}

func TestXfyunFetcher_HTMLNoUsageData_ReturnsError(t *testing.T) {
	html := `<!DOCTYPE html>
<html><body><div>no quota data here</div></body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error when HTML has no usage data")
	}
}
