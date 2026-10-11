package fetcher

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// odWorkspaceID 是 2026-10-11 实测 /api/v1/workspaces/current 返回的 workspaceId。
const odWorkspaceID = "p619hvqak6gwawty2rfzl7mb"

// 两条控制面路由的字面量。测试不引用实现里的常量与拼接函数——实现把路径、
// 窗口时长识别或 workspaceId 拼错时必须在这里变红。
const (
	odDiscoveryPath = "/api/v1/workspaces/current"
	odUsagePath     = "/api/v1/workspaces/" + odWorkspaceID + "/billing/coding-plan-usage"

	odDuration5h = 18000  // 5 小时窗口
	odDuration7d = 604800 // 7 天窗口
)

// odRequest 记录假控制面收到的请求。
type odRequest struct {
	path   string
	cookie string
}

// odWindow 拼一个额度窗口 JSON(字段名与官方响应一致)。
func odWindow(durationSeconds int, usedCredits, limitCredits, resetsAt string) string {
	return fmt.Sprintf(`{"policyId":"model_limit_policy_%d","durationSeconds":%d,"resetMode":"anchored_recurring","usedCredits":"%s","limitCredits":"%s","remainingCredits":"0","windowStart":"2026-10-11T04:35:11.000Z","resetsAt":"%s"}`,
		durationSeconds, durationSeconds, usedCredits, limitCredits, resetsAt)
}

// newODServerRaw 起一个假控制面:发现接口与额度接口各自返回给定的状态码与响应体,
// 收到的请求记录进返回的切片。绝不发起真实网络请求。
func newODServerRaw(t *testing.T, discoverStatus int, discoverBody string, usageStatus int, usageBody string) (*httptest.Server, *[]odRequest) {
	t.Helper()
	got := &[]odRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = append(*got, odRequest{path: r.URL.Path, cookie: r.Header.Get("Cookie")})
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case odDiscoveryPath:
			w.WriteHeader(discoverStatus)
			w.Write([]byte(discoverBody))
		case odUsagePath:
			w.WriteHeader(usageStatus)
			w.Write([]byte(usageBody))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"unexpected_path"}`))
		}
	}))
	t.Cleanup(server.Close)
	return server, got
}

// newODServer 是正常路径的假控制面:发现返回 odWorkspaceID,额度返回给定 windows JSON。
func newODServer(t *testing.T, windowsJSON string) (*httptest.Server, *[]odRequest) {
	t.Helper()
	return newODServerRaw(t,
		http.StatusOK, `{"workspaceId":"`+odWorkspaceID+`"}`,
		http.StatusOK, `{"workspaceId":"`+odWorkspaceID+`","eligible":true,"tier":"go","windows":`+windowsJSON+`}`)
}

// odLoginPath 是假登录页:Cookie 过期时控制面用 302 把请求甩到这里。
// httptest 里 302 需要 Location 头才像真跳转,用例显式指向这个假登录页。
const odLoginPath = "/cloud/login"

// odAssertCookieExpired 断言错误是「Cookie 失效」语义(与 401/403 同一句提示)。
func odAssertCookieExpired(t *testing.T, what, errMsg string) {
	t.Helper()
	if errMsg == "" {
		t.Fatalf("%s: expected Cookie-expired error, got empty", what)
	}
	for _, want := range []string{"Cookie", "过期", "open-design.ai"} {
		if !strings.Contains(errMsg, want) {
			t.Errorf("%s: expected %q in error, got '%s'", what, want, errMsg)
		}
	}
}

// odAssertNotFollowed 断言实现没有跟随 302 到登录页(请求记录里不得出现 odLoginPath)。
func odAssertNotFollowed(t *testing.T, requests []odRequest) {
	t.Helper()
	for _, req := range requests {
		if req.path == odLoginPath {
			t.Errorf("跟随了 302 跳转请求 %s,应禁止自动重定向: %+v", odLoginPath, requests)
		}
	}
}

// newODRedirectServer 起一个假控制面,在指定的那一步返回 302 → Location 假登录页。
// 假登录页回 200 HTML(平台跳登录页的真实形态):跟随跳转的实现会把这段 HTML 当 JSON
// 解析,报成「解析响应失败」而丢掉「请重新登录」指引 —— 正是本用例要拦的行为。
// redirectDiscovery=true 时发现接口即 302;false 时发现正常、额度接口 302。
func newODRedirectServer(t *testing.T, redirectDiscovery bool) (*httptest.Server, *[]odRequest) {
	t.Helper()
	got := &[]odRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = append(*got, odRequest{path: r.URL.Path, cookie: r.Header.Get("Cookie")})
		switch r.URL.Path {
		case odLoginPath:
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html><body>Sign in to open-design.ai</body></html>`))
		case odDiscoveryPath:
			if redirectDiscovery {
				w.Header().Set("Location", odLoginPath)
				w.WriteHeader(http.StatusFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"workspaceId":"` + odWorkspaceID + `"}`))
		case odUsagePath:
			w.Header().Set("Location", odLoginPath)
			w.WriteHeader(http.StatusFound)
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"unexpected_path"}`))
		}
	}))
	t.Cleanup(server.Close)
	return server, got
}

