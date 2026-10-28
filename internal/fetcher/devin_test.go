package fetcher

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// 本文件覆盖票 01 的 AS-1..AS-9:手填 token 分支的请求形状、字段容错、
// 剩余→已用反转、驱动窗口、倒计时、隐藏窗口、Remaining 文案与错误映射。
// 全部走 httptest 注入 server,不发真实网络请求;本机登录态分支属票 02,此处不涉及。

// devinTestServer 起一个 httptest server,并返回与之绑定的 fetcher(注入 serverURL)。
func devinTestServer(t *testing.T, handler http.HandlerFunc) (*DevinFetcher, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	f := NewDevinFetcher("dv-test-token")
	f.serverURL = srv.URL
	return f, srv
}

// devinOKBody 是一份实测形状的 200 响应:日剩余 20(已用 80)、周剩余 70(已用 30),
// 重置时间是 protojson 的字符串形式,overageBalanceMicros 也是字符串。
const devinOKBody = `{"userStatus":{"planStatus":{
  "dailyQuotaRemainingPercent":20,
  "weeklyQuotaRemainingPercent":70,
  "dailyQuotaResetAtUnix":"1793260800",
  "weeklyQuotaResetAtUnix":"1793520000",
  "overageBalanceMicros":"70870000",
  "planInfo":{"planName":"Pro","hideDailyQuota":false,"hideWeeklyQuota":false}}}}`

// devinCheckBase 断言错误路径仍保留 Platform / Kind / LastUpdated
// (registry_test 会断言空凭证 Fetch 也有 Platform,错误路径必须同口径)。
func devinCheckBase(t *testing.T, r QuotaResult) {
	t.Helper()
	if r.Platform != "Devin" {
		t.Errorf("Platform: got %q, want %q", r.Platform, "Devin")
	}
	if r.Kind != KindUsage {
		t.Errorf("Kind: got %q, want %q", r.Kind, KindUsage)
	}
	if r.LastUpdated.IsZero() {
		t.Error("LastUpdated 不得为零值(错误路径也要保留)")
	}
}

// ---------- AS-1 请求形状 ----------

// TestDevinStatusURL 纯字符串断言(不发请求):空 server 走默认端点,非空则去尾斜杠拼同一路径。
func TestDevinStatusURL(t *testing.T) {
	const path = "/exa.seat_management_pb.SeatManagementService/GetUserStatus"
	cases := []struct {
		server string
		want   string
	}{
		{"", "https://server.codeium.com" + path},
		{"https://server.codeium.com", "https://server.codeium.com" + path},
		{"https://server.codeium.com/", "https://server.codeium.com" + path},
		{"https://example.test///", "https://example.test" + path},
	}
	for _, c := range cases {
		if got := devinStatusURL(c.server); got != c.want {
			t.Errorf("devinStatusURL(%q) = %q, want %q", c.server, got, c.want)
		}
	}
}

