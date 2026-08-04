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

	// 响应结构: {"code":0,"data":{"usage":{"percent":0.22,"items":[{"name":"plan_total_token","used":8331114938,"limit":38000000000,"percent":0.22}]}}}
	var raw struct {
		Code int `json:"code"`
		Data struct {
			Usage struct {
				Percent float64 `json:"percent"`
				Items   []struct {
					Name    string  `json:"name"`
					Used    float64 `json:"used"`
					Limit   float64 `json:"limit"`
					Percent float64 `json:"percent"`
				} `json:"items"`
			} `json:"usage"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		result.Error = fmt.Sprintf("解析响应失败: %v", err)
		return result
	}

	// 取 usage.items 中第一个有 used/limit 的条目
	for _, item := range raw.Data.Usage.Items {
		if item.Limit > 0 {
			result.Used = item.Used
			result.Total = item.Limit
			result.Percent = item.Used / item.Limit * 100
			result.Remaining = fmt.Sprintf("%s / %s Credits", formatNum(item.Used), formatNum(item.Limit))
			return result
		}
	}

	// 兜底:用顶层 percent
	if raw.Data.Usage.Percent > 0 {
		result.Percent = raw.Data.Usage.Percent * 100
		result.Remaining = fmt.Sprintf("%.1f%%", raw.Data.Usage.Percent*100)
		return result
	}

	result.Error = "响应中未找到用量数据"
	return result
}
