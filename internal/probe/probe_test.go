package probe

import (
	"testing"
	"time"
)

// testNow 固定「本次抓取时刻」：2026-10-13 10:45:00 +08:00。
func testNow() time.Time {
	return time.Date(2026, 10, 13, 10, 45, 0, 0, time.FixedZone("CST", 8*3600))
}

// TestNormalizeResetAt_KimiSubsecond_ReformattedToUTCSeconds kimi 上游串带亚秒
// （kimi.go 写入形态），必须重排为 UTC 秒精度，不许原样透传。
func TestNormalizeResetAt_KimiSubsecond_ReformattedToUTCSeconds(t *testing.T) {
	got := normalizeResetAt("2026-10-20T08:00:00.123456Z", testNow())
	if got == nil {
		t.Fatal("kimi 亚秒串应解析出时刻，实际得到 null")
	}
	if *got != "2026-10-20T08:00:00Z" {
		t.Fatalf("亚秒串未重排为 UTC 秒精度：got %q, want %q", *got, "2026-10-20T08:00:00Z")
	}
}

// TestNormalizeResetAt_OffsetIsConvertedToUTC 带本地偏移的绝对时间换算到 UTC。
func TestNormalizeResetAt_OffsetIsConvertedToUTC(t *testing.T) {
	got := normalizeResetAt("2026-10-16T10:07:39+08:00", testNow())
	if got == nil {
		t.Fatal("带偏移的绝对时间应解析出时刻，实际得到 null")
	}
	if *got != "2026-10-16T02:07:39Z" {
		t.Fatalf("未换算到 UTC：got %q, want %q", *got, "2026-10-16T02:07:39Z")
	}
}

// TestNormalizeResetAt_NoZoneString_InterpretedAsCST xfyun 上游串无时区
// （xfyun.go:131 写入形态），按 +08:00 解释再转 UTC；按 UTC 解释会差 8 小时。
func TestNormalizeResetAt_NoZoneString_InterpretedAsCST(t *testing.T) {
	got := normalizeResetAt("2026-10-16 10:07:39", testNow())
	if got == nil {
		t.Fatal("无时区串应按 +08:00 解释出时刻，实际得到 null")
	}
	if *got != "2026-10-16T02:07:39Z" {
		t.Fatalf("无时区串未按 +08:00 解释：got %q, want %q", *got, "2026-10-16T02:07:39Z")
	}
}

// TestNormalizeResetAt_EmptyOrUnparsable_Null mimo / deepseek 不设 ResetAt（空串）。
func TestNormalizeResetAt_EmptyOrUnparsable_Null(t *testing.T) {
	cases := []struct{ name, raw string }{
		{"空串", ""},
		{"无法解析", "下周三下午"},
	}
	for _, c := range cases {
		if got := normalizeResetAt(c.raw, testNow()); got != nil {
			t.Errorf("%s：want null, got %q", c.name, *got)
		}
	}
}

// TestNormalizeResetAt_NotAfterNow_Null opencode-go 的 resetSec=0 会产出「现在」，
// 非严格未来（!parsed.After(now)）一律 null。
func TestNormalizeResetAt_NotAfterNow_Null(t *testing.T) {
	now := testNow()
	cases := []struct{ name, raw string }{
		{"正好是现在", now.Format(time.RFC3339)},
		{"已经过去", "2026-10-13T02:00:00Z"},
	}
	for _, c := range cases {
		if got := normalizeResetAt(c.raw, now); got != nil {
			t.Errorf("%s：want null, got %q", c.name, *got)
		}
	}
}