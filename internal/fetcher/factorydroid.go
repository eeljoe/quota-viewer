package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FactoryDroidFetcher 通过 Factory 私有 API(与官方 web 端同款路由)获取 Droid 订阅额度。
// 端点在 https://api.factory.ai 下,未写入公开文档,路由与官方 web bundle 同源
// (与 CodexBar 的 FactoryStatusProbe、token-monitor 的 factory limits provider 一致):
//   - GET /api/billing/limits                  → token-rate limits 模型:5h/周/月窗口百分比 + 预付余额
//   - GET /api/organization/subscription/usage → 旧账单模型兜底:Standard/Premium 用量
//
// 认证: Authorization: Bearer <API Key>。Key 在 app.factory.ai/settings/api-keys 生成
// (fk- 前缀);留空时自动读 ~/.factory/.env 的 FACTORY_API_KEY(Droid CLI 同款)。
// 备注: 私有路由,结构可能随官方升级变化;失效时对照官方 web bundle 更新本文件。
type FactoryDroidFetcher struct {
	apiKey      string
	baseURL     string // 可重写,用于测试
	readEnvFile bool   // apiKey 为空时是否尝试读取 ~/.factory/.env
	envFile     string // 覆盖默认 ~/.factory/.env(测试注入)
}

// NewFactoryDroidFetcher 创建一个新的 FactoryDroidFetcher。
// apiKey 为空时,若本机存在 ~/.factory/.env(含 FACTORY_API_KEY)自动读取。
func NewFactoryDroidFetcher(apiKey string) *FactoryDroidFetcher {
	home, _ := os.UserHomeDir()
	return &FactoryDroidFetcher{
		apiKey:      strings.TrimSpace(apiKey),
		baseURL:     "https://api.factory.ai",
		readEnvFile: true,
		envFile:     filepath.Join(home, ".factory", ".env"),
	}
}

// 以下响应结构对应官方 web 端的私有路由(字段名取自 web bundle,经 token-monitor 交叉验证)。
type factoryWindow struct {
	UsedPercent      *float64 `json:"usedPercent"`
	SecondsRemaining *float64 `json:"secondsRemaining"`
}

type factoryPool struct {
	FiveHour *factoryWindow `json:"fiveHour"`
	Weekly   *factoryWindow `json:"weekly"`
	Monthly  *factoryWindow `json:"monthly"`
}

type factoryLimitsResp struct {
	UsesTokenRateLimitsBilling bool `json:"usesTokenRateLimitsBilling"`
	Limits                     *struct {
		Standard *factoryPool `json:"standard"`
		Core     *factoryPool `json:"core"`
	} `json:"limits"`
	ExtraUsageBalanceCents *float64 `json:"extraUsageBalanceCents"`
	// OveragePreference 是上游给出的"超额走哪管血"显式信号("droidCore" = 走 Core 池)。
	OveragePreference string `json:"overagePreference"`
}

type factoryLegacyBucket struct {
	UserTokens     *float64 `json:"userTokens"`
	TotalAllowance *float64 `json:"totalAllowance"`
	UsedRatio      *float64 `json:"usedRatio"`
}

type factoryLegacyResp struct {
	Usage *struct {
		EndDate  int64                `json:"endDate"` // epoch 毫秒,账单期结束
		Standard *factoryLegacyBucket `json:"standard"`
		Premium  *factoryLegacyBucket `json:"premium"`
	} `json:"usage"`
}

// Fetch 拉取并解析 Factory Droid 额度。优先 token-rate limits 模型(5h/周/月窗口,
// 与官方 Usage 页一致),未启用该计费模式时回退旧账单模型的 Standard/Premium 用量。
// Percent 取全部窗口中更紧张的一个(窗口告警契约,与 Kimi/Ollama/Command Code 一致);
// Used/Total 以 5 小时主窗口为准(Core 接管时改跟 Core 驱动窗口);
// ResetAt/ResetLabel 跟 Remaining 排头那一段的窗口走(见 parseTokenRateLimits)。
func (f *FactoryDroidFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "Factory Droid",
		Kind:        KindUsage,
		LastUpdated: time.Now(),
	}

	key := f.apiKey
	if key == "" && f.readEnvFile {
		key = readFactoryEnvKey(f.envFile)
	}
	if key == "" {
		result.Error = "未配置 Factory API Key(在 app.factory.ai/settings/api-keys 生成填入,或写入 ~/.factory/.env)"
		return result
	}

	client := &http.Client{Timeout: 10 * time.Second}
	body, err := f.getJSON(client, key, "/api/billing/limits")
	if err != nil {
		result.Error = err.Error()
		return result
	}
	var limits factoryLimitsResp
	if err := json.Unmarshal(body, &limits); err != nil {
		limits.UsesTokenRateLimitsBilling = false // 结构异常 → 走旧模型兜底
	}

	if limits.UsesTokenRateLimitsBilling && limits.Limits != nil && limits.Limits.Standard != nil {
		return f.parseTokenRateLimits(result, &limits)
	}

	// 旧账单模型兜底(Standard/Premium 订阅用量)。
	legacyBody, err := f.getJSON(client, key, "/api/organization/subscription/usage?useCache=true")
	if err != nil {
		result.Error = err.Error()
		return result
	}
	var legacy factoryLegacyResp
	if err := json.Unmarshal(legacyBody, &legacy); err != nil || legacy.Usage == nil {
		result.Error = "无法解析额度数据(响应结构可能已变化)"
		return result
	}
	return f.parseLegacyUsage(result, &legacy)
}

