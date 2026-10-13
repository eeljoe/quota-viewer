package fetcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// usageJSON 拼出 /zen/go/v1/usage 的响应体:每个窗口的 JSON 片段由调用方给出,
// 空片段表示该窗口在响应中缺失。
func usageJSON(rolling, weekly, monthly string) string {
	var parts []string
	for _, w := range []struct{ name, body string }{
		{"rolling", rolling}, {"weekly", weekly}, {"monthly", monthly},
	} {
		if w.body != "" {
			parts = append(parts, fmt.Sprintf("%q:%s", w.name, w.body))
		}
	}
	return `{"usage":{` + strings.Join(parts, ",") + `}}`
}

// window 拼出单个窗口的 JSON 片段。
func window(percent string, resetsAt string) string {
	body := `"status":"ok","percent":` + percent
	if resetsAt != "" {
		body += `,"resetsAt":` + resetsAt
	}
	return "{" + body + "}"
}

// newUsageServer 起一个 httptest 服务返回给定响应体,并校验请求契约
// (Bearer 头、路径 /zen/go/v1/usage、User-Agent 头名与值)。
func newUsageServer(t *testing.T, body string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/zen/go/v1/usage" {
			t.Errorf("expected path /zen/go/v1/usage, got %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("expected Authorization Bearer header, got %q", r.Header.Get("Authorization"))
		}
		// 头名与值都要钉死:只断言"某个 UA 头非空"时,把实现改成 X-User-Agent 仍会
		// 通过(Go 的 http.Header.Get 不区分大小写,但换名后取值会落空 → 这里必须红)。
		if got, want := r.Header.Get("User-Agent"), "quota-viewer/1.0"; got != want {
			t.Errorf("expected User-Agent header %q, got %q", want, got)
		}
		w.Header().Set("Content-Type", "application/json")
		if status != 0 {
			w.WriteHeader(status)
		}
		w.Write([]byte(body))
	}))
}

// newFetcher 造一个凭证已注入、且不会读真实 home 的 fetcher。
func newFetcher(t *testing.T, apiKey, baseURL string) *OpenCodeGoFetcher {
	t.Helper()
	f := NewOpenCodeGoFetcher(apiKey)
	f.baseURL = baseURL
	// 兜底隔离:即使实现读 home,也读不到真实凭证。
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	return f
}

// AC-2:真机取回的三条窗口(percent 全 0)⇒ 选 rolling;ResetAt 截到秒精度。
func TestOpenCodeGoFetcher_RealResponse_SelectsRollingAndTruncatesReset(t *testing.T) {
	body := `{"usage":{"rolling":{"status":"ok","percent":0,"resetsAt":"2026-10-13T08:51:47.097Z"},` +
		`"weekly":{"status":"ok","percent":0,"resetsAt":"2026-10-19T00:00:00.000Z"},` +
		`"monthly":{"status":"ok","percent":0,"resetsAt":"2026-11-13T03:06:11.000Z"}}}`
	server := newUsageServer(t, body, 200)
	defer server.Close()

	f := newFetcher(t, "oc_sk_test_key", server.URL)
	result := f.Fetch()

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 0 {
		t.Errorf("expected Percent=0 (percent is already 0-100), got %f", result.Percent)
	}
	if result.Total != 100 {
		t.Errorf("expected Total=100, got %f", result.Total)
	}
	if result.Platform != "OpenCode Go" {
		t.Errorf("expected Platform='OpenCode Go', got %q", result.Platform)
	}
	wantRemaining := "5小时窗口 · 已用 0% · 剩余 100%"
	if result.Remaining != wantRemaining {
		t.Errorf("expected Remaining=%q, got %q", wantRemaining, result.Remaining)
	}
	wantReset := "2026-10-13T08:51:47Z"
	if result.ResetAt != wantReset {
		t.Errorf("expected ResetAt=%q (UTC second precision), got %q", wantReset, result.ResetAt)
	}
}

