package fetcher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeepSeekFetcher_EmptyKey_ReturnsError(t *testing.T) {
	f := NewDeepSeekFetcher("")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty api key")
	}
	if result.Kind != KindBalance {
		t.Errorf("expected Kind=balance, got '%s'", result.Kind)
	}
}

func TestDeepSeekFetcher_BalanceOK_ParsesRemaining(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer sk-test" {
			t.Errorf("expected Authorization 'Bearer sk-test', got '%s'", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"is_available": true,
			"balance_infos": [
				{"currency": "CNY", "total_balance": "110.00", "granted_balance": "10.00", "topped_up_balance": "100.00"}
			]
		}`))
	}))
	defer server.Close()

	f := NewDeepSeekFetcher("sk-test")
	f.apiURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if !strings.Contains(result.Remaining, "110.00") {
		t.Errorf("expected '110.00' in Remaining, got '%s'", result.Remaining)
	}
	if !strings.Contains(result.Remaining, "CNY") {
		t.Errorf("expected 'CNY' in Remaining, got '%s'", result.Remaining)
	}
	if result.Percent != 0 {
		t.Errorf("expected Percent=0 for balance kind, got %f", result.Percent)
	}
}

func TestDeepSeekFetcher_401_ReturnsInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewDeepSeekFetcher("sk-test")
	f.apiURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "无效") {
		t.Errorf("expected 'API Key 无效', got '%s'", result.Error)
	}
}

func TestDeepSeekFetcher_MultiCurrency_PicksNonZeroBalance(t *testing.T) {
	// 用户真实场景:USD 0.00 + CNY 247.51 → 应选中 CNY
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"is_available": true,
			"balance_infos": [
				{"currency": "USD", "total_balance": "0.00", "granted_balance": "0.00", "topped_up_balance": "0.00"},
				{"currency": "CNY", "total_balance": "247.51", "granted_balance": "0.00", "topped_up_balance": "247.51"}
			]
		}`))
	}))
	defer server.Close()

	f := NewDeepSeekFetcher("sk-test")
	f.apiURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if !strings.Contains(result.Remaining, "247.51") || !strings.Contains(result.Remaining, "CNY") {
		t.Errorf("expected CNY 247.51 in Remaining, got '%s'", result.Remaining)
	}
	if !strings.Contains(result.Remaining, "¥") {
		t.Errorf("expected currency symbol ¥ in Remaining, got '%s'", result.Remaining)
	}
}

func TestDeepSeekFetcher_NotAvailable_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"is_available": false, "balance_infos": []}`))
	}))
	defer server.Close()

	f := NewDeepSeekFetcher("sk-test")
	f.apiURL = server.URL
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error when is_available=false")
	}
}

func TestDeepSeekFetcher_ZeroBalance_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"is_available": true,
			"balance_infos": [{"currency": "CNY", "total_balance": "0.00", "granted_balance": "0.00", "topped_up_balance": "0.00"}]
		}`))
	}))
	defer server.Close()

	f := NewDeepSeekFetcher("sk-test")
	f.apiURL = server.URL
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error when total_balance is 0")
	}
}