// odFetch 用给定 baseURL 与 Cookie 跑一次 Fetch(注入 httptest server)。
func odFetch(t *testing.T, serverURL, cookie string) QuotaResult {
	t.Helper()
	f := NewOpenDesignFetcher(cookie)
	f.baseURL = serverURL
	return f.Fetch()
}

// odAssertPercent 比较百分比(期望值由用例手工从 fixture 算出,容差 1e-6)。
func odAssertPercent(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("%s: expected %.6f, got %.6f", what, want, got)
	}
}

// 空 Cookie → 报错,且不发任何请求。
func TestOpenDesignFetcher_EmptyCookie_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("空 Cookie 不应发出请求,却请求了 %s", r.URL.Path)
	}))
	defer server.Close()

	f := NewOpenDesignFetcher("   ")
	f.baseURL = server.URL
	result := f.Fetch()

	if result.Error == "" {
		t.Error("expected error for empty cookie")
	}
	if result.Platform != "Open Design" {
		t.Errorf("expected platform 'Open Design', got '%s'", result.Platform)
	}
}

// 实测原文(2026-10-11):5h 1275/50000 = 2.55%,周 1275/150000 = 0.85%。
// Percent 取更紧张的 5h;ResetAt 是 5h 的 resetsAt(毫秒精度 → 项目惯用秒精度 UTC)。
func TestOpenDesignFetcher_TwoWindows_RealPayload(t *testing.T) {
	windows := "[" + odWindow(odDuration5h, "1275", "50000", "2026-10-11T09:35:11.000Z") +
		"," + odWindow(odDuration7d, "1275", "150000", "2026-10-18T04:35:11.000Z") + "]"
	server, _ := newODServer(t, windows)

	result := odFetch(t, server.URL, "next-auth.session-token=abc")

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Kind != KindUsage {
		t.Errorf("expected KindUsage, got '%s'", result.Kind)
	}
	if result.Total != 100 {
		t.Errorf("expected Total=100, got %f", result.Total)
	}
	odAssertPercent(t, "Used(5小时窗口已用%)", result.Used, 2.55)
	odAssertPercent(t, "Percent=max(5h,7d)", result.Percent, 2.55)
	if result.ResetAt != "2026-10-11T09:35:11Z" {
		t.Errorf("expected ResetAt='2026-10-11T09:35:11Z', got '%s'", result.ResetAt)
	}
	if !strings.Contains(result.Remaining, "5小时") || !strings.Contains(result.Remaining, "周") {
		t.Errorf("expected 5小时 + 周 segments in Remaining, got '%s'", result.Remaining)
	}
}

// 周窗口比 5 小时紧张 → Percent 必须等于周值(窗口告警契约)。同时把 windows 数组
// 倒序(7 天在前):窗口必须按 durationSeconds 识别,按数组下标取窗口的实现会让
// Used/ResetAt 落到周窗口上而变红。
func TestOpenDesignFetcher_WeeklyTighter_DrivesPercent(t *testing.T) {
	windows := "[" + odWindow(odDuration7d, "90000", "150000", "2026-10-18T04:35:11.000Z") +
		"," + odWindow(odDuration5h, "1000", "50000", "2026-10-11T09:35:11.000Z") + "]"
	server, _ := newODServer(t, windows)

	result := odFetch(t, server.URL, "c=1")

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	odAssertPercent(t, "Percent=max(5h,7d)", result.Percent, 60) // 90000/150000
	odAssertPercent(t, "Used(5小时窗口已用%)", result.Used, 2)         // 1000/50000
	if result.ResetAt != "2026-10-11T09:35:11Z" {
		t.Errorf("ResetAt 必须来自 5 小时窗口, got '%s'", result.ResetAt)
	}
	if result.Remaining != "5小时 2.0% 已用 · 周 60.0% 已用" {
		t.Errorf("expected '5小时 2.0%% 已用 · 周 60.0%% 已用', got '%s'", result.Remaining)
	}
}

