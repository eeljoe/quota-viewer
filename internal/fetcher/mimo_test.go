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
	if !strings.Contains(result.Error, "Cookie 已过期") {
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
	if !strings.Contains(result.Error, "Cookie 已过期") {
		t.Errorf("expected 'Cookie 已过期', got '%s'", result.Error)
	}
}

func TestMiMoFetcher_JSONResponse_CreditsFields_ParsesUsage(t *testing.T) {
	// 响应字段名尚未确认,使用常见命名 usedCredits/totalCredits
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 Accept header 是 application/json
		accept := r.Header.Get("Accept")
		if accept != "application/json" {
			t.Errorf("expected Accept 'application/json', got '%s'", accept)
		}
		// 验证 Referer
		ref := r.Header.Get("Referer")
		if ref != "https://platform.xiaomimimo.com/console/plan-manage" {
			t.Errorf("expected Referer 'https://platform.xiaomimimo.com/console/plan-manage', got '%s'", ref)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"usedCredits": 8239030362,
			"totalCredits": 38000000000,
			"resetTime": "2026-07-21T00:00:00Z"
		}`))
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
	// 8239030362/38000000000 ≈ 21.68%
	if result.Percent < 21.6 || result.Percent > 21.7 {
		t.Errorf("expected ~21.68%%, got %f%%", result.Percent)
	}
	if !strings.Contains(result.Remaining, "Credits") {
		t.Errorf("expected 'Credits' in Remaining, got '%s'", result.Remaining)
	}
	if result.ResetAt != "2026-07-21T00:00:00Z" {
		t.Errorf("expected ResetAt '2026-07-21T00:00:00Z', got '%s'", result.ResetAt)
	}
}

func TestMiMoFetcher_JSONResponse_StringNumbers_ParsesUsage(t *testing.T) {
	// 部分接口数值以字符串返回,应同样可解析
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"creditsUsed": "8000000000",
			"creditsTotal": "40000000000"
		}`))
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

func TestMiMoFetcher_JSONResponse_UsedTotalFields_ParsesUsage(t *testing.T) {
	// 另一种常见字段命名 used/total
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"used": 300, "total": 1000}`))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Used != 300 {
		t.Errorf("expected Used=300, got %f", result.Used)
	}
	if result.Total != 1000 {
		t.Errorf("expected Total=1000, got %f", result.Total)
	}
	if result.Percent < 29.9 || result.Percent > 30.1 {
		t.Errorf("expected ~30%%, got %f%%", result.Percent)
	}
}

func TestMiMoFetcher_JSONNoUsageData_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "no quota data here"}`))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error when JSON has no usage data")
	}
}
