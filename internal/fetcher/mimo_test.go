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
			"code": 0,
			"message": "",
			"data": {
				"usage": {
					"percent": 0.22,
					"items": [
						{"name": "plan_total_token", "used": 8331114938, "limit": 38000000000, "percent": 0.22},
						{"name": "compensation_total_token", "used": 0, "limit": 0, "percent": 0}
					]
				}
			}
		}`))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Used != 8331114938 {
		t.Errorf("expected Used=8331114938, got %f", result.Used)
	}
	if result.Total != 38000000000 {
		t.Errorf("expected Total=38000000000, got %f", result.Total)
	}
	// 8331114938/38000000000 ≈ 21.9%
	if result.Percent < 21.8 || result.Percent > 22.0 {
		t.Errorf("expected ~21.9%%, got %f%%", result.Percent)
	}
	if !strings.Contains(result.Remaining, "Credits") {
		t.Errorf("expected 'Credits' in Remaining, got '%s'", result.Remaining)
	}
}

func TestMiMoFetcher_JSONResponse_MonthUsage_FallbackPercent(t *testing.T) {
	// 当 items 为空时,用顶层 percent 兜底
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"data": {
				"usage": {
					"percent": 0.35,
					"items": []
				}
			}
		}`))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent < 34.9 || result.Percent > 35.1 {
		t.Errorf("expected ~35%%, got %f%%", result.Percent)
	}
}

func TestMiMoFetcher_JSONNoUsageData_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":0,"data":{"usage":{"percent":0,"items":[]}}}`))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error when JSON has no usage data")
	}
}

// TestMiMo_FetchAll_TwoBuckets 同一 Cookie 返回两个额度桶(套餐额度 + 补偿额度)时,
// 必须产出两个监控对象。旧实现只取第一个有 limit 的 item,补偿额度被静默丢弃。
func TestMiMo_FetchAll_TwoBuckets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code": 0,
			"data": {"usage": {"percent": 0.22, "items": [
				{"name": "plan_total_token", "used": 38000000000, "limit": 40000000000, "percent": 0.95},
				{"name": "compensation_total_token", "used": 100, "limit": 1000, "percent": 0.1}
			]}}
		}`))
	}))
	defer server.Close()

	rs := NewMiMoFetcher("session=abc", server.URL).FetchAll()
	if len(rs) != 2 {
		t.Fatalf("监控对象数: got %d, want 2", len(rs))
	}
	if rs[0].SubLabel != "套餐额度" || rs[1].SubLabel != "补偿额度" {
		t.Errorf("桶名应翻译成可读标签: %q / %q", rs[0].SubLabel, rs[1].SubLabel)
	}
	if got := rs[1].Percent; got != 10 {
		t.Errorf("补偿额度 percent: got %v, want 10", got)
	}
	if rs[0].Abbr != "" || rs[1].Abbr == "" {
		t.Errorf("球格缩写应逐桶区分: %q / %q", rs[0].Abbr, rs[1].Abbr)
	}
}

// TestMiMo_Fetch_仍取第一个桶 单对象入口保持旧语义(既有调用方与测试不受影响)。
func TestMiMo_Fetch_StillFirstBucket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"usage":{"percent":0.22,"items":[
			{"name":"plan_total_token","used":38000000000,"limit":40000000000,"percent":0.95},
			{"name":"compensation_total_token","used":100,"limit":1000,"percent":0.1}
		]}}}`))
	}))
	defer server.Close()

	f := NewMiMoFetcher("session=abc", server.URL)
	single := f.Fetch()
	all := f.FetchAll()
	if single.Percent != all[0].Percent || single.Remaining != all[0].Remaining {
		t.Errorf("Fetch 应等于 FetchAll()[0]: %+v vs %+v", single, all[0])
	}
}
