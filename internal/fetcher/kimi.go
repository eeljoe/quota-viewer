package fetcher

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// KimiFetcher 通过 Kimi Code API Key 查询额度。
// 端点: GET https://api.kimi.com/coding/v1/usages
// 认证: Authorization: Bearer sk-kimi-xxx
// User-Agent: KimiCLI/1.6
//
// 响应为嵌套对象,usage 为周/主额度,limits[0] 与 usages.limit_5h 为 5 小时窗口,
// usages.limit_7d 为 7 天窗口。Percent 取两窗口中更紧张的一个。
type KimiFetcher struct {
	apiKey string
	apiURL string // 可为空,默认为线上端点(便于测试覆盖)
}

func NewKimiFetcher(apiKey string) *KimiFetcher {
	return &KimiFetcher{apiKey: apiKey}
}

// kimiUsageResponse 对应 GET /coding/v1/usages 的实际响应(嵌套对象)。
type kimiUsageResponse struct {
	User  kimiUser  `json:"user"`
	Usage kimiUsage `json:"usage"`
	// Limits 为 5 小时窗口等细分限制;Limits[0] 为 5 小时窗口。
	Limits []kimiLimit `json:"limits"`
	// Usages 为窗口比率(limit_5h / limit_7d);used_ratio 为 0-1 的已用比例。
	Usages kimiUsages `json:"usages"`
	// 兼容旧版数组响应:若返回的是 {"data":[...]} 则走旧路径。
	Data []kimiUsageItemLegacy `json:"data"`
}

type kimiUsages struct {
	Limit5h kimiRatio `json:"limit_5h"`
	Limit7d kimiRatio `json:"limit_7d"`
}

// kimiRatio 为窗口已用比率;used_ratio 用指针以区分"字段缺失"与"0%"。
type kimiRatio struct {
	UsedRatio *float64 `json:"used_ratio"`
	ResetTime string   `json:"reset_time"`
}

type kimiUser struct {
	UserID     string         `json:"userId"`
	Region     string         `json:"region"`
	Membership kimiMembership `json:"membership"`
}

type kimiMembership struct {
	Level string `json:"level"`
}

// kimiUsage 为主(周)额度。limit/used/remaining 为字符串形式。
type kimiUsage struct {
	Limit     string `json:"limit"`
	Used      string `json:"used"`
	Remaining string `json:"remaining"`
	ResetTime string `json:"resetTime"`
}

type kimiLimit struct {
	Window kimiWindow `json:"window"`
	Detail kimiDetail `json:"detail"`
}

type kimiWindow struct {
	Duration int64  `json:"duration"`
	TimeUnit string `json:"timeUnit"`
}

type kimiDetail struct {
	Limit     string `json:"limit"`
	Remaining string `json:"remaining"`
	ResetTime string `json:"resetTime"`
}

// kimiUsageItemLegacy 旧版 {"data":[{model_name,used,limit,...}]} 响应条目。
type kimiUsageItemLegacy struct {
	ModelName string `json:"model_name"`
	Used      int64  `json:"used"`
	Limit     int64  `json:"limit"`
	Remaining int64  `json:"remaining"`
	ResetIn   int64  `json:"reset_in"`
	ResetAt   string `json:"reset_at"`
}