// TestDevinFetcher_RequestShape 断言实际发出去的请求:POST、路径、三个头、UA、body 关键字段。
func TestDevinFetcher_RequestShape(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotCT     string
		gotCPV    string
		gotAccept string
		gotUA     string
		gotMeta   map[string]string
		hit       int32
	)
	f, _ := devinTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hit, 1)
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotCT = r.Header.Get("Content-Type")
		gotCPV = r.Header.Get("Connect-Protocol-Version")
		gotAccept = r.Header.Get("Accept")
		gotUA = r.Header.Get("User-Agent")

		raw, _ := io.ReadAll(r.Body)
		var payload struct {
			Metadata map[string]string `json:"metadata"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Errorf("body 不是合法 JSON: %v (%s)", err, raw)
		}
		gotMeta = payload.Metadata

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(devinOKBody))
	})

	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if hit != 1 {
		t.Errorf("期望恰好 1 次请求,实际 %d", hit)
	}
	if gotMethod != "POST" {
		t.Errorf("Method: got %q, want POST", gotMethod)
	}
	if gotPath != "/exa.seat_management_pb.SeatManagementService/GetUserStatus" {
		t.Errorf("Path: got %q", gotPath)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type: got %q", gotCT)
	}
	if gotCPV != "1" {
		t.Errorf("Connect-Protocol-Version: got %q, want \"1\"", gotCPV)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept: got %q", gotAccept)
	}
	if gotUA != devinUserAgent {
		t.Errorf("User-Agent: got %q, want 包级常量 %q", gotUA, devinUserAgent)
	}
	if gotMeta["api_key"] != "dv-test-token" {
		t.Errorf("metadata.api_key: got %q", gotMeta["api_key"])
	}
	if gotMeta["ide_version"] != "1.126.0" {
		t.Errorf("metadata.ide_version: got %q, want \"1.126.0\"", gotMeta["ide_version"])
	}
	if gotMeta["ide_name"] != "windsurf" || gotMeta["extension_name"] != "windsurf" {
		t.Errorf("metadata.ide_name/extension_name 必须是 windsurf, got %+v", gotMeta)
	}
	if gotMeta["session_id"] == "" {
		t.Error("metadata.session_id 不得为空(接口要求有占位值)")
	}
}

// ---------- AS-2 字段容错 ----------

// TestDevinParsePlanStatus_Int64BothForms 重置时间与超额余额,字符串/数字两种形态都要能解析。
func TestDevinParsePlanStatus_Int64BothForms(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"字符串", `{"userStatus":{"planStatus":{"dailyQuotaRemainingPercent":50,
			"dailyQuotaResetAtUnix":"1793260800","weeklyQuotaResetAtUnix":"1793520000",
			"overageBalanceMicros":"70870000"}}}`},
		{"数字", `{"userStatus":{"planStatus":{"dailyQuotaRemainingPercent":50,
			"dailyQuotaResetAtUnix":1793260800,"weeklyQuotaResetAtUnix":1793520000,
			"overageBalanceMicros":70870000}}}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, err := devinParsePlanStatus([]byte(c.body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if st.DailyResetAt != 1793260800 {
				t.Errorf("DailyResetAt: got %d, want 1793260800", st.DailyResetAt)
			}
			if st.WeeklyResetAt != 1793520000 {
				t.Errorf("WeeklyResetAt: got %d, want 1793520000", st.WeeklyResetAt)
			}
			if st.OverageMicros == nil || *st.OverageMicros != 70870000 {
				t.Errorf("OverageMicros: got %v, want 70870000", st.OverageMicros)
			}
		})
	}
}

// TestDevinParsePlanStatus_OverageMissingIsNil 缺失超额余额必须是 nil,不是 0 指针。
func TestDevinParsePlanStatus_OverageMissingIsNil(t *testing.T) {
	st, err := devinParsePlanStatus([]byte(`{"userStatus":{"planStatus":{"dailyQuotaRemainingPercent":50}}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.OverageMicros != nil {
		t.Errorf("缺失 overageBalanceMicros 时应为 nil, got %d", *st.OverageMicros)
	}
	if st.DailyResetAt != 0 {
		t.Errorf("缺失 dailyQuotaResetAtUnix 时应为 0, got %d", st.DailyResetAt)
	}
}

// TestDevinParsePlanStatus_PercentTolerant 百分比即便被序列化成字符串/小数也要能取到值。
func TestDevinParsePlanStatus_PercentTolerant(t *testing.T) {
	st, err := devinParsePlanStatus([]byte(`{"userStatus":{"planStatus":{
		"dailyQuotaRemainingPercent":"35.5","weeklyQuotaRemainingPercent":12}}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.DailyRemaining == nil || *st.DailyRemaining != 35.5 {
		t.Errorf("DailyRemaining: got %v, want 35.5", st.DailyRemaining)
	}
	if st.WeeklyRemaining == nil || *st.WeeklyRemaining != 12 {
		t.Errorf("WeeklyRemaining: got %v, want 12", st.WeeklyRemaining)
	}
}

// TestDevinParsePlanStatus_StructuralErrors 结构不符必须报错,不 panic、不渲染成 0% 已用。
func TestDevinParsePlanStatus_StructuralErrors(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"缺 userStatus", `{}`},
		{"缺 planStatus", `{"userStatus":{}}`},
		{"planStatus 为 null", `{"userStatus":{"planStatus":null}}`},
		{"两个窗口都缺", `{"userStatus":{"planStatus":{"planInfo":{"planName":"Pro"}}}}`},
		{"非 JSON", `<html>not json</html>`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := devinParsePlanStatus([]byte(c.body)); err == nil {
				t.Fatal("期望解析错误,实际 nil")
			}
		})
	}
}