// factorySegment 是 Remaining 里的一段展示文本:带紧张度(已用百分比)、所属窗口
// (倒计时取自排头那段)与两个标志位(pinned = 接管主体固定首位,extra = 无百分比的尾部)。
type factorySegment struct {
	text    string
	percent float64
	pinned  bool
	extra   bool
	window  *factoryWindow
	label   string
}

// parseTokenRateLimits 解析新账单模型。窗口只有百分比、无绝对值,沿用 Ollama 先例:
// Total=100、Used=5h 百分比;Percent 取全部窗口(含 Core 池)最紧张的一个。
// Remaining 的段按**紧张度降序**排列,最紧张的窗口打头;月度/Core 仍是"成为最紧张者
// 才出现"——长窗口耗尽不能只算不显示,但平时不做展示噪音。之所以按紧张度排序而不是
// 固定顺序:头部那行只有 210px 宽,最紧张的数字排在第三段会被截掉,球红了却看不到原因
// (2026-10-19 现场)。ResetAt / ResetLabel 同样跟排头那一段的窗口走。
//
// Core 接管:Standard 池顶格(>=100%)且上游 overagePreference 表明超额走 Core 时,
// 展示主体(Percent/Used/ResetAt/Remaining)整体切到 Core 池,且 Core 段固定排首位——
// 否则 Standard 的 100% 会把球面钉死成"用满",用户看不出 Core(第二管血)还剩多少量
// (2026-10-13 误判事故)。
func (f *FactoryDroidFetcher) parseTokenRateLimits(result QuotaResult, limits *factoryLimitsResp) QuotaResult {
	std := limits.Limits.Standard
	fiveHour := std.FiveHour

	poolMax := func(p *factoryPool) float64 {
		if p == nil {
			return 0
		}
		m := 0.0
		for _, w := range []*factoryWindow{p.FiveHour, p.Weekly, p.Monthly} {
			if w != nil && w.UsedPercent != nil && clampPercent(*w.UsedPercent) > m {
				m = clampPercent(*w.UsedPercent)
			}
		}
		return m
	}
	stdMax := poolMax(std)
	// Core 池是独立计费窗口,仅在确实有数据时纳入告警与展示。
	coreHasData := false
	core := limits.Limits.Core
	if core != nil {
		for _, w := range []*factoryWindow{core.FiveHour, core.Weekly, core.Monthly} {
			if w != nil && w.UsedPercent != nil {
				coreHasData = true
				break
			}
		}
	}
	coreMax := 0.0
	if coreHasData {
		coreMax = poolMax(core)
	}
	// 接管 = Standard 顶格(含等于 100) + Core 有数据 + 超额确实走 Core
	// (overagePreference 为空视为上游未表态,沿用"会走 Core"的老行为)。
	takesOver := stdMax >= 100 && coreHasData &&
		(limits.OveragePreference == "" || limits.OveragePreference == "droidCore")

	if fiveHour == nil || fiveHour.UsedPercent == nil {
		result.Error = "未找到 5 小时窗口额度,响应结构可能已变化"
		return result
	}

	result.Total = 100
	result.Used = clampPercent(*fiveHour.UsedPercent)
	result.Percent = stdMax

	// 展示段 = (紧张度, 所属窗口, 文本)。顺序在末尾统一按紧张度重排:
	// 头部只有 210px 宽,最紧张的窗口排第三段会被截掉 —— 球因它变红,用户却看不到数字。
	var segs []factorySegment
	if takesOver {
		// 接管时主体就是 Core:驱动窗口与重置时间都跟 Core 最紧张的窗口走。
		driver, driverName := factoryDrivingWindow(core)
		result.Percent = coreMax
		if driver != nil && driver.UsedPercent != nil {
			result.Used = clampPercent(*driver.UsedPercent)
		}
		segs = append(segs, factorySegment{
			text:    fmt.Sprintf("Core %.1f%% 已用（Standard 已限流）", coreMax),
			percent: coreMax,
			pinned:  true,
			window:  driver,
			label:   factoryResetLabel("Core ", driverName),
		})
	} else {
		segs = append(segs, factorySegment{
			text:    fmt.Sprintf("5小时 %.1f%% 已用", result.Used),
			percent: result.Used,
			window:  fiveHour,
			label:   factoryResetLabel("", "5小时"),
		})
		if coreMax > result.Percent {
			result.Percent = coreMax
		}
	}
	if std.Weekly != nil && std.Weekly.UsedPercent != nil {
		p := clampPercent(*std.Weekly.UsedPercent)
		segs = append(segs, factorySegment{
			text:    fmt.Sprintf("周 %.1f%% 已用", p),
			percent: p,
			window:  std.Weekly,
			label:   factoryResetLabel("", "周"),
		})
	}
	// 常驻展示窗口(5h/周)中最紧张者;月度仅在严格超过它时才需要补充说明
	shownMax := result.Used
	if takesOver {
		shownMax = coreMax
	}
	if std.Weekly != nil && std.Weekly.UsedPercent != nil {
		if p := clampPercent(*std.Weekly.UsedPercent); p > shownMax {
			shownMax = p
		}
	}
	if m := std.Monthly; m != nil && m.UsedPercent != nil {
		if p := clampPercent(*m.UsedPercent); p > shownMax {
			segs = append(segs, factorySegment{
				text:    fmt.Sprintf("月 %.1f%% 已用", p),
				percent: p,
				window:  m,
				label:   factoryResetLabel("", "月"),
			})
		}
	}
	if coreMax > 0 && !takesOver {
		// Core 仓只在 Standard 限流后接管计费,一有消耗就值得展示
		// (用户常驻开源模型 + overage 回退 Core 的场景);接管时主体就是 Core,不再重复。
		driver, driverName := factoryDrivingWindow(core)
		segs = append(segs, factorySegment{
			text:    fmt.Sprintf("Core %.1f%% 已用", coreMax),
			percent: coreMax,
			window:  driver,
			label:   factoryResetLabel("Core ", driverName),
		})
	}
	if limits.ExtraUsageBalanceCents != nil && *limits.ExtraUsageBalanceCents > 0 {
		segs = append(segs, factorySegment{
			text:  fmt.Sprintf("Extra $%.2f", *limits.ExtraUsageBalanceCents/100),
			extra: true,
		})
	}

	// 排序:接管主体固定首位 → 其余按已用百分比降序(平手保留 5小时→周→月 的先者)→ Extra 收尾。
	sort.SliceStable(segs, func(i, j int) bool {
		if segs[i].pinned != segs[j].pinned {
			return segs[i].pinned
		}
		if segs[i].extra != segs[j].extra {
			return !segs[i].extra
		}
		if segs[i].pinned || segs[i].extra {
			return false
		}
		return segs[i].percent > segs[j].percent
	})

	texts := make([]string, len(segs))
	for i, s := range segs {
		texts[i] = s.text
	}
	result.Remaining = strings.Join(texts, " · ")

	// 倒计时跟排头那一段的窗口走(它才是驱动球色/进度条的那个窗口),标签标明它是谁的。
	// 重置时间一律由 secondsRemaining 推算:真机里 windowEnd 可能是过期值,不可信。
	// 该窗口没有秒数 ⇒ ResetAt 与 ResetLabel 都留空,不借别的窗口的时间(2026-10-19 用户拍板)。
	if len(segs) > 0 {
		lead := segs[0]
		if !lead.extra && lead.window != nil &&
			lead.window.SecondsRemaining != nil && *lead.window.SecondsRemaining > 0 {
			result.ResetAt = time.Now().
				Add(time.Duration(*lead.window.SecondsRemaining * float64(time.Second))).
				UTC().Format(time.RFC3339)
			result.ResetLabel = lead.label
		}
	}
	return result
}