func (k *KimiFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "Kimi",
		LastUpdated: time.Now(),
	}

	if k.apiKey == "" {
		result.Error = "未配置 Kimi API Key"
		return result
	}

	url := k.apiURL
	if url == "" {
		url = "https://api.kimi.com/coding/v1/usages"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}
	req.Header.Set("Authorization", "Bearer "+k.apiKey)
	req.Header.Set("User-Agent", "KimiCLI/1.6")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		result.Error = "API Key 无效或已过期(请确认使用 sk-kimi-xxx 格式的 Key)"
		return result
	}
	if resp.StatusCode != 200 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	var body kimiUsageResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Error = fmt.Sprintf("解析响应失败: %v", err)
		return result
	}

	// 5 小时窗口驱动 Used/Total/Remaining 首项与 ResetAt;
	// Percent 取 5h 与周窗口中更紧张者,任一窗口耗尽都必须告警。
	sessionFound := false
	if len(body.Limits) > 0 && body.Limits[0].Detail.Limit != "" {
		d := body.Limits[0].Detail
		remaining, _ := kimiParseStringFloat(d.Remaining)
		limit, _ := kimiParseStringFloat(d.Limit)
		used := limit - remaining
		result.Used = used
		result.Total = limit
		if limit > 0 {
			result.Percent = used / limit * 100
		}
		result.Remaining = fmt.Sprintf("%s / %s (5小时)", formatNum(used), formatNum(limit))
		result.ResetAt = d.ResetTime
		sessionFound = true
	} else if body.Usages.Limit5h.UsedRatio != nil {
		ratio := kimiRatioPercent(*body.Usages.Limit5h.UsedRatio)
		result.Used = ratio
		result.Total = 100
		result.Percent = ratio
		result.Remaining = fmt.Sprintf("%s%% (5小时)", kimiFormatPercent(ratio))
		result.ResetAt = body.Usages.Limit5h.ResetTime
		sessionFound = true
	}

	weeklyPercent, weeklyFound := kimiWeeklyPercent(body)
	if weeklyFound && weeklyPercent > result.Percent {
		result.Percent = weeklyPercent
	}

	if sessionFound {
		if weeklyFound {
			result.Remaining += fmt.Sprintf(" · 周 %s%% 已用", kimiFormatPercent(weeklyPercent))
		}
		return result
	}

	// 兜底:仅有周额度 usage 对象
	if weeklyFound {
		used, err := kimiParseStringFloat(body.Usage.Used)
		if err != nil {
			used = weeklyPercent
			result.Total = 100
		} else {
			limit, _ := kimiParseStringFloat(body.Usage.Limit)
			result.Total = limit
		}
		result.Used = used
		result.Percent = weeklyPercent
		if result.Total > 0 {
			result.Remaining = fmt.Sprintf("%s / %s", formatNum(result.Used), formatNum(result.Total))
		} else {
			result.Remaining = fmt.Sprintf("%s%% 已用", kimiFormatPercent(weeklyPercent))
		}
		result.ResetAt = body.Usage.ResetTime
		return result
	}

	// 兼容旧版 {"data":[{model_name:"all",...}]} 数组响应
	for _, item := range body.Data {
		if item.ModelName == "all" {
			result.Used = float64(item.Used)
			result.Total = float64(item.Limit)
			if item.Limit > 0 {
				result.Percent = float64(item.Used) / float64(item.Limit) * 100
			}
			result.Remaining = fmt.Sprintf("%s / %s", formatNum(float64(item.Used)), formatNum(float64(item.Limit)))
			result.ResetAt = item.ResetAt
			return result
		}
	}
	if len(body.Data) > 0 {
		item := body.Data[0]
		result.Used = float64(item.Used)
		result.Total = float64(item.Limit)
		if item.Limit > 0 {
			result.Percent = float64(item.Used) / float64(item.Limit) * 100
		}
		result.Remaining = fmt.Sprintf("%s / %s", formatNum(float64(item.Used)), formatNum(float64(item.Limit)))
		result.ResetAt = item.ResetAt
		return result
	}

	result.Error = "响应中未找到用量数据"
	return result
}

// kimiParseStringFloat 解析 Kimi 返回的字符串形式数值(如 "75"、"100")。
func kimiParseStringFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("空字符串")
	}
	return strconv.ParseFloat(s, 64)
}

// kimiWeeklyPercent 返回 7 天(周)窗口已用百分比。
// 优先 usages.limit_7d.used_ratio(显式比率),缺失时回退 usage.used/limit 字符串。
func kimiWeeklyPercent(body kimiUsageResponse) (float64, bool) {
	if body.Usages.Limit7d.UsedRatio != nil {
		return kimiRatioPercent(*body.Usages.Limit7d.UsedRatio), true
	}
	if body.Usage.Limit == "" && body.Usage.Used == "" {
		return 0, false
	}
	used, err := kimiParseStringFloat(body.Usage.Used)
	if err != nil {
		return 0, false
	}
	limit, err := kimiParseStringFloat(body.Usage.Limit)
	if err != nil || limit <= 0 {
		return 0, false
	}
	return used / limit * 100, true
}

// kimiRatioPercent 把 0-1 的 used_ratio 转成百分比。
func kimiRatioPercent(ratio float64) float64 {
	return ratio * 100
}

// kimiFormatPercent 百分比展示:最多保留一位小数,去掉多余的 .0。
func kimiFormatPercent(v float64) string {
	return strconv.FormatFloat(math.Round(v*10)/10, 'f', -1, 64)
}