// TestDevinParsePlanStatus_NullTolerance 显式 null 的 int64 字段必须解析成功且为 0,
// 不能被当成"结构不符"报错(票面第 4 条要求 devinJSONInt64 同时接受 123 / "123" / null)。
// 百分比字段给 null 表示"本次没给该窗口",按 AS-2 只有两侧都缺才算结构错误。
func TestDevinParsePlanStatus_NullTolerance(t *testing.T) {
	// 三个 int64 字段(日/周重置时间、超额余额)显式 null。
	st, err := devinParsePlanStatus([]byte(`{"userStatus":{"planStatus":{
		"dailyQuotaRemainingPercent":35,
		"dailyQuotaResetAtUnix":null,
		"weeklyQuotaResetAtUnix":null,
		"overageBalanceMicros":null}}}`))
	if err != nil {
		t.Fatalf("null 的 int64 字段必须能解析(票面第 4 条), got error: %v", err)
	}
	if st.DailyResetAt != 0 {
		t.Errorf("dailyQuotaResetAtUnix:null 应解析成 0, got %d", st.DailyResetAt)
	}
	if st.WeeklyResetAt != 0 {
		t.Errorf("weeklyQuotaResetAtUnix:null 应解析成 0, got %d", st.WeeklyResetAt)
	}
	if st.OverageMicros != nil {
		t.Errorf("overageBalanceMicros:null 应为 nil(缺失同义), got %d", *st.OverageMicros)
	}
	if st.DailyRemaining == nil || *st.DailyRemaining != 35 {
		t.Errorf("DailyRemaining: got %v, want 35", st.DailyRemaining)
	}

	// 百分比字段显式 null = 该窗口本次没给值 ⇒ nil,不是 0% 剩余,也不是解析错误。
	st, err = devinParsePlanStatus([]byte(`{"userStatus":{"planStatus":{
		"dailyQuotaRemainingPercent":35,"weeklyQuotaRemainingPercent":null}}}`))
	if err != nil {
		t.Fatalf("单侧百分比为 null 不该报结构错误, got error: %v", err)
	}
	if st.WeeklyRemaining != nil {
		t.Errorf("weeklyQuotaRemainingPercent:null 应为 nil(不是 0), got %v", *st.WeeklyRemaining)
	}
}

// ---------- AS-3 反转与夹取 ----------

// TestDevinApplyPlanStatus_InversionAndClamp 已用 = 100 - 剩余,越界夹到 [0,100]。
func TestDevinApplyPlanStatus_InversionAndClamp(t *testing.T) {
	f := NewDevinFetcher("t")
	cases := []struct {
		name    string
		daily   float64
		weekly  float64
		wantPct float64
	}{
		{"剩余 100 ⇒ 0% 已用", 100, 100, 0},
		{"剩余 40 ⇒ 60% 已用", 40, 90, 60},
		{"剩余 60 ⇒ 40% 已用", 60, 95, 40},
		{"剩余 -10 ⇒ 夹到 100% 已用", -10, 95, 100},
		{"剩余 130 ⇒ 夹到 0% 已用", 130, 100, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := f.applyPlanStatus(&devinPlanStatus{
				DailyRemaining:  devinF(c.daily),
				WeeklyRemaining: devinF(c.weekly),
			})
			if r.Error != "" {
				t.Fatalf("unexpected error: %s", r.Error)
			}
			if r.Percent != c.wantPct {
				t.Errorf("Percent: got %v, want %v", r.Percent, c.wantPct)
			}
			if r.Used != c.wantPct {
				t.Errorf("Used: got %v, want %v", r.Used, c.wantPct)
			}
			if r.Total != 100 {
				t.Errorf("Total: got %v, want 100", r.Total)
			}
		})
	}
}

