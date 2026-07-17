package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// MiMoFetcher 通过 Cookie 调用小米 MiMo 额度 JSON API。
// 端点: GET https://platform.xiaomimimo.com/api/v1/tokenPlan/usage
// 认证: Cookie 头(需包含 httponly cookie,前端 document.cookie 取不到)
// Referer: https://platform.xiaomimimo.com/console/plan-manage
type MiMoFetcher struct {
	cookie string
	apiURL string
}

func NewMiMoFetcher(cookie string, apiURL string) *MiMoFetcher {
	if apiURL == "" {
		apiURL = "https://platform.xiaomimimo.com/api/v1/tokenPlan/usage"
	}
	return &MiMoFetcher{cookie: cookie, apiURL: apiURL}
}

func (m *MiMoFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "小米MiMo",
		LastUpdated: time.Now(),
	}

	if m.cookie == "" {
		result.Error = "未配置 MiMo Cookie"
		return result
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", m.apiURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

	req.Header.Set("Cookie", m.cookie)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://platform.xiaomimimo.com/console/plan-manage")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		result.Error = "Cookie 已过期或不足(需包含 httponly cookie),请更新"
		return result
	}
	if resp.StatusCode == 302 {
		result.Error = "Cookie 已过期或不足(需包含 httponly cookie),请更新"
		return result
	}
	if resp.StatusCode != 200 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	// 响应字段名尚未确认,使用通用 JSON 解析后按常见字段名尝试提取。
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = fmt.Sprintf("解析响应失败: %v", err)
		return result
	}

	// 在响应中递归查找 used/total 配对(字段名未知,尝试常见命名)。
	usedFields := []string{"usedCredits", "creditsUsed", "used", "used_tokens", "usedTokens", "consumed", "consumedCredits"}
	totalFields := []string{"totalCredits", "creditsTotal", "total", "total_tokens", "totalTokens", "limit", "creditsLimit"}

	used, hasUsed := mimoFindNumber(raw, usedFields)
	total, hasTotal := mimoFindNumber(raw, totalFields)

	if !hasUsed || !hasTotal {
		// 仅找到 used 或百分比时也尽量给出信息
		result.Error = "响应中未找到 used/total 用量字段,响应结构可能已变更"
		return result
	}

	result.Used = used
	result.Total = total
	if total > 0 {
		result.Percent = used / total * 100
	}
	result.Remaining = fmt.Sprintf("%.0f / %.0f Credits", used, total)

	// 重置时间(若存在)
	for _, k := range []string{"resetTime", "resetAt", "reset_at", "expiresAt", "expires_at"} {
		if v, ok := mimoFindString(raw, k); ok {
			result.ResetAt = v
			break
		}
	}

	return result
}

// mimoFindNumber 在 JSON 根对象中查找第一个能解析为数值的字段(按给定候选名顺序)。
// 支持数值与字符串形式的数字。不递归到子对象。
func mimoFindNumber(raw map[string]json.RawMessage, candidates []string) (float64, bool) {
	for _, key := range candidates {
		val, ok := raw[key]
		if !ok {
			continue
		}
		// 先尝试直接解析为 number
		var f float64
		if err := json.Unmarshal(val, &f); err == nil {
			return f, true
		}
		// 再尝试字符串形式
		var s string
		if err := json.Unmarshal(val, &s); err == nil {
			if f, err := parseFloat(s); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}

// mimoFindString 在 JSON 根对象中查找字符串字段。
func mimoFindString(raw map[string]json.RawMessage, key string) (string, bool) {
	val, ok := raw[key]
	if !ok {
		return "", false
	}
	var s string
	if err := json.Unmarshal(val, &s); err == nil {
		return s, true
	}
	return "", false
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
