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
	if result.Balance != 110.00 {
		t.Errorf("expected Balance=110.00, got %f", result.Balance)
	}
	if result.Currency != "CNY" {
		t.Errorf("expected Currency=CNY, got %s", result.Currency)
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

func TestApplyBudget_NormalCalc(t *testing.T) {
	r := QuotaResult{Kind: KindBalance, Balance: 150, Currency: "CNY", Remaining: "原始"}
	ApplyBudget(&r, 500)
	if r.Used != 350 {
		t.Errorf("expected Used=350, got %f", r.Used)
	}
	if r.Total != 500 {
		t.Errorf("expected Total=500, got %f", r.Total)
	}
	if r.Percent < 69.9 || r.Percent > 70.1 {
		t.Errorf("expected ~70%%, got %f", r.Percent)
	}
	if !strings.Contains(r.Remaining, "150.00") || !strings.Contains(r.Remaining, "500.00") || !strings.Contains(r.Remaining, "预算") {
		t.Errorf("expected budget Remaining, got %s", r.Remaining)
	}
}

func TestApplyBudget_ZeroBudget_UsesDefault(t *testing.T) {
	r := QuotaResult{Kind: KindBalance, Balance: 150, Currency: "CNY", Remaining: "原始", Percent: 0}
	ApplyBudget(&r, 0)
	// budget=0 → 使用默认 300: Used=300-150=150, Percent=50%
	if r.Used != 150 {
		t.Errorf("expected Used=150 (default budget 300), got %f", r.Used)
	}
	if r.Total != 300 {
		t.Errorf("expected Total=300 (default), got %f", r.Total)
	}
	if r.Percent < 49.9 || r.Percent > 50.1 {
		t.Errorf("expected ~50%%, got %f", r.Percent)
	}
	if !strings.Contains(r.Remaining, "预算") {
		t.Errorf("expected budget Remaining, got %s", r.Remaining)
	}
}

func TestApplyBudget_BalanceExceedsBudget_ClampsUsed(t *testing.T) {
	r := QuotaResult{Kind: KindBalance, Balance: 600, Currency: "CNY", Remaining: "原始"}
	ApplyBudget(&r, 500)
	if r.Used != 0 {
		t.Errorf("expected Used=0 when balance>budget, got %f", r.Used)
	}
	if r.Percent != 0 {
		t.Errorf("expected Percent=0, got %f", r.Percent)
	}
}

func TestApplyBudget_UsageKind_NoOp(t *testing.T) {
	r := QuotaResult{Kind: KindUsage, Balance: 100, Currency: "CNY", Percent: 50, Remaining: "原始"}
	ApplyBudget(&r, 500)
	if r.Percent != 50 {
		t.Errorf("expected Percent unchanged for usage kind, got %f", r.Percent)
	}
	if r.Remaining != "原始" {
		t.Errorf("expected Remaining unchanged for usage kind, got %s", r.Remaining)
	}
}

func TestApplyBudget_ErrorResult_NoOp(t *testing.T) {
	r := QuotaResult{Kind: KindBalance, Balance: 100, Currency: "CNY", Error: "出错了"}
	ApplyBudget(&r, 500)
	if r.Percent != 0 {
		t.Errorf("expected Percent=0 for error result, got %f", r.Percent)
	}
	if r.Used != 0 {
		t.Errorf("expected Used=0 for error result, got %f", r.Used)
	}
}