// ---------- AS-4 最紧张窗口驱动 ----------

// TestDevinApplyPlanStatus_DrivingWindow 驱动窗口 = 可见窗口中已用更大者;Used/Percent 同值。
func TestDevinApplyPlanStatus_DrivingWindow(t *testing.T) {
	f := NewDevinFetcher("t")

	// 日剩余 20(已用 80)、周剩余 70(已用 30)⇒ 日驱动。
	r := f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(20),
		WeeklyRemaining: devinF(70),
		DailyResetAt:    1793260800,
		WeeklyResetAt:   1793520000,
	})
	if r.Error != "" {
		t.Fatalf("unexpected error: %s", r.Error)
	}
	if r.Percent != 80 || r.Used != 80 || r.Total != 100 {
		t.Errorf("日驱动: Percent/Used/Total = %v/%v/%v, want 80/80/100", r.Percent, r.Used, r.Total)
	}
	if r.ResetLabel != "日额度" {
		t.Errorf("日驱动 ResetLabel: got %q, want 日额度", r.ResetLabel)
	}

	// 日剩余 70(已用 30)、周剩余 20(已用 80)⇒ 周驱动。
	r = f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(70),
		WeeklyRemaining: devinF(20),
		DailyResetAt:    1793260800,
		WeeklyResetAt:   1793520000,
	})
	if r.Percent != 80 || r.Used != 80 {
		t.Errorf("周驱动: Percent/Used = %v/%v, want 80/80", r.Percent, r.Used)
	}
	if r.ResetLabel != "周额度" {
		t.Errorf("周驱动 ResetLabel: got %q, want 周额度", r.ResetLabel)
	}
}

// TestDevinApplyPlanStatus_TiePicksDailyWindow 两侧已用并列时必须取日窗口:
// 球面百分比相同看不出差别,但倒计时的标签与时间会错标成周(票面第 6 条「并列取日」)。
// 日/周重置 epoch 故意不同,所以 ResetAt 一条断言就能分开"取了哪个窗口"。
func TestDevinApplyPlanStatus_TiePicksDailyWindow(t *testing.T) {
	f := NewDevinFetcher("t")

	// 日剩余 20、周剩余 20 ⇒ 两侧都已用 80,并列。
	r := f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(20),
		WeeklyRemaining: devinF(20),
		DailyResetAt:    1793260800, // 2026-10-29T08:00:00Z
		WeeklyResetAt:   1793520000, // 2026-11-01T08:00:00Z
	})
	if r.Error != "" {
		t.Fatalf("unexpected error: %s", r.Error)
	}
	if r.Percent != 80 || r.Used != 80 || r.Total != 100 {
		t.Errorf("并列: Percent/Used/Total = %v/%v/%v, want 80/80/100", r.Percent, r.Used, r.Total)
	}
	if r.ResetLabel != "日额度" {
		t.Errorf("并列必须取日: ResetLabel got %q, want 日额度", r.ResetLabel)
	}
	if r.ResetAt != "2026-10-29T08:00:00Z" {
		t.Errorf("并列必须取日的重置时间: ResetAt got %q, want 2026-10-29T08:00:00Z (2026-11-01T08:00:00Z 是周)",
			r.ResetAt)
	}
}

// ---------- AS-5 倒计时 ----------

