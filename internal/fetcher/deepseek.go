package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DeepSeekFetcher 通过 DeepSeek 官方 API Key 查询账户余额。
// 端点: GET https://api.deepseek.com/user/balance
// 认证: Authorization: Bearer sk-xxx
// 余额型 Provider: Kind=balance,Percent 无意义(恒 0),Remaining 展示余额。
type DeepSeekFetcher struct {
	apiKey string
	apiURL string
}

// NewDeepSeekFetcher 创建一个新的 DeepSeekFetcher。
// apiURL 可为空,默认为线上端点(便于测试覆盖)。
func NewDeepSeekFetcher(apiKey string) *DeepSeekFetcher {
	return &DeepSeekFetcher{apiKey: apiKey}
}

type deepseekBalanceInfo struct {
	Currency        string `json:"currency"`
	TotalBalance    string `json:"total_balance"`
	GrantedBalance  string `json:"granted_balance"`
	ToppedUpBalance string `json:"topped_up_balance"`
}

type deepseekBalanceResponse struct {
	IsAvailable  bool                  `json:"is_available"`
	BalanceInfos []deepseekBalanceInfo `json:"balance_infos"`
}

func (d *DeepSeekFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "DeepSeek",
		Kind:        KindBalance,
		LastUpdated: time.Now(),
	}

	if d.apiKey == "" {
		result.Error = "未配置 DeepSeek API Key"
		return result
	}

	url := d.apiURL
	if url == "" {
		url = "https://api.deepseek.com/user/balance"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}
	req.Header.Set("Authorization", "Bearer "+d.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 401, 403:
		result.Error = "API Key 无效或已过期"
		return result
	}
	if resp.StatusCode != 200 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	var body deepseekBalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Error = fmt.Sprintf("解析响应失败: %v", err)
		return result
	}

	if !body.IsAvailable {
		result.Error = "账户余额不可用或已耗尽"
		return result
	}
	if len(body.BalanceInfos) == 0 {
		result.Error = "响应中未找到余额数据"
		return result
	}

	// 平台可能返回多个币种(如 USD 0.00 + CNY 247.51),取第一条非零余额
	var best deepseekBalanceInfo
	for _, b := range body.BalanceInfos {
		if total, _ := strconv.ParseFloat(b.TotalBalance, 64); total > 0 {
			best = b
			break
		}
	}
	if best.Currency == "" {
		result.Error = "账户余额不足"
		return result
	}

	result.Remaining = fmt.Sprintf("余额 %s%s (%s)", currencySymbol(best.Currency), best.TotalBalance, best.Currency)
	return result
}

// currencySymbol 返回货币符号(未知币种返回空,直接显示数值)。
func currencySymbol(cur string) string {
	switch strings.ToUpper(cur) {
	case "CNY":
		return "¥"
	case "USD":
		return "$"
	}
	return ""
}