// 只有 5 小时窗口 → 不崩、不输出 NaN/Inf,Remaining 只含 5 小时段。
func TestOpenDesignFetcher_OnlyFiveHourWindow_NoWeeklySegment(t *testing.T) {
	server, _ := newODServer(t, "["+odWindow(odDuration5h, "1000", "50000", "2026-10-11T09:35:11Z")+"]")

	result := odFetch(t, server.URL, "c=1")

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	odAssertPercent(t, "Percent", result.Percent, 2)
	odAssertPercent(t, "Used", result.Used, 2)
	if result.Remaining != "5小时 2.0% 已用" {
		t.Errorf("expected '5小时 2.0%% 已用', got '%s'", result.Remaining)
	}
}

// 没有 5 小时窗口(windows 为空 / 只有 7 天窗口)→ 无法驱动主展示,报错而不是崩/出 NaN。
func TestOpenDesignFetcher_NoFiveHourWindow_ReturnsError(t *testing.T) {
	cases := []struct {
		name    string
		windows string
	}{
		{"windows 为空数组", "[]"},
		{"只有 7 天窗口", "[" + odWindow(odDuration7d, "90000", "150000", "2026-10-18T04:35:11.000Z") + "]"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server, _ := newODServer(t, c.windows)

			result := odFetch(t, server.URL, "c=1")

			if result.Error == "" {
				t.Errorf("expected error when 5h window is missing (%s)", c.name)
			}
		})
	}
}

// 5 小时窗口的 credits 不可解析(limit 为 0)→ 报错,不做除零。
func TestOpenDesignFetcher_FiveHourLimitZero_ReturnsError(t *testing.T) {
	server, _ := newODServer(t, "["+odWindow(odDuration5h, "100", "0", "2026-10-11T09:35:11Z")+"]")

	result := odFetch(t, server.URL, "c=1")

	if result.Error == "" {
		t.Fatalf("expected error for zero limit, got Percent=%v Remaining='%s'", result.Percent, result.Remaining)
	}
}

// 7 天窗口的 credits 不可解析(limit 为 0)→ 只写 5 小时段,不输出 NaN/Inf 周段。
func TestOpenDesignFetcher_WeeklyLimitZero_SkipsWeeklySegment(t *testing.T) {
	windows := "[" + odWindow(odDuration5h, "1000", "50000", "2026-10-11T09:35:11Z") +
		"," + odWindow(odDuration7d, "500", "0", "2026-10-18T04:35:11.000Z") + "]"
	server, _ := newODServer(t, windows)

	result := odFetch(t, server.URL, "c=1")

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	odAssertPercent(t, "Percent(仅 5 小时可用)", result.Percent, 2)
	if result.Remaining != "5小时 2.0% 已用" {
		t.Errorf("expected '5小时 2.0%% 已用' (no NaN segment), got '%s'", result.Remaining)
	}
}

// 第一步(Cookie 失效)→ 提示重新登录 open-design.ai。
func TestOpenDesignFetcher_401Discovery_ReturnsCookieExpired(t *testing.T) {
	server, _ := newODServerRaw(t, http.StatusUnauthorized, `{"error":"unauthorized"}`, http.StatusOK, `{}`)

	result := odFetch(t, server.URL, "c=1")

	for _, want := range []string{"Cookie", "过期", "open-design.ai"} {
		if !strings.Contains(result.Error, want) {
			t.Errorf("expected %q in error, got '%s'", want, result.Error)
		}
	}
}

// 第二步(Cookie 权限不足)同样映射为 Cookie 失效提示,不能被吞成 HTTP 403。
func TestOpenDesignFetcher_403OnUsage_ReturnsCookieExpired(t *testing.T) {
	server, _ := newODServerRaw(t,
		http.StatusOK, `{"workspaceId":"`+odWorkspaceID+`"}`,
		http.StatusForbidden, `{"error":"workspace_member_required"}`)

	result := odFetch(t, server.URL, "c=1")

	if result.Error == "" {
		t.Fatal("expected error for 403")
	}
	if !strings.Contains(result.Error, "Cookie") {
		t.Errorf("expected Cookie hint for 403, got '%s'", result.Error)
	}
}

