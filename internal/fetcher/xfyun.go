package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// XfyunFetcher 通过 Cookie 调用讯飞星辰 MaaS 额度 JSON API。
// 端点: GET https://maas.xfyun.cn/api/v1/gpt-finetune/coding-plan/list
// 认证: Cookie 头
// Referer: https://maas.xfyun.cn/packageSubscription
type XfyunFetcher struct {
	cookie string
	apiURL string
}

func NewXfyunFetcher(cookie string, apiURL string) *XfyunFetcher {
	if apiURL == "" {
		apiURL = "https://maas.xfyun.cn/api/v1/gpt-finetune/coding-plan/list"
	}
	return &XfyunFetcher{cookie: cookie, apiURL: apiURL}
}

// xfyunResponse 对应 /coding-plan/list 的 JSON 响应。
type xfyunResponse struct {
	Code int          `json:"code"`
	Data xfyunPageData `json:"data"`
}

type xfyunPageData struct {
	Page int           `json:"page"`
	Rows []xfyunPlanRow `json:"rows"`
}

type xfyunPlanRow struct {
	AppID             string              `json:"appId"`
	CodingPlanUsageDTO xfyunCodingPlanDTO `json:"codingPlanUsageDTO"`
	ExpiresAt         string              `json:"expiresAt"`
	ID                int64               `json:"id"`
}

// xfyunCodingPlanDTO 包含 5 小时窗口、周窗口、套餐总量的用量与限额。
type xfyunCodingPlanDTO struct {
	RP5hUsage     float64 `json:"rp5hUsage"`
	RP5hLimit     float64 `json:"rp5hLimit"`
	RPwUsage      float64 `json:"rpwUsage"`
	RPwLimit      float64 `json:"rpwLimit"`
	PackageUsage  float64 `json:"packageUsage"`
	PackageLimit  float64 `json:"packageLimit"`
	PackageLeft   float64 `json:"packageLeft"`
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
		// Cookie 过期时讯飞会 302 跳转到登录页,不自动跟随
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", x.apiURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

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

	var body xfyunResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Error = fmt.Sprintf("解析响应失败: %v", err)
		return result
	}

	if body.Code != 0 {
		result.Error = fmt.Sprintf("接口返回错误码: %d", body.Code)
		return result
	}

	if len(body.Data.Rows) == 0 {
		result.Error = "响应中未找到套餐数据"
		return result
	}

	// 取第一条套餐;展示套餐总量
	row := body.Data.Rows[0]
	dto := row.CodingPlanUsageDTO

	used := dto.PackageUsage
	limit := dto.PackageLimit
	result.Used = used
	result.Total = limit
	if limit > 0 {
		result.Percent = used / limit * 100
	}
	result.Remaining = fmt.Sprintf("%s / %s 次 (总量)", formatNum(used), formatNum(limit))
	result.ResetAt = row.ExpiresAt

	return result
}
