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
	Platform    string    `json:"platform"`  // 展示名("Kimi" / "讯飞星辰" / ...)
	ID          string    `json:"id"`        // provider id(注册表,前端对齐球格)
	Abbr        string    `json:"abbr"`      // 球格缩写("K" / "讯" / "Go" / ...)
	Kind        string    `json:"kind"`      // "usage"(默认) / "balance"
	Used        float64   `json:"used"`      // 已用量
	Total       float64   `json:"total"`     // 总量(平台返回则填,否则 0)
	Percent     float64   `json:"percent"`   // Used/Total * 100;无总量时由剩余百分比反推
	Balance     float64   `json:"balance"`   // 余额型 Provider 的原始余额数值(Kind=balance 时有效)
	Currency    string    `json:"currency"`  // 余额货币代码(如 "CNY" / "USD")
	Remaining   string    `json:"remaining"` // 原始剩余描述(如 "1,200/18,000 次" 或 "无限制")
	ResetAt     string    `json:"reset_at"`  // 下次重置时间(ISO 8601,空则未知)
	// ResetLabel 是 ResetAt 所属的窗口名(如 "月额度" / "Core 周额度")。
	// 告警由"最紧张的窗口"驱动时,倒计时必须标明它是谁的,否则会被读成别的窗口。
	// 空串 = 未知来源,前端退回通用措辞。
	ResetLabel  string    `json:"reset_label,omitempty"`
	LastUpdated time.Time `json:"last_updated"`
	Error       string    `json:"error"` // 非空表示失败

	// 监控对象身份(多 Key / 多套餐用;单对象 Provider 两字段留空即旧行为)。
	Slot     string `json:"slot,omitempty"`      // 稳定标识,如 "minimax/0" 或 "minimax/1/video";编排层填充
	SubLabel string `json:"sub_label,omitempty"` // 展示后缀,如套餐名 "视频赠送卡" 或 "Key 2"
}

// Fetcher 是额度查询器的统一接口。
type Fetcher interface {
	Fetch() QuotaResult
}

// MultiFetcher 是可选接口:一次查询返回多个监控对象。
//
// 用于「同一平台 + 同一凭证」下存在多个独立额度池的 Provider——
// 例如 MiniMax 的 model_remains[] 返回主套餐 + 视频/音乐赠送卡,
// 每张卡各有独立的 5 小时 / 周窗口。不实现本接口的 Fetcher 保持单对象语义,
// 上层由 FetchAll 统一兜底,新增 Provider 无需关心多对象。
type MultiFetcher interface {
	FetchAll() []QuotaResult
}

// FetchAll 是上层抓取的统一入口:实现 MultiFetcher 的取全部监控对象,
// 否则把 Fetch() 的单结果包成单元素切片。
func FetchAll(f Fetcher) []QuotaResult {
	if mf, ok := f.(MultiFetcher); ok {
		if rs := mf.FetchAll(); len(rs) > 0 {
			return rs
		}
		// 实现方返回空切片视为契约违约,退回单结果语义避免上层拿到"零监控对象"。
		return []QuotaResult{f.Fetch()}
	}
	return []QuotaResult{f.Fetch()}
}
