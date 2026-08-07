package fetcher

import "fmt"

// defaultBudget 是余额型 Provider 的默认预算(用户未设时使用)。
const defaultBudget = 300

// ApplyBudget 为余额型 Provider 应用预算,计算消耗百分比。
// 用户未设预算(budget <= 0)时使用 defaultBudget;计算逻辑:
// 已消耗 = 预算 - 当前余额;百分比 = 已消耗 / 预算 * 100。
// 余额超过预算时钳制已消耗为 0(用户充值超预算的情况)。
func ApplyBudget(r *QuotaResult, budget float64) {
	if r.Kind != KindBalance || r.Balance < 0 || r.Error != "" {
		return
	}
	if budget <= 0 {
		budget = defaultBudget
	}
	r.Total = budget
	r.Used = budget - r.Balance
	if r.Used < 0 {
		r.Used = 0
	}
	r.Percent = r.Used / budget * 100
	sym := currencySymbol(r.Currency)
	r.Remaining = fmt.Sprintf("%s%.2f / %s%.2f (预算)", sym, r.Balance, sym, budget)
}