// AC-3:选最大窗口;rolling 3.1 / weekly 43 / monthly 38.3 → weekly(43)。
func TestOpenCodeGoFetcher_PicksMaxPercent(t *testing.T) {
	body := usageJSON(
		window("3.1", `"2026-10-13T08:51:47Z"`),
		window("43", `"2026-10-19T00:00:00Z"`),
		window("38.3", `"2026-11-13T03:06:11Z"`),
	)
	server := newUsageServer(t, body, 200)
	defer server.Close()

	result := newFetcher(t, "key", server.URL).Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 43 {
		t.Errorf("expected Percent=43 (weekly highest), got %f", result.Percent)
	}
	wantRemaining := "周窗口 · 已用 43% · 剩余 57%"
	if result.Remaining != wantRemaining {
		t.Errorf("expected Remaining=%q, got %q", wantRemaining, result.Remaining)
	}
	if result.ResetAt != "2026-10-19T00:00:00Z" {
		t.Errorf("expected ResetAt from weekly window, got %q", result.ResetAt)
	}
}

// AC-3:同值时 rolling > weekly > monthly。
func TestOpenCodeGoFetcher_TiePrefersRolling(t *testing.T) {
	body := usageJSON(
		window("10", `"2026-10-13T08:51:47Z"`),
		window("10", `"2026-10-19T00:00:00Z"`),
		window("10", `"2026-11-13T03:06:11Z"`),
	)
	server := newUsageServer(t, body, 200)
	defer server.Close()

	result := newFetcher(t, "key", server.URL).Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 10 {
		t.Errorf("expected Percent=10, got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "5小时窗口") {
		t.Errorf("expected rolling window label on tie, got %q", result.Remaining)
	}
	if result.ResetAt != "2026-10-13T08:51:47Z" {
		t.Errorf("expected rolling ResetAt on tie, got %q", result.ResetAt)
	}
}

// AC-4:缺 weekly 字段仍出结果。
func TestOpenCodeGoFetcher_MissingWeeklyStillSucceeds(t *testing.T) {
	body := usageJSON(window("3.1", `"2026-10-13T08:51:47Z"`), "", window("38.3", `"2026-11-13T03:06:11Z"`))
	server := newUsageServer(t, body, 200)
	defer server.Close()

	result := newFetcher(t, "key", server.URL).Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 38.3 {
		t.Errorf("expected Percent=38.3 (monthly), got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "月窗口") {
		t.Errorf("expected 月窗口 in Remaining, got %q", result.Remaining)
	}
}

// AC-4:三窗口全缺 → 未找到有效配额窗口。
func TestOpenCodeGoFetcher_NoWindows_ReturnsNoValidWindowError(t *testing.T) {
	server := newUsageServer(t, `{"usage":{}}`, 200)
	defer server.Close()

	result := newFetcher(t, "key", server.URL).Fetch()
	if result.Error != "未找到有效配额窗口" {
		t.Errorf("expected '未找到有效配额窗口', got %q", result.Error)
	}
}

// AC-4 边界:窗口存在但没有 percent 字段 → 同样算无效窗口。
func TestOpenCodeGoFetcher_WindowsWithoutPercent_ReturnsNoValidWindowError(t *testing.T) {
	server := newUsageServer(t, `{"usage":{"rolling":{"status":"ok","resetsAt":"2026-10-13T08:51:47Z"}}}`, 200)
	defer server.Close()

	result := newFetcher(t, "key", server.URL).Fetch()
	if result.Error != "未找到有效配额窗口" {
		t.Errorf("expected '未找到有效配额窗口', got %q", result.Error)
	}
}

// AC-5:401 → API Key 无效。
func TestOpenCodeGoFetcher_401_ReturnsInvalidAPIKey(t *testing.T) {
	server := newUsageServer(t, `{"type":"error","error":{"type":"AuthError","message":"Unauthorized"}}`, 401)
	defer server.Close()

	result := newFetcher(t, "bad-key", server.URL).Fetch()
	if result.Error != "API Key 无效" {
		t.Errorf("expected 'API Key 无效', got %q", result.Error)
	}
}

// AC-5:403 + EntitlementError → 未订阅文案。
func TestOpenCodeGoFetcher_403_EntitlementError(t *testing.T) {
	body := `{"type":"error","error":{"type":"EntitlementError","message":"OpenCode Go subscription required."}}`
	server := newUsageServer(t, body, 403)
	defer server.Close()

	result := newFetcher(t, "old-sub-key", server.URL).Fetch()
	if result.Error != "未订阅 OpenCode Go 套餐(或订阅已失效)" {
		t.Errorf("expected 未订阅 文案, got %q", result.Error)
	}
}

// AC-5:403 其他 body → API Key 无权限。
func TestOpenCodeGoFetcher_403_OtherError(t *testing.T) {
	for _, body := range []string{
		`{"type":"error","error":{"type":"ForbiddenError","message":"no"}}`,
		`not json at all`,
	} {
		server := newUsageServer(t, body, 403)
		result := newFetcher(t, "key", server.URL).Fetch()
		server.Close()
		if result.Error != "API Key 无权限" {
			t.Errorf("body=%s: expected 'API Key 无权限', got %q", body, result.Error)
		}
	}
}

// AC-5:500 → HTTP 500。
func TestOpenCodeGoFetcher_HTTP500_ReturnsStatusCode(t *testing.T) {
	server := newUsageServer(t, "Internal Server Error", 500)
	defer server.Close()

	result := newFetcher(t, "key", server.URL).Fetch()
	if result.Error != "HTTP 500" {
		t.Errorf("expected 'HTTP 500', got %q", result.Error)
	}
}

// AC-5:200 但 body 不是 JSON → 解析失败文案。
func TestOpenCodeGoFetcher_InvalidJSON_ReturnsParseError(t *testing.T) {
	server := newUsageServer(t, "<html>not json</html>", 200)
	defer server.Close()

	result := newFetcher(t, "key", server.URL).Fetch()
	if result.Error != "无法解析额度数据(响应结构可能已变化)" {
		t.Errorf("expected 解析失败文案, got %q", result.Error)
	}
}

// AC-6:apiKey 非空 → 直接用,且不读 auth.json(用坏 auth.json 反证)。
func TestOpenCodeGoFetcher_NonEmptyKey_DoesNotReadAuthFile(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(usageJSON(window("7", `"2026-10-13T08:51:47Z"`), "", "")))
	}))
	defer server.Close()

	dir := t.TempDir()
	authPath := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(authPath, []byte(`{"opencode-go":{"type":"api","key":"from_auth_file"}}`), 0o600); err != nil {
		t.Fatalf("write auth file: %v", err)
	}

	f := newFetcher(t, "field_key", server.URL)
	f.authFile = authPath
	result := f.Fetch()

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotAuth != "Bearer field_key" {
		t.Errorf("expected field key to win, got %q", gotAuth)
	}
	if result.Percent != 7 {
		t.Errorf("expected Percent=7, got %f", result.Percent)
	}
}