// TestDevinApplyPlanStatus_ResetAtAndLabel ResetAt = 驱动窗口 epoch 的 UTC RFC3339;
// 驱动窗口无重置时间时两者都留空,不得借另一个可见窗口的时间。
func TestDevinApplyPlanStatus_ResetAtAndLabel(t *testing.T) {
	f := NewDevinFetcher("t")

	// 日驱动 + 日有重置时间。
	r := f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(10),
		WeeklyRemaining: devinF(90),
		DailyResetAt:    1793260800,
		WeeklyResetAt:   1793520000,
	})
	if r.ResetAt != "2026-10-29T08:00:00Z" {
		t.Errorf("日驱动 ResetAt: got %q, want 2026-10-29T08:00:00Z", r.ResetAt)
	}
	if r.ResetLabel != "日额度" {
		t.Errorf("日驱动 ResetLabel: got %q", r.ResetLabel)
	}

	// 周驱动 + 周有重置时间。
	r = f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(90),
		WeeklyRemaining: devinF(10),
		DailyResetAt:    1793260800,
		WeeklyResetAt:   1793520000,
	})
	if r.ResetAt != "2026-11-01T08:00:00Z" {
		t.Errorf("周驱动 ResetAt: got %q, want 2026-11-01T08:00:00Z", r.ResetAt)
	}
	if r.ResetLabel != "周额度" {
		t.Errorf("周驱动 ResetLabel: got %q", r.ResetLabel)
	}

	// 日驱动但日窗口没有重置时间(周有时间)⇒ 两者都留空,不借周的时间。
	r = f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(10),
		WeeklyRemaining: devinF(90),
		DailyResetAt:    0,
		WeeklyResetAt:   1793520000,
	})
	if r.ResetAt != "" || r.ResetLabel != "" {
		t.Errorf("驱动窗口无重置时间时应都留空, got ResetAt=%q ResetLabel=%q", r.ResetAt, r.ResetLabel)
	}
}

// ---------- AS-6 隐藏窗口 ----------

// TestDevinApplyPlanStatus_HideDaily 只隐藏日窗口 ⇒ 只看周。
func TestDevinApplyPlanStatus_HideDaily(t *testing.T) {
	f := NewDevinFetcher("t")
	r := f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(0), // 已用 100,若参与驱动会顶到 100
		WeeklyRemaining: devinF(82),
		DailyResetAt:    1793260800,
		WeeklyResetAt:   1793520000,
		HideDaily:       true,
	})
	if r.Error != "" {
		t.Fatalf("unexpected error: %s", r.Error)
	}
	if r.Percent != 18 {
		t.Errorf("隐藏日窗口后 Percent 应只由周决定: got %v, want 18", r.Percent)
	}
	if r.ResetLabel != "周额度" {
		t.Errorf("ResetLabel: got %q, want 周额度", r.ResetLabel)
	}
	if r.Remaining != "周 18.0% 已用" {
		t.Errorf("Remaining: got %q, want 只有周一段", r.Remaining)
	}
}

// TestDevinApplyPlanStatus_HideWeekly 对称:只隐藏周窗口 ⇒ 只看日。
func TestDevinApplyPlanStatus_HideWeekly(t *testing.T) {
	f := NewDevinFetcher("t")
	r := f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(82),
		WeeklyRemaining: devinF(0),
		DailyResetAt:    1793260800,
		WeeklyResetAt:   1793520000,
		HideWeekly:      true,
	})
	if r.Error != "" {
		t.Fatalf("unexpected error: %s", r.Error)
	}
	if r.Percent != 18 {
		t.Errorf("Percent: got %v, want 18", r.Percent)
	}
	if r.ResetLabel != "日额度" {
		t.Errorf("ResetLabel: got %q, want 日额度", r.ResetLabel)
	}
	if r.Remaining != "日 18.0% 已用" {
		t.Errorf("Remaining: got %q, want 只有日一段", r.Remaining)
	}
}

// TestDevinApplyPlanStatus_HideBoth 两个窗口都隐藏 ⇒ 报错且不带百分比与倒计时。
func TestDevinApplyPlanStatus_HideBoth(t *testing.T) {
	f := NewDevinFetcher("t")
	r := f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(20),
		WeeklyRemaining: devinF(30),
		DailyResetAt:    1793260800,
		WeeklyResetAt:   1793520000,
		HideDaily:       true,
		HideWeekly:      true,
	})
	if !strings.Contains(r.Error, "不提供可读的额度窗口") {
		t.Fatalf("Error 应含「不提供可读的额度窗口」, got %q", r.Error)
	}
	if r.Percent != 0 {
		t.Errorf("Percent: got %v, want 0", r.Percent)
	}
	if r.ResetAt != "" {
		t.Errorf("ResetAt: got %q, want 空", r.ResetAt)
	}
	devinCheckBase(t, r)
}

