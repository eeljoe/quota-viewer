package fetcher

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFactoryDroid_EmptyKey_ReturnsError 验证空 API Key(且无 .env)返回错误。
func TestFactoryDroid_EmptyKey_ReturnsError(t *testing.T) {
	f := NewFactoryDroidFetcher("")
	f.readEnvFile = false // 禁用 .env 读取,保证离线可测
	result := f.Fetch()
	if result.Error == "" {
		t.Error("expected error for empty api key")
	}
	if result.Platform != "Factory Droid" {
		t.Errorf("expected platform 'Factory Droid', got '%s'", result.Platform)
	}
}

// TestFactoryDroid_EmptyKey_ReadsEnvFile 验证空 key 时自动读取 ~/.factory/.env 的
// FACTORY_API_KEY(支持 export 前缀与行尾注释,Droid CLI 同款格式)。
func TestFactoryDroid_EmptyKey_ReadsEnvFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("USERPROFILE", tmp)
	envDir := filepath.Join(tmp, ".factory")
	if err := os.MkdirAll(envDir, 0755); err != nil {
		t.Fatal(err)
	}
	envContent := "export FACTORY_API_KEY='fk_secret' # Droid key\n"
	if err := os.WriteFile(filepath.Join(envDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(validFactoryLimitsJSON))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotAuth != "Bearer fk_secret" {
		t.Errorf("expected Bearer fk_secret, got '%s'", gotAuth)
	}
}

// TestFactoryDroid_TokenRateLimits_FormatsResult 验证新账单模型(token-rate limits)解析:
// 5h 窗口驱动主展示,Percent 取全部窗口最紧张的,预付余额写入 Remaining。
func TestFactoryDroid_TokenRateLimits_FormatsResult(t *testing.T) {
	var gotClientHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClientHeader = r.Header.Get("x-factory-client")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(validFactoryLimitsJSON))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotClientHeader != "web-app" {
		t.Errorf("expected x-factory-client 'web-app', got '%s'", gotClientHeader)
	}
	// 5h 20% / 周 45% / 月 12% → Percent = 45(最紧张窗口)
	if result.Percent != 45 {
		t.Errorf("expected Percent=45, got %f", result.Percent)
	}
	if result.Total != 100 || result.Used != 20 {
		t.Errorf("expected Used=20/Total=100 (5h 窗口驱动主展示), got %f/%f", result.Used, result.Total)
	}
	// 展示与 Kimi/Ollama 对齐:常驻 5小时/周;月度非最紧张窗口不显示;Extra 有余款才显示
	if result.Remaining != "5小时 20.0% 已用 · 周 45.0% 已用 · Extra $3.20" {
		t.Errorf("unexpected Remaining: '%s'", result.Remaining)
	}
	// ResetAt 来自 5h 窗口的 secondsRemaining(3600s)
	if reset, err := time.Parse(time.RFC3339, result.ResetAt); err != nil {
		t.Errorf("expected parseable ResetAt, got '%s'", result.ResetAt)
	} else if diff := time.Until(reset); diff < 55*time.Minute || diff > 65*time.Minute {
		t.Errorf("expected ResetAt ~1h from now, got '%s'", result.ResetAt)
	}
}

// TestFactoryDroid_WeeklyExhausted_DrivesPercent 验证周/月窗口耗尽时球色必须告警:
// Percent 取全部窗口中更紧张的一个(5h 0% / 周 100% 不应恒绿)。
func TestFactoryDroid_WeeklyExhausted_DrivesPercent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "usesTokenRateLimitsBilling": true,
  "limits": {"standard": {"fiveHour": {"usedPercent": 0}, "weekly": {"usedPercent": 100}}}
}`))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 100 {
		t.Errorf("expected Percent=100 (weekly exhausted drives ball color), got %f", result.Percent)
	}
}

// TestFactoryDroid_CoreWindow_Alerts 验证 Core 池(独立计费窗口)耗尽同样驱动告警。
func TestFactoryDroid_CoreWindow_Alerts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "usesTokenRateLimitsBilling": true,
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 10}},
    "core": {"monthly": {"usedPercent": 100}}
  }
}`))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 100 {
		t.Errorf("expected Percent=100 (core exhausted drives ball color), got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "Core") {
		t.Errorf("expected Core info in Remaining, got '%s'", result.Remaining)
	}
}

