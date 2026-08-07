package fetcher

import "time"

// 结果类型常量。
const (
	// KindUsage 用量型(默认):Percent = 已用百分比。
	KindUsage = "usage"
	// KindBalance 余额型(如 DeepSeek):Percent 无意义,Remaining 展示余额。
	KindBalance = "balance"
)

// QuotaResult 是所有 fetcher 的统一返回结构。
type QuotaResult struct {
	Platform    string    `json:"platform"`    // 展示名("Kimi" / "讯飞星辰" / ...)
	ID          string    `json:"id"`          // provider id(注册表,前端对齐球格)
	Abbr        string    `json:"abbr"`        // 球格缩写("K" / "讯" / "Go" / ...)
	Kind        string    `json:"kind"`        // "usage"(默认) / "balance"
	Used        float64   `json:"used"`        // 已用量
	Total       float64   `json:"total"`       // 总量(平台返回则填,否则 0)
	Percent     float64   `json:"percent"`     // Used/Total * 100;无总量时由剩余百分比反推
	Balance     float64   `json:"balance"`     // 余额型 Provider 的原始余额数值(Kind=balance 时有效)
	Currency    string    `json:"currency"`    // 余额货币代码(如 "CNY" / "USD")
	Remaining   string    `json:"remaining"`   // 原始剩余描述(如 "1,200/18,000 次" 或 "无限制")
	ResetAt     string    `json:"reset_at"`    // 下次重置时间(ISO 8601,空则未知)
	LastUpdated time.Time `json:"last_updated"`
	Error       string    `json:"error"`       // 非空表示失败
}

// Fetcher 是额度查询器的统一接口。
type Fetcher interface {
	Fetch() QuotaResult
}