// AC-6:apiKey 为空且 auth.json 有 opencode-go.key → 用文件里的 key。
func TestOpenCodeGoFetcher_EmptyKey_ReadsAuthFile(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(usageJSON(window("12", `"2026-10-13T08:51:47Z"`), "", "")))
	}))
	defer server.Close()

	dir := t.TempDir()
	authPath := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(authPath, []byte(`{"other":{"type":"oauth","key":"nope"},"opencode-go":{"type":"api","key":"auth_file_key"}}`), 0o600); err != nil {
		t.Fatalf("write auth file: %v", err)
	}

	f := newFetcher(t, "", server.URL)
	f.authFile = authPath
	result := f.Fetch()

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotAuth != "Bearer auth_file_key" {
		t.Errorf("expected key from auth.json, got %q", gotAuth)
	}
}

// AC-6:auth.json 里没有 opencode-go / 文件不存在 / 内容不是 JSON → 未配置错误(不 panic)。
func TestOpenCodeGoFetcher_EmptyKey_UnconfiguredCases(t *testing.T) {
	const wantErr = "未配置 OpenCode Go API Key(在 opencode.ai 订阅后复制 API Key 填入,或在 opencode CLI 里 /connect 后自动读取)"

	cases := []struct {
		name    string
		content string // 空 = 不创建文件
	}{
		{name: "no_provider_entry", content: `{"some-other-provider":{"type":"api","key":"x"}}`},
		{name: "provider_without_key", content: `{"opencode-go":{"type":"oauth"}}`},
		{name: "missing_file", content: ""},
		{name: "invalid_json", content: `{not json`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			authPath := filepath.Join(t.TempDir(), "auth.json")
			if tc.content != "" {
				if err := os.WriteFile(authPath, []byte(tc.content), 0o600); err != nil {
					t.Fatalf("write auth file: %v", err)
				}
			}

			f := newFetcher(t, "", "http://127.0.0.1:1") // 不发请求:应在发起前就报未配置
			f.authFile = authPath
			result := f.Fetch()

			if result.Error != wantErr {
				t.Errorf("expected 未配置 文案, got %q", result.Error)
			}
			if result.Platform != "OpenCode Go" {
				t.Errorf("expected Platform='OpenCode Go', got %q", result.Platform)
			}
		})
	}
}