// parseLegacyUsage 解析旧账单模型(Standard/Premium 订阅用量)。
func (f *FactoryDroidFetcher) parseLegacyUsage(result QuotaResult, legacy *factoryLegacyResp) QuotaResult {
	usage := legacy.Usage
	stdPercent, stdOK := factoryLegacyPercent(usage.Standard)
	premiumPercent, premiumOK := factoryLegacyPercent(usage.Premium)
	if !stdOK && !premiumOK {
		result.Error = "无法解析额度数据(响应结构可能已变化)"
		return result
	}

	percent := 0.0
	var remain []string
	for _, p := range []struct {
		label string
		value float64
		ok    bool
	}{
		{"Standard", stdPercent, stdOK},
		{"Premium", premiumPercent, premiumOK},
	} {
		if !p.ok {
			continue
		}
		if p.value > percent {
			percent = p.value
		}
		remain = append(remain, fmt.Sprintf("%s %.1f%%", p.label, p.value))
	}
	result.Used = stdPercent
	result.Total = 100
	result.Percent = percent
	if usage.EndDate > 0 {
		result.ResetAt = time.UnixMilli(usage.EndDate).UTC().Format(time.RFC3339)
	}
	result.Remaining = strings.Join(remain, " · ") + " 已用"
	return result
}