// ---------- 单侧缺百分比(票面第 6 条 r1 补定) ----------

// TestDevinApplyPlanStatus_MissingPercentIsInvisible 本次响应没给某窗口剩余百分比 ⇒ 该窗口不可见:
// 不当作「0% 剩余」渲染成 100% 已用、不 panic、Remaining 只剩另一侧一段、
// 驱动窗口落在给了百分比的那一侧。两侧都缺由解析层报结构错误(见 StructuralErrors)。
func TestDevinApplyPlanStatus_MissingPercentIsInvisible(t *testing.T) {
	f := NewDevinFetcher("t")

	// 只给了周(日窗口 DailyRemaining == nil)。日的重置时间故意给非零:
	// 若把缺百分比当成可见,它会以"0% 剩余 ⇒ 100% 已用"顶成驱动窗口并抢走倒计时。
	r := f.applyPlanStatus(&devinPlanStatus{
		WeeklyRemaining: devinF(30),
		DailyResetAt:    1793260800, // 2026-10-29T08:00:00Z
		WeeklyResetAt:   1793520000, // 2026-11-01T08:00:00Z
	})
	if r.Error != "" {
		t.Fatalf("单侧缺百分比不该整体报错, got %q", r.Error)
	}
	if r.Percent != 70 || r.Used != 70 || r.Total != 100 {
		t.Errorf("只给周(剩余 30): Percent/Used/Total = %v/%v/%v, want 70/70/100",
			r.Percent, r.Used, r.Total)
	}
	if r.ResetLabel != "周额度" {
		t.Errorf("驱动窗口应落在有百分比的一侧: ResetLabel got %q, want 周额度", r.ResetLabel)
	}
	if r.ResetAt != "2026-11-01T08:00:00Z" {
		t.Errorf("ResetAt 应取周的重置时间: got %q, want 2026-11-01T08:00:00Z", r.ResetAt)
	}
	if r.Remaining != "周 70.0% 已用" {
		t.Errorf("Remaining 只应有周一段: got %q, want %q", r.Remaining, "周 70.0% 已用")
	}

	// 对称:只给了日。
	r = f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining: devinF(30),
		DailyResetAt:   1793260800, // 2026-10-29T08:00:00Z
		WeeklyResetAt:  1793520000, // 2026-11-01T08:00:00Z
	})
	if r.Error != "" {
		t.Fatalf("单侧缺百分比不该整体报错, got %q", r.Error)
	}
	if r.Percent != 70 || r.Used != 70 {
		t.Errorf("只给日(剩余 30): Percent/Used = %v/%v, want 70/70", r.Percent, r.Used)
	}
	if r.ResetLabel != "日额度" {
		t.Errorf("ResetLabel: got %q, want 日额度", r.ResetLabel)
	}
	if r.ResetAt != "2026-10-29T08:00:00Z" {
		t.Errorf("ResetAt 应取日的重置时间: got %q, want 2026-10-29T08:00:00Z", r.ResetAt)
	}
	if r.Remaining != "日 70.0% 已用" {
		t.Errorf("Remaining 只应有日一段: got %q, want %q", r.Remaining, "日 70.0% 已用")
	}
}

// ---------- AS-7 Remaining 文案 ----------

