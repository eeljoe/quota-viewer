package fetcher

import (
	"net/http"
	"net/http/httptest"
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

func TestXfyunFetcher_ValidResponse_ParsesData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"data": {
				"used": 5000,
				"total": 18000,
				"resetTime": "2026-07-21T08:00:00Z"
			}
		}`))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Used != 5000 {
		t.Errorf("expected Used=5000, got %f", result.Used)
	}
	if result.Total != 18000 {
		t.Errorf("expected Total=18000, got %f", result.Total)
	}
	if result.Percent < 27.7 || result.Percent > 27.8 {
		t.Errorf("expected ~27.78%%, got %f%%", result.Percent)
	}
	if result.ResetAt != "2026-07-21T08:00:00Z" {
		t.Errorf("expected reset time, got '%s'", result.ResetAt)
	}
}

func TestXfyunFetcher_TotalAndRemainingWithoutUsed_DerivesUsed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"data": {
				"total": 18000,
				"remaining": 13000,
				"resetTime": "2026-07-21T08:00:00Z"
			}
		}`))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Used != 5000 {
		t.Errorf("expected Used=5000 (total-remaining), got %f", result.Used)
	}
	if result.Total != 18000 {
		t.Errorf("expected Total=18000, got %f", result.Total)
	}
	if result.Percent < 27.7 || result.Percent > 27.8 {
		t.Errorf("expected ~27.78%%, got %f%%", result.Percent)
	}
	if result.ResetAt != "2026-07-21T08:00:00Z" {
		t.Errorf("expected reset time, got '%s'", result.ResetAt)
	}
}

func TestXfyunFetcher_ArrayData_ParsesFirst(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"data": [
				{"used": 100, "total": 200, "resetTime": "2026-07-21T00:00:00Z"}
			]
		}`))
	}))
	defer server.Close()

	f := NewXfyunFetcher("test=cookie", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Used != 100 || result.Total != 200 {
		t.Errorf("expected 100/200, got %f/%f", result.Used, result.Total)
	}
}
