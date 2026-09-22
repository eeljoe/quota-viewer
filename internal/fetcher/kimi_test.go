package fetcher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKimiFetcher_EmptyKey_ReturnsError(t *testing.T) {
	f := NewKimiFetcher("")
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty API key")
	}
	if result.Platform != "Kimi" {
		t.Errorf("expected platform 'Kimi', got '%s'", result.Platform)
	}
}

func TestKimiFetcher_NestedResponse_Parses5hWindow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer sk-kimi-test" {
			t.Errorf("expected 'Bearer sk-kimi-test', got '%s'", auth)
		}
		ua := r.Header.Get("User-Agent")
		if ua != "KimiCLI/1.6" {
			t.Errorf("expected User-Agent 'KimiCLI/1.6', got '%s'", ua)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"user": {"userId": "cr98n7kudu67tmbt5gq0", "region": "REGION_CN", "membership": {"level": "LEVEL_INTERMEDIATE"}},
			"usage": {"limit": "100", "used": "75", "remaining": "25", "resetTime": "2026-07-18T13:13:12.634389Z"},
			"limits": [{"window": {"duration": 300, "timeUnit": "TIME_UNIT_MINUTE"}, "detail": {"limit": "100", "remaining": "25", "resetTime": "2026-07-17T16:13:12.634389Z"}}],
			"totalQuota": {"limit": "100", "remaining": "99"},
			"parallel": {"limit": "20"},
			"authentication": {"method": "METHOD_API_KEY"}
		}`))
	}))
	defer server.Close()

	f := &KimiFetcher{apiKey: "sk-kimi-test", apiURL: server.URL}
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	// 5h window: limit=100, remaining=25 → used=75
	if result.Used != 75 {
		t.Errorf("expected Used=75 (5h window), got %f", result.Used)
	}
	if result.Total != 100 {
		t.Errorf("expected Total=100 (5h window), got %f", result.Total)
	}
	if result.Percent < 74.9 || result.Percent > 75.1 {
		t.Errorf("expected ~75%%, got %f%%", result.Percent)
	}
	if !strings.Contains(result.Remaining, "5小时") {
		t.Errorf("expected '5小时' in Remaining, got '%s'", result.Remaining)
	}
	if result.ResetAt != "2026-07-17T16:13:12.634389Z" {
		t.Errorf("expected 5h reset time, got '%s'", result.ResetAt)
	}
}

// 真实线上响应(2026-09-22):5 小时窗口 0% 已用,但周额度 100/100 已耗尽。
// 期望:Percent 由更紧张的周窗口驱动(≥100 → 球灯熄灭),Used/ResetAt 仍记 5 小时窗口。
func TestKimiFetcher_WeeklyExhausted_PercentAlerts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"usage": {"limit": "100", "used": "100", "resetTime": "2026-09-22T05:10:32.398314Z"},
			"limits": [{"window": {"duration": 300, "timeUnit": "TIME_UNIT_MINUTE"}, "detail": {"limit": "100", "remaining": "100", "resetTime": "2026-09-22T04:10:32.398314Z"}}],
			"usages": {
				"limit_5h": {"used_ratio": 0, "reset_time": "2026-09-22T04:10:32Z"},
				"limit_7d": {"used_ratio": 1, "reset_time": "2026-09-22T05:10:32Z"}
			}
		}`))
	}))
	defer server.Close()

	f := &KimiFetcher{apiKey: "sk-kimi-test", apiURL: server.URL}
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent < 99.9 {
		t.Errorf("expected weekly exhausted to drive Percent ~100, got %f", result.Percent)
	}
	// 5 小时窗口仍驱动 Used/Total/ResetAt
	if result.Used != 0 || result.Total != 100 {
		t.Errorf("expected Used=0 Total=100 (5h window), got Used=%f Total=%f", result.Used, result.Total)
	}
	if result.ResetAt != "2026-09-22T04:10:32.398314Z" {
		t.Errorf("expected 5h reset time, got '%s'", result.ResetAt)
	}
	if !strings.Contains(result.Remaining, "5小时") || !strings.Contains(result.Remaining, "周") {
		t.Errorf("expected both 5小时 and 周 in Remaining, got '%s'", result.Remaining)
	}
}

// 仅有 usages.limit_5h / limit_7d 比率时也应能解析出两窗口。
func TestKimiFetcher_UsagesRatioOnly_ParsesBothWindows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"limits": [{"window": {"duration": 300, "timeUnit": "TIME_UNIT_MINUTE"}, "detail": {"limit": "100", "remaining": "40", "resetTime": "2026-09-22T04:10:32Z"}}],
			"usages": {
				"limit_5h": {"used_ratio": 0.6, "reset_time": "2026-09-22T04:10:32Z"},
				"limit_7d": {"used_ratio": 0.3, "reset_time": "2026-09-22T05:10:32Z"}
			}
		}`))
	}))
	defer server.Close()

	f := &KimiFetcher{apiKey: "sk-kimi-test", apiURL: server.URL}
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent < 59.9 || result.Percent > 60.1 {
		t.Errorf("expected 5h 60%% to drive Percent, got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "周 30%") {
		t.Errorf("expected weekly 30%% in Remaining, got '%s'", result.Remaining)
	}
}

// 周额度低于 5 小时窗口时不得压低 Percent。
func TestKimiFetcher_WeeklyLowerThanSession_KeepsSessionPercent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"usage": {"limit": "100", "used": "20", "resetTime": "2026-09-22T05:10:32Z"},
			"limits": [{"window": {"duration": 300, "timeUnit": "TIME_UNIT_MINUTE"}, "detail": {"limit": "100", "remaining": "25", "resetTime": "2026-09-22T04:10:32Z"}}]
		}`))
	}))
	defer server.Close()

	f := &KimiFetcher{apiKey: "sk-kimi-test", apiURL: server.URL}
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent < 74.9 || result.Percent > 75.1 {
		t.Errorf("expected 5h 75%% to stay, got %f", result.Percent)
	}
}

func TestKimiFetcher_LegacyArrayResponse_StillParses(t *testing.T) {
	// 兼容旧版 {"data":[{model_name:"all",...}]} 响应
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": [
				{
					"model_name": "all",
					"used": 300,
					"limit": 1000,
					"remaining": 700,
					"reset_at": "2026-07-21T00:00:00Z"
				},
				{
					"model_name": "kimi-k2-0905",
					"used": 50,
					"limit": 200,
					"reset_at": "2026-07-21T00:00:00Z"
				}
			]
		}`))
	}))
	defer server.Close()

	f := &KimiFetcher{apiKey: "sk-kimi-test", apiURL: server.URL}
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
	// 300/1000 = 30%
	if result.Percent < 29.9 || result.Percent > 30.1 {
		t.Errorf("expected ~30%%, got %f%%", result.Percent)
	}
	if result.ResetAt != "2026-07-21T00:00:00Z" {
		t.Errorf("expected ResetAt '2026-07-21T00:00:00Z', got '%s'", result.ResetAt)
	}
}

func TestKimiFetcher_Unauthorized_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	f := &KimiFetcher{apiKey: "sk-kimi-bad", apiURL: server.URL}
	result := f.Fetch()
	if result.Error == "" {
		t.Fatal("expected error for 401")
	}
}

func TestKimiFetcher_EmptyUsage_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	f := &KimiFetcher{apiKey: "sk-kimi-test", apiURL: server.URL}
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error when response has no usage data")
	}
}