// TestFactoryDroid_MonthlyDrives_Shows 验证月度成为最紧张窗口时必须展示并驱动告警
// (长窗口耗尽漏报是 9/19、9/22 两次修复的同类坑,月度不能只算不显示)。
func TestFactoryDroid_MonthlyDrives_Shows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "usesTokenRateLimitsBilling": true,
  "limits": {"standard": {"fiveHour": {"usedPercent": 10}, "weekly": {"usedPercent": 20}, "monthly": {"usedPercent": 90}}}
}`))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 90 {
		t.Errorf("expected Percent=90 (monthly drives ball color), got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "月 90.0% 已用") {
		t.Errorf("expected monthly window surfaced when driving, got '%s'", result.Remaining)
	}
}

// TestFactoryDroid_CoreActive_Shows 验证 Core 仓(Standard 限流后的接管计费)
// 一有消耗就展示——即使其百分比低于 Standard 窗口(回退场景,用户常驻开源模型)。
func TestFactoryDroid_CoreActive_Shows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "usesTokenRateLimitsBilling": true,
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 40}},
    "core": {"fiveHour": {"usedPercent": 5}}
  }
}`))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if result.Percent != 40 {
		t.Errorf("expected Percent=40 (standard still drives), got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "Core 5.0% 已用") {
		t.Errorf("expected active Core tank surfaced, got '%s'", result.Remaining)
	}
}

// TestFactoryDroid_Legacy_Fallback 验证旧账单模型兜底:billing/limits 未启用
// token-rate 计费时,回退 /api/organization/subscription/usage。
func TestFactoryDroid_Legacy_Fallback(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/billing/limits":
			_, _ = w.Write([]byte(`{"usesTokenRateLimitsBilling": false}`))
		case "/api/organization/subscription/usage":
			gotQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{
  "usage": {
    "endDate": 1789344000000,
    "standard": {"userTokens": 250, "totalAllowance": 1000, "usedRatio": 0},
    "premium": {"userTokens": 50, "totalAllowance": 100, "usedRatio": 0.5}
  }
}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	if gotQuery == "" || !strings.Contains(gotQuery, "useCache=true") {
		t.Errorf("expected useCache=true on legacy route, got '%s'", gotQuery)
	}
	// Standard: ratio=0 但 userTokens=250/1000(脏 ratio 不可信)→ 25%;Premium 50% → Percent=50
	if result.Percent != 50 {
		t.Errorf("expected Percent=50, got %f", result.Percent)
	}
	if !strings.Contains(result.Remaining, "Standard 25.0%") || !strings.Contains(result.Remaining, "Premium 50.0%") {
		t.Errorf("unexpected Remaining: '%s'", result.Remaining)
	}
}

// TestFactoryDroid_401_ReturnsInvalidKey 验证 401 返回 Key 失效错误。
func TestFactoryDroid_401_ReturnsInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "无效") {
		t.Errorf("expected invalid key error, got '%s'", result.Error)
	}
}

// TestFactoryDroid_429_ReturnsRateLimited 验证 429 透出限流错误。
func TestFactoryDroid_429_ReturnsRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "429") {
		t.Errorf("expected rate limit error, got '%s'", result.Error)
	}
}

// TestFactoryDroid_BadJSON_ReturnsError 验证两路响应都异常时明确报错。
func TestFactoryDroid_BadJSON_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	f := NewFactoryDroidFetcher("fk_test")
	f.baseURL = server.URL
	result := f.Fetch()
	if !strings.Contains(result.Error, "无法解析") {
		t.Errorf("expected parse error, got '%s'", result.Error)
	}
}

