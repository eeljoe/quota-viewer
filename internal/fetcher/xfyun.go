package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// XfyunFetcher 通过 Cookie 抓取讯飞星辰 MaaS 额度。
type XfyunFetcher struct {
	cookie string
	apiURL string // 额度 XHR 接口地址,首次需 F12 确认
}

func NewXfyunFetcher(cookie string, apiURL string) *XfyunFetcher {
	// 默认 URL,待 F12 确认后更新
	if apiURL == "" {
		apiURL = "https://maas.xfyun.cn/api/packageSubscription"
	}
	return &XfyunFetcher{cookie: cookie, apiURL: apiURL}
}

// xfyunRawResponse 是通用的响应结构,字段名按实抓结果调整。
// 常见模式: { "data": { "used": X, "total": Y, "resetTime": "..." } }
// 或: { "data": [{ "name": "...", "used": X, "surplus": Y }] }
type xfyunRawResponse struct {
	Data json.RawMessage `json:"data"`
	Code int             `json:"code"`
}

func (x *XfyunFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "讯飞星辰",
		LastUpdated: time.Now(),
	}

	if x.cookie == "" {
		result.Error = "未配置讯飞 Cookie"
		return result
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		// 讯飞 401/302 时不自动跟随到登录页
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", x.apiURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

	// 设置 Cookie header
	req.Header.Set("Cookie", x.cookie)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://maas.xfyun.cn/packageSubscription")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 302 {
		result.Error = "Cookie 已过期,请更新"
		return result
	}
	if resp.StatusCode != 200 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result
	}

	// 解析通用结构
	var raw xfyunRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		// 可能不是标准 JSON,保存原始文本
		result.Remaining = string(body[:min(len(body), 200)])
		result.Error = "响应格式未识别(需 F12 确认实际接口)"
		return result
	}

	// 尝试解析 data 为对象
	var dataMap map[string]interface{}
	if err := json.Unmarshal(raw.Data, &dataMap); err == nil {
		result = parseXfyunDataMap(result, dataMap)
		return result
	}

	// 尝试解析 data 为数组
	var dataArray []map[string]interface{}
	if err := json.Unmarshal(raw.Data, &dataArray); err == nil && len(dataArray) > 0 {
		result = parseXfyunDataMap(result, dataArray[0])
		return result
	}

	result.Error = "无法解析额度数据(需 F12 确认实际字段)"
	return result
}

// parseXfyunDataMap 从 data 对象中提取用量信息。
// 字段名按常见命名尝试,实际需根据 F12 抓包结果调整。
func parseXfyunDataMap(result QuotaResult, m map[string]interface{}) QuotaResult {
	used := getFloat(m, "used", "usage", "usedAmount", "consume")
	total := getFloat(m, "total", "limit", "totalAmount", "quota")
	remaining := getFloat(m, "remaining", "surplus", "left")
	resetAt := getString(m, "resetTime", "reset_at", "resetAt", "expireTime", "expireTimeStr")

	if total > 0 {
		result.Total = total
		result.Used = used
		result.Percent = used / total * 100
		result.Remaining = fmt.Sprintf("%.0f / %.0f", used, total)
	} else if remaining > 0 && total > 0 {
		result.Total = total
		result.Used = total - remaining
		result.Percent = result.Used / total * 100
		result.Remaining = fmt.Sprintf("%.0f / %.0f", result.Used, total)
	} else {
		// 无数值,尝试取文本描述
		desc := getString(m, "desc", "description", "remaining", "surplusText")
		if desc != "" {
			result.Remaining = desc
		} else {
			result.Error = "未找到用量字段(需 F12 确认实际字段名)"
		}
	}

	result.ResetAt = resetAt
	return result
}

func getFloat(m map[string]interface{}, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch n := v.(type) {
			case float64:
				return n
			case int:
				return float64(n)
			case int64:
				return float64(n)
			case string:
				var f float64
				if _, err := fmt.Sscanf(n, "%f", &f); err == nil {
					return f
				}
			}
		}
	}
	return 0
}

func getString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
