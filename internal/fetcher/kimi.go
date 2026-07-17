package fetcher

import (
	"encoding/json"
	"fmt"
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
// 响应为嵌套对象,usage 为周/主额度,limits[0] 为 5 小时窗口。
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
	// 兼容旧版数组响应:若返回的是 {"data":[...]} 则走旧路径。
	Data []kimiUsageItemLegacy `json:"data"`
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

	// 优先:5 小时窗口(limits[0].detail)
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
		result.Remaining = fmt.Sprintf("%.0f / %.0f (5小时)", used, limit)
		result.ResetAt = d.ResetTime
		return result
	}

	// 兜底:周额度 usage 对象
	if body.Usage.Limit != "" || body.Usage.Used != "" {
		used, _ := kimiParseStringFloat(body.Usage.Used)
		limit, _ := kimiParseStringFloat(body.Usage.Limit)
		result.Used = used
		result.Total = limit
		if limit > 0 {
			result.Percent = used / limit * 100
		}
		result.Remaining = fmt.Sprintf("%.0f / %.0f", used, limit)
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
			result.Remaining = fmt.Sprintf("%.0f / %.0f", float64(item.Used), float64(item.Limit))
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
		result.Remaining = fmt.Sprintf("%.0f / %.0f", float64(item.Used), float64(item.Limit))
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