// resetsAt 缺失 / null → ResetAt 留空串,不影响选窗。
func TestOpenCodeGoFetcher_MissingOrNullResetsAt_LeavesResetEmpty(t *testing.T) {
	cases := map[string]string{
		"null":    `{"status":"ok","percent":5,"resetsAt":null}`,
		"absent":  `{"status":"ok","percent":5}`,
		"invalid": `{"status":"ok","percent":5,"resetsAt":"not-a-time"}`,
	}
	for name, w := range cases {
		t.Run(name, func(t *testing.T) {
			server := newUsageServer(t, usageJSON(w, "", ""), 200)
			defer server.Close()

			result := newFetcher(t, "key", server.URL).Fetch()
			if result.Error != "" {
				t.Fatalf("unexpected error: %s", result.Error)
			}
			if result.Percent != 5 {
				t.Errorf("expected Percent=5, got %f", result.Percent)
			}
			if result.ResetAt != "" {
				t.Errorf("expected empty ResetAt, got %q", result.ResetAt)
			}
		})
	}
}

// 网络错误 → 请求失败文案(不 panic)。
func TestOpenCodeGoFetcher_NetworkError(t *testing.T) {
	server := newUsageServer(t, `{}`, 200)
	url := server.URL
	server.Close() // 关掉服务 → 连接失败

	result := newFetcher(t, "key", url).Fetch()
	if !strings.Contains(result.Error, "请求失败") {
		t.Errorf("expected 请求失败 文案, got %q", result.Error)
	}
}

// Fetch 结果字段与仓库惯例一致:用量型显式声明 Kind=KindUsage
// (见 docs/ADDING_A_PROVIDER.md:131,与 commandcode/factorydroid/minimax/opendesign 同形)。
func TestOpenCodeGoFetcher_ResultFields(t *testing.T) {
	server := newUsageServer(t, usageJSON(window("5", `"2026-10-13T08:51:47.097Z"`), "", ""), 200)
	defer server.Close()

	before := time.Now()
	result := newFetcher(t, "key", server.URL).Fetch()
	if result.LastUpdated.Before(before) {
		t.Errorf("expected LastUpdated >= test start, got %v", result.LastUpdated)
	}
	if result.Kind != KindUsage {
		t.Errorf("expected Kind=%q (explicitly declared), got %q", KindUsage, result.Kind)
	}
}
