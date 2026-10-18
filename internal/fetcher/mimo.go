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

// Fetch 是单对象入口:取第一个有额度的桶(保持既有语义与测试)。
// 多额度桶场景用 FetchAll —— 每个桶是独立监控对象。
func (m *MiMoFetcher) Fetch() QuotaResult {
	all := m.FetchAll()
	if len(all) == 0 {
		return QuotaResult{Platform: "小米MiMo", LastUpdated: time.Now()}
	}
	return all[0]
}

// mimoItemLabel 把响应里的桶名翻译成面板上可读的中文标签。
func mimoItemLabel(name string) string {
	switch name {
	case "plan_total_token":
		return "套餐额度"
	case "compensation_total_token":
		return "补偿额度"
	default:
		return name
	}
}

// FetchAll 返回每个有额度的桶一个监控对象。任一全局失败(无 Cookie / 网络 /
// 非 200 / 无任何可用桶)返回单条错误结果,与旧的单对象语义一致。
func (m *MiMoFetcher) FetchAll() []QuotaResult {
	result := QuotaResult{
		Platform:    "小米MiMo",
		LastUpdated: time.Now(),
	}

	if m.cookie == "" {
		result.Error = "未配置 MiMo Cookie"
		return []QuotaResult{result}
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
		return []QuotaResult{result}
	}

	req.Header.Set("Cookie", m.cookie)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://platform.xiaomimimo.com/console/plan-manage")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return []QuotaResult{result}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		result.Error = "Cookie 已过期或不足(需包含 httponly cookie),请更新"
		return []QuotaResult{result}
	}
	if resp.StatusCode == 302 {
		result.Error = "Cookie 已过期或不足(需包含 httponly cookie),请更新"
		return []QuotaResult{result}
	}
	if resp.StatusCode != 200 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return []QuotaResult{result}
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
		return []QuotaResult{result}
	}

	// 每个有 limit 的桶 = 一个独立额度池(如套餐额度 / 补偿额度),各自一个监控对象。
	// 旧实现只取第一个,赠送/补偿额度被静默丢弃。
	out := make([]QuotaResult, 0, len(raw.Data.Usage.Items))
	setter := &planAbbrSetter{}
	for _, item := range raw.Data.Usage.Items {
		if item.Limit <= 0 {
			continue
		}
		r := result
		r.Used = item.Used
		r.Total = item.Limit
		r.Percent = item.Used / item.Limit * 100
		r.Remaining = fmt.Sprintf("%s / %s Credits", formatNum(item.Used), formatNum(item.Limit))
		r.SubLabel = mimoItemLabel(item.Name)
		r.Abbr = setter.next(len(out), item.Name)
		out = append(out, r)
	}
	if len(out) > 0 {
		return out
	}

	// 兜底:用顶层 percent
	if raw.Data.Usage.Percent > 0 {
		result.Percent = raw.Data.Usage.Percent * 100
		result.Remaining = fmt.Sprintf("%.1f%%", raw.Data.Usage.Percent*100)
		return []QuotaResult{result}
	}

	result.Error = "响应中未找到用量数据"
	return []QuotaResult{result}
}