// realFactoryLimitsJSON 是 2026-10-13 11:54 真机 /api/billing/limits 的逐字响应
// (事故现场:Standard 周 100% 顶格,超额已在走 Core)。
// 注意 core.fiveHour.windowEnd 是过期值、standard.fiveHour.secondsRemaining 是 null
// ⇒ 重置时间只能由 secondsRemaining 推算,不能信 windowEnd。
const realFactoryLimitsJSON = `{"usesTokenRateLimitsBilling":true,
 "limits":{"standard":{"fiveHour":{"usedPercent":0,"windowEnd":null,"secondsRemaining":null},
                       "weekly":{"usedPercent":100,"windowEnd":"2026-10-14T16:41:58.478Z","secondsRemaining":132577},
                       "monthly":{"usedPercent":82,"windowEnd":"2026-10-30T08:36:58.540Z","secondsRemaining":1485877}},
           "core":{"fiveHour":{"usedPercent":100,"windowEnd":"2026-10-13T06:05:15.349Z","secondsRemaining":7974},
                   "weekly":{"usedPercent":53,"windowEnd":"2026-10-19T20:04:35.466Z","secondsRemaining":576734},
                   "monthly":{"usedPercent":40,"windowEnd":"2026-11-11T20:04:35.466Z","secondsRemaining":2563934}}},
 "overagePreference":"droidCore","canManageOverage":true,"extraUsageBalanceCents":0,"extraUsageAllowed":true}`