// TestDevinApplyPlanStatus_RemainingText 按已用降序、一位小数、超额追加、不含套餐名。
func TestDevinApplyPlanStatus_RemainingText(t *testing.T) {
	f := NewDevinFetcher("t")

	// 日 20.0 已用 / 周 18.0 已用,无超额。
	r := f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(80),
		WeeklyRemaining: devinF(82),
	})
	if r.Remaining != "日 20.0% 已用 · 周 18.0% 已用" {
		t.Errorf("Remaining: got %q", r.Remaining)
	}

	// 日 18.0 / 周 20.0 ⇒ 周在前。
	r = f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(82),
		WeeklyRemaining: devinF(80),
	})
	if r.Remaining != "周 20.0% 已用 · 日 18.0% 已用" {
		t.Errorf("降序 Remaining: got %q", r.Remaining)
	}

	// 有超额余额 ⇒ 末尾追加一段金额(micros/1e6,两位小数)。
	r = f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(80),
		WeeklyRemaining: devinF(82),
		OverageMicros:   devinI64(70870000),
	})
	if r.Remaining != "日 20.0% 已用 · 周 18.0% 已用 · 超额金额 $70.87" {
		t.Errorf("带超额 Remaining: got %q", r.Remaining)
	}
	if strings.Contains(r.Remaining, "超额余额") {
		t.Errorf("文案应为「超额金额」, got %q", r.Remaining)
	}

	// 超额为 0 或负数 ⇒ 不追加。
	for _, over := range []int64{0, -1} {
		r = f.applyPlanStatus(&devinPlanStatus{
			DailyRemaining:  devinF(80),
			WeeklyRemaining: devinF(82),
			OverageMicros:   devinI64(over),
		})
		if strings.Contains(r.Remaining, "超额") {
			t.Errorf("overage=%d 不应追加超额段, got %q", over, r.Remaining)
		}
	}

	// 文案不含套餐名(planName 只解析不展示)。
	r = f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(80),
		WeeklyRemaining: devinF(82),
		PlanName:        "Pro",
	})
	if strings.Contains(r.Remaining, "Pro") || strings.Contains(r.Remaining, "Free") {
		t.Errorf("Remaining 不得出现套餐名, got %q", r.Remaining)
	}
}

// TestDevinApplyPlanStatus_RemainingText_TieDailyFirst 两侧已用并列时 Remaining 的日段在前
// (票面第 6 条「按已用降序(并列日在前)」)。排序一旦把并列当成可交换,两段顺序会随排序实现漂移。
func TestDevinApplyPlanStatus_RemainingText_TieDailyFirst(t *testing.T) {
	f := NewDevinFetcher("t")

	r := f.applyPlanStatus(&devinPlanStatus{
		DailyRemaining:  devinF(20), // 已用 80
		WeeklyRemaining: devinF(20), // 已用 80,与日并列
	})
	if r.Error != "" {
		t.Fatalf("unexpected error: %s", r.Error)
	}
	if r.Remaining != "日 80.0% 已用 · 周 80.0% 已用" {
		t.Errorf("并列时 Remaining 日段必须在前: got %q, want %q",
			r.Remaining, "日 80.0% 已用 · 周 80.0% 已用")
	}
}

// ---------- AS-8 空 token ----------

// TestDevinFetcher_EmptyToken_ExactErrorAndNoRequest token 为空 ⇒ 文案逐字相等,且零请求。
func TestDevinFetcher_EmptyToken_ExactErrorAndNoRequest(t *testing.T) {
	var hit int32
	f, _ := devinTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hit, 1)
		w.Write([]byte(devinOKBody))
	})
	f.apiKey = ""

	result := f.Fetch()
	if result.Error != "未配置 Devin token（登录 Devin Desktop 后留空即可，或粘贴 token）" {
		t.Errorf("Error 文案必须逐字相等:\n got %q\nwant %q", result.Error,
			"未配置 Devin token（登录 Devin Desktop 后留空即可，或粘贴 token）")
	}
	if hit != 0 {
		t.Errorf("token 为空时不得发请求, 实际 %d 次", hit)
	}
	devinCheckBase(t, result)
}

// TestDevinFetcher_TokenIsTrimmed token 前后空白不该被当成"没配置",也不该原样带进请求。
func TestDevinFetcher_TokenIsTrimmed(t *testing.T) {
	var gotKey string
	f, _ := devinTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var payload struct {
			Metadata map[string]string `json:"metadata"`
		}
		json.Unmarshal(raw, &payload)
		gotKey = payload.Metadata["api_key"]
		w.Write([]byte(devinOKBody))
	})
	f.apiKey = "  dv-test-token  "
	if r := f.Fetch(); r.Error != "" {
		t.Fatalf("unexpected error: %s", r.Error)
	}
	if gotKey != "dv-test-token" {
		t.Errorf("api_key 应被 TrimSpace: got %q", gotKey)
	}
}

// ---------- AS-9 错误映射 ----------