// 第一步就 302(平台把过期 Cookie 甩到登录页)→ 必须按 Cookie 失效处理:
// 既不能跟随跳转拿 200 HTML 报成「解析响应失败」,也不能落成裸 HTTP 302;
// 且发现步骤失败后不再去请求额度接口。
func TestOpenDesignFetcher_302OnDiscovery_ReturnsCookieExpiredAndDoesNotFollow(t *testing.T) {
	server, requests := newODRedirectServer(t, true)

	result := odFetch(t, server.URL, "c=1")

	odAssertCookieExpired(t, "discovery 302", result.Error)
	odAssertNotFollowed(t, *requests)
	if len(*requests) != 1 {
		t.Errorf("expected only the discovery request, got %d: %+v", len(*requests), *requests)
	}
}

// 第二步 302 → 同样按 Cookie 失效处理,且不跟随跳转到登录页。
func TestOpenDesignFetcher_302OnUsage_ReturnsCookieExpiredAndDoesNotFollow(t *testing.T) {
	server, requests := newODRedirectServer(t, false)

	result := odFetch(t, server.URL, "c=1")

	odAssertCookieExpired(t, "usage 302", result.Error)
	odAssertNotFollowed(t, *requests)
	if len(*requests) != 2 {
		t.Errorf("expected discovery + usage requests only, got %d: %+v", len(*requests), *requests)
	}
}

// 非 200(500)→ 通用 HTTP 错误。
func TestOpenDesignFetcher_HTTP500_ReturnsStatusCode(t *testing.T) {
	server, _ := newODServerRaw(t,
		http.StatusOK, `{"workspaceId":"`+odWorkspaceID+`"}`,
		http.StatusInternalServerError, ``)

	result := odFetch(t, server.URL, "c=1")

	if !strings.Contains(result.Error, "HTTP 500") {
		t.Errorf("expected 'HTTP 500' in error, got '%s'", result.Error)
	}
}

// 额度响应不是 JSON → 解析错误。
func TestOpenDesignFetcher_BadUsageJSON_ReturnsParseError(t *testing.T) {
	server, _ := newODServerRaw(t,
		http.StatusOK, `{"workspaceId":"`+odWorkspaceID+`"}`,
		http.StatusOK, `not json`)

	result := odFetch(t, server.URL, "c=1")

	if !strings.Contains(result.Error, "解析") {
		t.Errorf("expected parse error, got '%s'", result.Error)
	}
}

// 发现接口返回空/缺字段/malformed workspaceId → 报错、不 panic,且不再去请求额度接口
// (拿 current 当 id 会撞 400 workspace_member_required)。
func TestOpenDesignFetcher_WorkspaceDiscoveryBroken_ReturnsError(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"空 workspaceId", `{"workspaceId":""}`},
		{"缺 workspaceId 字段", `{}`},
		{"malformed JSON", `{"workspaceId":`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server, requests := newODServerRaw(t,
				http.StatusOK, c.body,
				http.StatusOK, `{"windows":[]}`)

			result := odFetch(t, server.URL, "c=1")

			if result.Error == "" {
				t.Fatalf("expected error for %s", c.name)
			}
			if len(*requests) != 1 {
				t.Errorf("expected only the discovery request, got %d: %+v", len(*requests), *requests)
			}
		})
	}
}

// 两步请求都必须带 Cookie 头,且两端 Cookie 都做过 TrimSpace。
func TestOpenDesignFetcher_SendsCookieHeader(t *testing.T) {
	server, requests := newODServer(t, "["+odWindow(odDuration5h, "1000", "50000", "2026-10-11T09:35:11Z")+"]")

	result := odFetch(t, server.URL, "  next-auth.session-token=abc; other=1  ")

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if len(*requests) != 2 {
		t.Fatalf("expected 2 requests (发现 + 额度), got %d: %+v", len(*requests), *requests)
	}
	for _, req := range *requests {
		if req.cookie != "next-auth.session-token=abc; other=1" {
			t.Errorf("request %s: expected trimmed cookie header, got '%s'", req.path, req.cookie)
		}
	}
}

// 额度接口路径必须用发现到的真实 workspaceId 拼接(不能用 current 别名)。
func TestOpenDesignFetcher_UsesDiscoveredWorkspaceIDInPath(t *testing.T) {
	server, requests := newODServer(t, "["+odWindow(odDuration5h, "1000", "50000", "2026-10-11T09:35:11Z")+"]")

	result := odFetch(t, server.URL, "c=1")

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	want := []string{odDiscoveryPath, odUsagePath}
	if len(*requests) != len(want) {
		t.Fatalf("expected %d requests, got %d: %+v", len(want), len(*requests), *requests)
	}
	for i, path := range want {
		if (*requests)[i].path != path {
			t.Errorf("requests[%d] path: expected %s, got %s", i, path, (*requests)[i].path)
		}
	}
}
