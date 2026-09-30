package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
// Percent 取全部窗口中更紧张的一个(窗口告警契约,与 Kimi/Ollama/Command Code 一致),
// Used/Total/ResetAt 以 5 小时主窗口为准。
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

// parseTokenRateLimits 解析新账单模型。窗口只有百分比、无绝对值,沿用 Ollama 先例:
// Total=100、Used=5h 百分比;Percent 取全部窗口(含 Core 池)最紧张的一个。
func (f *FactoryDroidFetcher) parseTokenRateLimits(result QuotaResult, limits *factoryLimitsResp) QuotaResult {
	std := limits.Limits.Standard
	fiveHour := std.FiveHour

	percent := 0.0
	poolPercent := func(p *factoryPool) {
		if p == nil {
			return
		}
		for _, w := range []*factoryWindow{p.FiveHour, p.Weekly, p.Monthly} {
			if w != nil && w.UsedPercent != nil && clampPercent(*w.UsedPercent) > percent {
				percent = clampPercent(*w.UsedPercent)
			}
		}
	}
	poolPercent(std)
	coreTightest := 0.0
	if core := limits.Limits.Core; core != nil {
		// Core 池是独立计费窗口,仅在确实有数据时纳入告警与展示。
		hasData := false
		for _, w := range []*factoryWindow{core.FiveHour, core.Weekly, core.Monthly} {
			if w != nil && w.UsedPercent != nil {
				hasData = true
				break
			}
		}
		if hasData {
			poolPercent(core)
			coreTightest = maxCorePercent(core)
		}
	}

	if fiveHour == nil || fiveHour.UsedPercent == nil {
		result.Error = "未找到 5 小时窗口额度,响应结构可能已变化"
		return result
	}

	result.Used = clampPercent(*fiveHour.UsedPercent)
	result.Total = 100
	result.Percent = percent
	if fiveHour.SecondsRemaining != nil && *fiveHour.SecondsRemaining > 0 {
		result.ResetAt = time.Now().
			Add(time.Duration(*fiveHour.SecondsRemaining * float64(time.Second))).
			UTC().Format(time.RFC3339)
	}

	var remain []string
	if std.Weekly != nil && std.Weekly.UsedPercent != nil {
		remain = append(remain, fmt.Sprintf("周 %.1f%% 已用", clampPercent(*std.Weekly.UsedPercent)))
	}
	if std.Monthly != nil && std.Monthly.UsedPercent != nil {
		remain = append(remain, fmt.Sprintf("月 %.1f%% 已用", clampPercent(*std.Monthly.UsedPercent)))
	}
	if coreTightest > 0 {
		remain = append(remain, fmt.Sprintf("Core %.1f%% 已用", coreTightest))
	}
	if limits.ExtraUsageBalanceCents != nil && *limits.ExtraUsageBalanceCents > 0 {
		remain = append(remain, fmt.Sprintf("Extra $%.2f", *limits.ExtraUsageBalanceCents/100))
	}
	result.Remaining = strings.Join(remain, " · ")
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

// maxCorePercent 返回 Core 池中最紧张的窗口百分比。
func maxCorePercent(p *factoryPool) float64 {
	tightest := 0.0
	for _, w := range []*factoryWindow{p.FiveHour, p.Weekly, p.Monthly} {
		if w != nil && w.UsedPercent != nil && clampPercent(*w.UsedPercent) > tightest {
			tightest = clampPercent(*w.UsedPercent)
		}
	}
	return tightest
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