// factoryLegacyPercent 由旧模型的绝对值/比率推已用百分比。
// 已知脏数据:usedRatio 可能恒 0 而绝对值显示有消耗,故绝对值可信时优先。
func factoryLegacyPercent(b *factoryLegacyBucket) (float64, bool) {
	if b == nil {
		return 0, false
	}
	allowanceReliable := b.TotalAllowance != nil && *b.TotalAllowance > 0 && *b.TotalAllowance <= 1e12
	usedReliable := b.UserTokens != nil && *b.UserTokens >= 0
	ratio := -1.0
	ratioSane := false
	if b.UsedRatio != nil {
		ratio = *b.UsedRatio
		ratioSane = ratio >= -0.001 && ratio <= 1.001 &&
			!(ratio == 0 && usedReliable && allowanceReliable && *b.UserTokens > 0)
	}
	switch {
	case ratioSane:
		return clampPercent(ratio * 100), true
	case allowanceReliable && usedReliable:
		return clampPercent(*b.UserTokens / *b.TotalAllowance * 100), true
	case b.UsedRatio != nil && ratio >= -0.1 && ratio <= 100.1:
		// 兜底:比率以 0-100 而非 0-1 传入的老格式
		return clampPercent(ratio), true
	default:
		return 0, false
	}
}

// factoryDrivingWindow 在池内挑出"最紧张"的窗口(已用百分比最大;平手按
// fiveHour → weekly → monthly 取先者),并返回它的中文窗口名(供倒计时标签用)。
// 无任何有数据的窗口时返回 (nil, "")。
func factoryDrivingWindow(p *factoryPool) (*factoryWindow, string) {
	if p == nil {
		return nil, ""
	}
	var driver *factoryWindow
	name := ""
	for _, w := range []struct {
		name string
		win  *factoryWindow
	}{{"5小时", p.FiveHour}, {"周", p.Weekly}, {"月", p.Monthly}} {
		if w.win == nil || w.win.UsedPercent == nil {
			continue
		}
		if driver == nil || clampPercent(*w.win.UsedPercent) > clampPercent(*driver.UsedPercent) {
			driver, name = w.win, w.name
		}
	}
	return driver, name
}

// factoryResetLabel 拼倒计时的窗口标签(如 "月额度" / "Core 周额度");
// 窗口名未知时返回空串,前端退回「距下次刷新」的通用措辞。
func factoryResetLabel(pool, window string) string {
	if window == "" {
		return ""
	}
	return pool + window + "额度"
}

func clampPercent(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// getJSON 发起一次带 Bearer 认证与官方 web 端同款请求头的 GET,返回 200 的响应体。
// 401/403 判定为 Key 失效,429 透出限流,其余非 200 透出状态码。
func (f *FactoryDroidFetcher) getJSON(client *http.Client, key, route string) ([]byte, error) {
	url := strings.TrimRight(f.baseURL, "/") + route
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://app.factory.ai")
	req.Header.Set("Referer", "https://app.factory.ai/")
	req.Header.Set("x-factory-client", "web-app")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 401 || resp.StatusCode == 403:
		return nil, fmt.Errorf("API Key 无效或已过期,请在 app.factory.ai/settings/api-keys 重新生成")
	case resp.StatusCode == 429:
		return nil, fmt.Errorf("请求过于频繁(HTTP 429),请稍后再试")
	case resp.StatusCode != 200:
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// readFactoryEnvKey 从 ~/.factory/.env 读取 FACTORY_API_KEY(Droid CLI 同款格式:
// 支持 export 前缀、单双引号与行尾注释)。文件缺失或未配置时返回空字符串。
func readFactoryEnvKey(envFile string) string {
	data, err := os.ReadFile(envFile)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		line = strings.TrimPrefix(line, "export ")
		if !strings.HasPrefix(line, "FACTORY_API_KEY") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, "FACTORY_API_KEY"))
		rest = strings.TrimSpace(strings.TrimPrefix(rest, "="))
		if i := strings.Index(rest, "#"); i >= 0 {
			rest = strings.TrimSpace(rest[:i])
		}
		rest = strings.Trim(rest, "'\"")
		if rest != "" {
			return rest
		}
	}
	return ""
}
