package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// MiMoFetcher 通过 Cookie 抓取小米 MiMo 额度。
type MiMoFetcher struct {
	cookie string
	apiURL string
}

func NewMiMoFetcher(cookie string, apiURL string) *MiMoFetcher {
	if apiURL == "" {
		apiURL = "https://platform.xiaomimimo.com/api/plan/quota"
	}
	return &MiMoFetcher{cookie: cookie, apiURL: apiURL}
}

type mimoRawResponse struct {
	Data json.RawMessage `json:"data"`
	Code int             `json:"code"`
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

	var raw mimoRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		result.Remaining = string(body[:min(len(body), 200)])
		result.Error = "响应格式未识别(需 F12 确认实际接口)"
		return result
	}

	// 复用 Task 5 的通用解析函数
	var dataMap map[string]interface{}
	if err := json.Unmarshal(raw.Data, &dataMap); err == nil {
		result = parseXfyunDataMap(result, dataMap)
		return result
	}

	var dataArray []map[string]interface{}
	if err := json.Unmarshal(raw.Data, &dataArray); err == nil && len(dataArray) > 0 {
		result = parseXfyunDataMap(result, dataArray[0])
		return result
	}

	result.Error = "无法解析额度数据(需 F12 确认实际字段)"
	return result
}
