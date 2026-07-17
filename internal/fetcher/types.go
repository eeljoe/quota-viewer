package fetcher

import "time"

// QuotaResult 是所有 fetcher 的统一返回结构。
type QuotaResult struct {
	Platform    string    `json:"platform"`     // "Kimi" / "讯飞星辰" / "小米MiMo"
	Used        float64   `json:"used"`         // 已用量
	Total       float64   `json:"total"`        // 总量(平台返回则填,否则 0)
	Percent     float64   `json:"percent"`      // Used/Total * 100;无总量时由剩余百分比反推
	Remaining   string    `json:"remaining"`    // 原始剩余描述(如 "1,200/18,000 次" 或 "无限制")
	ResetAt     string    `json:"reset_at"`     // 下次重置时间(ISO 8601,空则未知)
	LastUpdated time.Time `json:"last_updated"`
	Error       string    `json:"error"`        // 非空表示失败
}

// Fetcher 是额度查询器的统一接口。
type Fetcher interface {
	Fetch() QuotaResult
}
