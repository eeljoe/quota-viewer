package fetcher

import (
	"net/http"
	"net/http/httptest"
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

func TestMiMoFetcher_ValidResponse_ParsesData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"data": {
				"used": 60000000,
				"total": 200000000,
				"resetTime": "2026-08-01T00:00:00Z"
			}
		}`))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Used != 60000000 {
		t.Errorf("expected Used=60000000, got %f", result.Used)
	}
	if result.Total != 200000000 {
		t.Errorf("expected Total=200000000, got %f", result.Total)
	}
	if result.Percent < 29.9 || result.Percent > 30.1 {
		t.Errorf("expected ~30%%, got %f%%", result.Percent)
	}
}