// TestFactoryDroid_CoreTakeover_SwitchesDisplay 验证 Standard 顶格且超额走 Core 时,
// 展示主体(Percent/Used/ResetAt/Remaining)切到 Core 池;不满足接管条件时与现状逐字一致。
func TestFactoryDroid_CoreTakeover_SwitchesDisplay(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		percent    float64
		used       float64
		resetSecs  float64 // 期望的剩余秒数;<0 表示 ResetAt 必须是空串(驱动窗口无剩余秒数)
		remaining  string  // 期望全串(空串表示只校验前缀/tail)
		hasPrefix  string
		hasTail    string
		coreCounts int // >=0 时校验 "Core" 出现次数;-1 表示不校验
	}{
		{
			// AC-1:真机夹具 → 接管,重置时间取 Core 5h 的 7974s
			name: "AC1_真机夹具接管", body: realFactoryLimitsJSON,
			percent: 100, used: 100, resetSecs: 7974,
			remaining:  "Core 100.0% 已用（Standard 已限流） · 周 100.0% 已用",
			coreCounts: 1,
		},
		{
			// AC-2:本票核心判据——球面必须反映出 Core 还有 80 的量
			name: "AC2_Core还有80的量", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "droidCore",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 0, "secondsRemaining": 111}, "weekly": {"usedPercent": 100}},
    "core": {"fiveHour": {"usedPercent": 20, "secondsRemaining": 600}, "weekly": {"usedPercent": 12}}
  }
}`,
			percent: 20, used: 20, resetSecs: 600,
			remaining:  "Core 20.0% 已用（Standard 已限流） · 周 100.0% 已用",
			coreCounts: -1,
		},
		{
			// AC-3:Standard 未顶格 → 不接管,Core 只作为附带说明
			name: "AC3_Standard未顶格不接管", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "droidCore",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 50, "secondsRemaining": 300}, "weekly": {"usedPercent": 50}},
    "core": {"fiveHour": {"usedPercent": 80}}
  }
}`,
			percent: 80, used: 50, resetSecs: 300,
			hasTail:    "Core 80.0% 已用",
			coreCounts: -1,
		},
		{
			// AC-4:Core 无数据(全 null)→ 不接管,Percent 不许掉成 0
			name: "AC4_Core全null不接管", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "droidCore",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 0, "secondsRemaining": 300}, "weekly": {"usedPercent": 100}},
    "core": {"fiveHour": {"usedPercent": null, "secondsRemaining": null}, "weekly": {"usedPercent": null}}
  }
}`,
			percent: 100, used: 0, resetSecs: 300,
			coreCounts: 0, hasTail: "周 100.0% 已用",
		},
		{
			// AC-4':Core 整体缺失 → 同样不接管
			name: "AC4b_Core缺失不接管", body: `{
  "usesTokenRateLimitsBilling": true,
  "limits": {"standard": {"fiveHour": {"usedPercent": 0}, "weekly": {"usedPercent": 100}}}
}`,
			percent: 100, used: 0, resetSecs: -1,
			remaining: "5小时 0.0% 已用 · 周 100.0% 已用", coreCounts: 0,
		},
		{
			// AC-5:顶格判据是 >= 100,99.9 不算顶格
			name: "AC5_99.9不算顶格", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "droidCore",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 30}, "weekly": {"usedPercent": 99.9}},
    "core": {"fiveHour": {"usedPercent": 60}}
  }
}`,
			percent: 99.9, used: 30, resetSecs: -1,
			remaining:  "5小时 30.0% 已用 · 周 99.9% 已用 · Core 60.0% 已用",
			coreCounts: -1,
		},
		{
			// AC-6:超额没走 Core(走 extraUsage)→ 不许把主体切过去
			name: "AC6_extraUsage不接管", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "extraUsage",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 0}, "weekly": {"usedPercent": 100}},
    "core": {"fiveHour": {"usedPercent": 60}}
  }
}`,
			percent: 100, used: 0, resetSecs: -1,
			remaining:  "5小时 0.0% 已用 · 周 100.0% 已用 · Core 60.0% 已用",
			coreCounts: -1,
		},
		{
			// AC-7:字段缺失(老上游)按空串处理 → 接管
			name: "AC7_字段缺失仍接管", body: `{
  "usesTokenRateLimitsBilling": true,
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 0, "secondsRemaining": 900}, "weekly": {"usedPercent": 100}},
    "core": {"fiveHour": {"usedPercent": 60, "secondsRemaining": 1800}}
  }
}`,
			percent: 60, used: 60, resetSecs: 1800,
			hasPrefix:  "Core 60.0% 已用（Standard 已限流）",
			coreCounts: -1,
		},
		{
			// AC-8:Core 也顶住了 → 接管,球熄灭是对的
			name: "AC8_Core也顶格", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "droidCore",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 0, "secondsRemaining": 77}, "weekly": {"usedPercent": 100}},
    "core": {"fiveHour": {"usedPercent": 100, "secondsRemaining": 5000}}
  }
}`,
			percent: 100, used: 100, resetSecs: 5000,
			hasPrefix:  "Core 100.0% 已用（Standard 已限流）",
			coreCounts: -1,
		},
		{
			// 驱动窗口是 Core 周而非 Core 5h(tie-break 按 fiveHour→weekly→monthly;ResetAt 跟驱动窗口走)
			name: "Core周最紧张时接管", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "droidCore",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 0}, "weekly": {"usedPercent": 100}},
    "core": {"fiveHour": {"usedPercent": 10, "secondsRemaining": 111}, "weekly": {"usedPercent": 70, "secondsRemaining": 4321}, "monthly": {"usedPercent": 65}}
  }
}`,
			percent: 70, used: 70, resetSecs: 4321,
			hasPrefix:  "Core 70.0% 已用（Standard 已限流）",
			coreCounts: -1,
		},
		{
			// 驱动窗口的 secondsRemaining 为 null → ResetAt 留空串(不许用 windowEnd 顶)
			name: "接管但驱动窗口无秒数", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "droidCore",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 0, "secondsRemaining": 3600}, "weekly": {"usedPercent": 100}},
    "core": {"fiveHour": {"usedPercent": 45, "windowEnd": "2026-10-13T06:05:15.349Z", "secondsRemaining": null}}
  }
}`,
			percent: 45, used: 45, resetSecs: -1,
			hasPrefix:  "Core 45.0% 已用（Standard 已限流）",
			coreCounts: -1,
		},
		{
			// 平手取先者(a):Core 5h 与周同为 60 → 驱动窗口必须是 5h,
			// ResetAt 跟它走(≈now+111s),不许落到更靠后的 weekly(4321s)。
			name: "Core平手取先者_reset来自5h", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "droidCore",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 0}, "weekly": {"usedPercent": 100}},
    "core": {"fiveHour": {"usedPercent": 60, "secondsRemaining": 111}, "weekly": {"usedPercent": 60, "secondsRemaining": 4321}}
  }
}`,
			percent: 60, used: 60, resetSecs: 111,
			remaining:  "Core 60.0% 已用（Standard 已限流） · 周 100.0% 已用",
			coreCounts: 1,
		},
		{
			// 平手取先者(b):Core 5h(无秒数)与周同为 60 → 驱动窗口是 5h,
			// 它没有秒数 ⇒ ResetAt 必须空串,不许借用 weekly 的 4321s。
			name: "Core平手取先者_5h无秒数则空串", body: `{
  "usesTokenRateLimitsBilling": true,
  "overagePreference": "droidCore",
  "limits": {
    "standard": {"fiveHour": {"usedPercent": 0}, "weekly": {"usedPercent": 100}},
    "core": {"fiveHour": {"usedPercent": 60, "secondsRemaining": null}, "weekly": {"usedPercent": 60, "secondsRemaining": 4321}}
  }
}`,
			percent: 60, used: 60, resetSecs: -1,
			remaining:  "Core 60.0% 已用（Standard 已限流） · 周 100.0% 已用",
			coreCounts: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			f := NewFactoryDroidFetcher("fk_test")
			f.baseURL = server.URL
			t0 := time.Now()
			result := f.Fetch()
			t1 := time.Now()
			if result.Error != "" {
				t.Fatalf("unexpected error: %s", result.Error)
			}
			if result.Percent != tc.percent {
				t.Errorf("Percent = %f, want %f", result.Percent, tc.percent)
			}
			if result.Used != tc.used {
				t.Errorf("Used = %f, want %f", result.Used, tc.used)
			}
			if result.Total != 100 {
				t.Errorf("Total = %f, want 100", result.Total)
			}
			if tc.resetSecs < 0 {
				if result.ResetAt != "" {
					t.Errorf("ResetAt = '%s', want empty (驱动窗口无剩余秒数)", result.ResetAt)
				}
			} else {
				// 亚秒夹逼:ResetAt 只能落在 [t0+want-1s, t1+want] 内。下界那 1s 是
				// RFC3339 秒级截断的必然损耗,不是容差;上界不留松弛 ⇒ 秒数换算
				// 差 30s 这类变异(如把秒当毫秒用)照样被夹死。
				reset, err := time.Parse(time.RFC3339, result.ResetAt)
				if err != nil {
					t.Fatalf("ResetAt '%s' 不可解析: %v", result.ResetAt, err)
				}
				want := time.Duration(tc.resetSecs * float64(time.Second))
				if reset.Before(t0.Add(want-time.Second)) || reset.After(t1.Add(want)) {
					t.Errorf("ResetAt = '%s'(抓取时刻后 %.0fs), want 抓取时刻 + %.0fs 的亚秒夹逼(下界 -1s)",
						result.ResetAt, reset.Sub(t0).Seconds(), tc.resetSecs)
				}
			}
			if tc.remaining != "" && result.Remaining != tc.remaining {
				t.Errorf("Remaining = '%s', want '%s'", result.Remaining, tc.remaining)
			}
			if tc.hasPrefix != "" && !strings.HasPrefix(result.Remaining, tc.hasPrefix) {
				t.Errorf("Remaining = '%s', want prefix '%s'", result.Remaining, tc.hasPrefix)
			}
			if tc.hasTail != "" && !strings.HasSuffix(result.Remaining, tc.hasTail) {
				t.Errorf("Remaining = '%s', want suffix '%s'", result.Remaining, tc.hasTail)
			}
			if tc.coreCounts >= 0 {
				if got := strings.Count(result.Remaining, "Core"); got != tc.coreCounts {
					t.Errorf("Remaining 出现 Core %d 次, want %d: '%s'", got, tc.coreCounts, result.Remaining)
				}
			}
		})
	}
}

// validFactoryLimitsJSON 是真实 /api/billing/limits 返回结构的代表
// (token-rate limits 计费:5h/周/月窗口 + 预付 Extra Usage 余额)。
const validFactoryLimitsJSON = `{
  "usesTokenRateLimitsBilling": true,
  "limits": {
    "standard": {
      "fiveHour": {"usedPercent": 20, "secondsRemaining": 3600},
      "weekly": {"usedPercent": 45},
      "monthly": {"usedPercent": 12}
    }
  },
  "extraUsageBalanceCents": 320
}`
