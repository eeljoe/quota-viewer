package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// KimiFetcher 通过 Kimi Code API Key 查询额度。
// 端点: GET https://api.kimi.com/coding/v1/usages
// 认证: Authorization: Bearer sk-kimi-xxx
type KimiFetcher struct {
	apiKey string
}

func NewKimiFetcher(apiKey string) *KimiFetcher {
	return &KimiFetcher{apiKey: apiKey}
}

type kimiUsagesResponse struct {
	Data []kimiUsageItem `json:"data"`
}

type kimiUsageItem struct {
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

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", "https://api.kimi.com/coding/v1/usages", nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}
	req.Header.Set("Authorization", "Bearer "+k.apiKey)
	req.Header.Set("User-Agent", "KimiCLI/1.6")

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

	var body kimiUsagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Error = fmt.Sprintf("解析响应失败: %v", err)
		return result
	}

	// 找 model_name == "all" 的汇总条目
	for _, item := range body.Data {
		if item.ModelName == "all" {
			result.Used = float64(item.Used)
			result.Total = float64(item.Limit)
			if item.Limit > 0 {
				result.Percent = float64(item.Used) / float64(item.Limit) * 100
			}
			result.Remaining = fmt.Sprintf("%s / %s", formatNum(item.Used), formatNum(item.Limit))
			result.ResetAt = item.ResetAt
			return result
		}
	}

	// 如果没有 "all" 条目,取第一条
	if len(body.Data) > 0 {
		item := body.Data[0]
		result.Used = float64(item.Used)
		result.Total = float64(item.Limit)
		if item.Limit > 0 {
			result.Percent = float64(item.Used) / float64(item.Limit) * 100
		}
		result.Remaining = fmt.Sprintf("%s / %s", formatNum(item.Used), formatNum(item.Limit))
		result.ResetAt = item.ResetAt
		return result
	}

	result.Error = "响应中未找到用量数据"
	return result
}

func formatNum(n int64) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