// TestDevinFetcher_ErrorMapping 401/403 → 失效文案;500 → 含 HTTP 500;非 JSON → 含「无法解析额度数据」。
func TestDevinFetcher_ErrorMapping(t *testing.T) {
	const expired = "Devin token 已失效，请重新登录 Devin Desktop 或更新 token"

	t.Run("403", func(t *testing.T) {
		f, _ := devinTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"code":"permission_denied","message":"Invalid token"}`))
		})
		r := f.Fetch()
		if r.Error != expired {
			t.Errorf("403 Error: got %q, want %q", r.Error, expired)
		}
		devinCheckBase(t, r)
	})

	t.Run("401", func(t *testing.T) {
		f, _ := devinTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
		r := f.Fetch()
		if r.Error != expired {
			t.Errorf("401 Error: got %q, want %q", r.Error, expired)
		}
		devinCheckBase(t, r)
	})

	t.Run("500", func(t *testing.T) {
		f, _ := devinTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		r := f.Fetch()
		if !strings.Contains(r.Error, "HTTP 500") {
			t.Errorf("500 Error 应含 HTTP 500: got %q", r.Error)
		}
		devinCheckBase(t, r)
	})

	t.Run("200 但非 JSON", func(t *testing.T) {
		f, _ := devinTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>oops</html>`))
		})
		r := f.Fetch()
		if !strings.Contains(r.Error, "无法解析额度数据") {
			t.Errorf("非 JSON Error 应含「无法解析额度数据」: got %q", r.Error)
		}
		devinCheckBase(t, r)
	})

	t.Run("200 但缺 planStatus", func(t *testing.T) {
		f, _ := devinTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"userStatus":{}}`))
		})
		r := f.Fetch()
		if !strings.Contains(r.Error, "无法解析额度数据") {
			t.Errorf("缺 planStatus Error: got %q", r.Error)
		}
		devinCheckBase(t, r)
	})
}

// ---------- 端到端(Fetch 成功路径) ----------

// TestDevinFetcher_Fetch_OK 一次完整成功路径:80% 已用、日驱动、倒计时与 Remaining 都齐。
func TestDevinFetcher_Fetch_OK(t *testing.T) {
	f, _ := devinTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(devinOKBody))
	})
	r := f.Fetch()
	if r.Error != "" {
		t.Fatalf("unexpected error: %s", r.Error)
	}
	if r.Percent != 80 || r.Used != 80 || r.Total != 100 {
		t.Errorf("Percent/Used/Total = %v/%v/%v, want 80/80/100", r.Percent, r.Used, r.Total)
	}
	if r.ResetAt != "2026-10-29T08:00:00Z" || r.ResetLabel != "日额度" {
		t.Errorf("ResetAt/ResetLabel = %q/%q", r.ResetAt, r.ResetLabel)
	}
	if r.Remaining != "日 80.0% 已用 · 周 30.0% 已用 · 超额金额 $70.87" {
		t.Errorf("Remaining: got %q", r.Remaining)
	}
	devinCheckBase(t, r)
}

// TestDevinFetcher_DefaultServerURL 不注入 serverURL 时走默认端点常量,
// 且构造函数不得把默认值提前写进可注入字段(票 02 靠这个字段让本机 apiServerUrl 生效)。
func TestDevinFetcher_DefaultServerURL(t *testing.T) {
	f := NewDevinFetcher("dv-test-token")
	if f.serverURL != "" {
		t.Errorf("构造函数的 serverURL 必须留空(默认值只由 devinStatusURL 兜底), got %q", f.serverURL)
	}
	want := "https://server.codeium.com/exa.seat_management_pb.SeatManagementService/GetUserStatus"
	if got := devinStatusURL(f.serverURL); got != want {
		t.Errorf("默认端点: got %q, want %q", got, want)
	}
}

// ---------- 小工具 ----------

func devinF(v float64) *float64 { return &v }
func devinI64(v int64) *int64   { return &v }

// 编译期断言:DevinFetcher 必须满足 Fetcher 接口(注册表 Build 依赖它)。
var _ Fetcher = (*DevinFetcher)(nil)
