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

func TestXfyunFetcher_JSONResponse_ParsesPackageTotal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if accept != "application/json" {
			t.Errorf("expected Accept 'application/json', got '%s'", accept)
		}
		ref := r.Header.Get("Referer")
		if ref != "https://maas.xfyun.cn/packageSubscription" {
			t.Errorf("expected Referer, got '%s'", ref)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"data": {
				"page": 1,
				"rows": [
					{
						"appId": "mc135ca1",
						"codingPlanUsageDTO": {
							"rp5hUsage": 768.0,
							"rp5hLimit": 6000,
							"rpwUsage": 3302.8,
							"rpwLimit": 45000,
							"packageUsage": 3300.8,
							"packageLimit": 270000,
							"packageLeft": 266699.2
						},
						"expiresAt": "2026-10-16 10:07:39",
						"id": 2617260370923522
					}
				]
			}
		}`))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Used != 3300.8 {
		t.Errorf("expected Used=3300.8 (package), got %f", result.Used)
	}
	if result.Total != 270000 {
		t.Errorf("expected Total=270000 (package), got %f", result.Total)
	}
	if result.Percent < 1.1 || result.Percent > 1.3 {
		t.Errorf("expected ~1.2%%, got %f%%", result.Percent)
	}
	if !strings.Contains(result.Remaining, "总量") {
		t.Errorf("expected '总量' in Remaining, got '%s'", result.Remaining)
	}
	if result.ResetAt != "2026-10-16 10:07:39" {
		t.Errorf("expected ResetAt, got '%s'", result.ResetAt)
	}
}

func TestXfyunFetcher_NonZeroCode_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code": 401, "data": null}`))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for non-zero code")
	}
}

func TestXfyunFetcher_EmptyRows_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code": 0, "data": {"page": 1, "rows": []}}`))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error when rows is empty")
	}
}
