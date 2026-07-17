package fetcher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiMoFetcher_EmptyCookie_ReturnsError(t *testing.T) {
	f := NewMiMoFetcher("", "")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty cookie")
	}
	if result.Platform != "小米MiMo" {
		t.Errorf("expected platform '小米MiMo', got '%s'", result.Platform)
	}
}

func TestMiMoFetcher_401_ReturnsCookieExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "Cookie 已过期,请更新" {
		t.Errorf("expected 'Cookie 已过期', got '%s'", result.Error)
	}
}

func TestMiMoFetcher_302_ReturnsCookieExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://platform.xiaomimimo.com/login")
		w.WriteHeader(302)
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "Cookie 已过期,请更新" {
		t.Errorf("expected 'Cookie 已过期', got '%s'", result.Error)
	}
}

func TestMiMoFetcher_HTMLResponse_ParsesUsage(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<body>
	<div class="Part1_usageContainer__bc0f900c">
		<span>8,239,030,362 / 38,000,000,000</span>
		<div class="Part1_usagePercent__bc0f900c">已使用 22.0%</div>
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

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Used != 8239030362 {
		t.Errorf("expected Used=8239030362, got %f", result.Used)
	}
	if result.Total != 38000000000 {
		t.Errorf("expected Total=38000000000, got %f", result.Total)
	}
	// 优先采用显式百分比 22.0%
	if result.Percent < 21.9 || result.Percent > 22.1 {
		t.Errorf("expected ~22.0%%, got %f%%", result.Percent)
	}
	if !strings.Contains(result.Remaining, "Credits") {
		t.Errorf("expected 'Credits' in Remaining, got '%s'", result.Remaining)
	}
}

func TestMiMoFetcher_HTMLResponse_DerivesPercentWhenAbsent(t *testing.T) {
	// 没有 "已使用 X%" 时,应从 used/total 计算
	html := `<!DOCTYPE html>
<html><body>
	<div>8,000,000,000 / 40,000,000,000</div>
</body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Used != 8000000000 {
		t.Errorf("expected Used=8000000000, got %f", result.Used)
	}
	if result.Total != 40000000000 {
		t.Errorf("expected Total=40000000000, got %f", result.Total)
	}
	// 8B/40B = 20%
	if result.Percent < 19.9 || result.Percent > 20.1 {
		t.Errorf("expected ~20.0%%, got %f%%", result.Percent)
	}
}

func TestMiMoFetcher_HTMLPercentOnly_Fallback(t *testing.T) {
	// 只有百分比没有 used/total 时,应回退到仅百分比
	html := `<!DOCTYPE html>
<html><body>
	<div class="usage">已使用 35.5%</div>
</body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent < 35.4 || result.Percent > 35.6 {
		t.Errorf("expected ~35.5%%, got %f%%", result.Percent)
	}
	if !strings.Contains(result.Remaining, "35.5") {
		t.Errorf("expected percent in Remaining, got '%s'", result.Remaining)
	}
}

func TestMiMoFetcher_HTMLNoData_ReturnsError(t *testing.T) {
	html := `<!DOCTYPE html>
<html><body><div>no quota data here</div></body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error when HTML has no quota data")
	}
}
