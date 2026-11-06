package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MistralFetcher 通过 Mistral API Key 查询月度预算用量（第 12 家 Provider）。
//
// 端点: GET {base}/v1/billing/budget
// 认证: x-api-key: <Mistral API Key>
//
// 口径说明（2026-11-06 用真实 Key 实测，非推测）:
//   - Mistral 没有"剩余额度"这种口径，只有**每月预算 + 已用百分比**。
//     Education 计划给两张独立预算：Vibe（CLI/IDE/Web 编程会话）$150/月、
//     API（Studio 模型调用）$15/月，共享同一个月度重置点。
//   - 响应的顶层还镜像了一份 api_budget 的字段（usage_percentage / initial_budget /
//     currency / reset_at）—— 那是给只认顶层的老客户端留的兼容面，**本实现不读**，
//     否则 Free 套餐或服务端改版时会把 API 预算误当总额。
//   - 球色由 Vibe 预算驱动（用户 2026-11-06 拍板：两张 10:1，取 Vibe）；API 预算只在
//     Remaining 里作补充段。
//   - usage_percentage 是**已用**百分比（0-100），实测有 8 位小数；这种小消耗下百分比
//     没有分辨率（$0.0069 已用 = 0.0046%），所以金额进 Remaining 文本。
//   - 该路由是官方 Beta 接口（OpenAPI tag: beta.billing, security: DashboardContextAuth），
//     无稳定性承诺；失效时对照 https://docs.mistral.ai/openapi.yaml 的
//     `/v1/billing/budget` 与 `BudgetDetail` 更新本文件。
type MistralFetcher struct {
	apiKey  string
	baseURL string // 可重写,用于测试
}

// mistralBudgetPath 是预算查询路由。
const mistralBudgetPath = "/v1/billing/budget"

// NewMistralFetcher 创建一个新的 MistralFetcher。
func NewMistralFetcher(apiKey string) *MistralFetcher {
	return &MistralFetcher{
		apiKey:  strings.TrimSpace(apiKey),
		baseURL: "https://api.mistral.ai",
	}
}

// mistralBudgetDetail 是单张预算(OpenAPI `BudgetDetail`)。
//
// 全部数值用指针:字段缺席与"值为 0"是两回事——服务端省略零值字段是该团队的一贯行为
// (Devin 那次就是靠这点定位到真实 bug),所以不能把缺席读成 0%。
type mistralBudgetDetail struct {
	UsagePercentage *float64 `json:"usage_percentage"`
	InitialBudget   *float64 `json:"initial_budget"`
	Currency        *string  `json:"currency"`
	ResetAt         string   `json:"reset_at"`
}

// mistralBudgetResp 对应 /v1/billing/budget 的响应。
//
// 只取两张嵌套预算;顶层的同名镜像字段故意不映射(见文件头注释)。
type mistralBudgetResp struct {
	Vibe *mistralBudgetDetail `json:"vibe_budget"`
	API  *mistralBudgetDetail `json:"api_budget"`
}

// usable 判断一张预算是否可用于驱动展示(存在且给了已用百分比)。
func (d *mistralBudgetDetail) usable() bool {
	return d != nil && d.UsagePercentage != nil
}

// currency 返回 ISO 货币码;字段缺席时返回空串。
func (d *mistralBudgetDetail) currency() string {
	if d == nil || d.Currency == nil {
		return ""
	}
	return *d.Currency
}

// Fetch 拉取并解析 Mistral 预算用量。
func (f *MistralFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "Mistral",
		Kind:        KindUsage,
		LastUpdated: time.Now(),
	}

	if f.apiKey == "" {
		result.Error = "未配置 Mistral API Key"
		return result
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := strings.TrimRight(f.baseURL, "/") + mistralBudgetPath
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}
	req.Header.Set("x-api-key", f.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		result.Error = "API Key 无效或已过期"
		return result
	}
	if resp.StatusCode != 200 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	var body mistralBudgetResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Error = fmt.Sprintf("解析响应失败: %v", err)
		return result
	}

	return f.applyBudget(&body)
}

// applyBudget 把预算快照映射成 QuotaResult(AS-3..AS-10)。
//
// 驱动窗口:Vibe 可用就用 Vibe,否则退 API,都不可用报结构错误。
// 兜底方向是"往上退到 API"而不是"报错":API 预算同样是一个真实的月度额度,
// 用户至少能看到它,比整行空白好。
func (f *MistralFetcher) applyBudget(body *mistralBudgetResp) QuotaResult {
	result := QuotaResult{
		Platform:    "Mistral",
		Kind:        KindUsage,
		LastUpdated: time.Now(),
	}

	driver := body.Vibe
	label := "Vibe 额度"
	if !driver.usable() {
		driver = body.API
		label = "API 额度"
	}
	if !driver.usable() {
		result.Error = "响应中未找到预算数据"
		return result
	}

	used := clampPercent(*driver.UsagePercentage)
	result.Used = used
	result.Total = 100
	result.Percent = used
	result.ResetAt = driver.ResetAt
	result.ResetLabel = label
	result.Remaining = mistralRemaining(body.Vibe, body.API)

	return result
}

// mistralRemaining 生成面板右侧的剩余描述("Vibe $0.01/$150 · API $0.00/$15")。
//
// 每张可用预算一段,顺序固定 Vibe → API(与它们的重要性一致,不按数值排序:
// 两张本就是不同用途,把 API 排前面只会让主战场那一张被挤到右边)。
// Remaining 有 210px 上限,所以整数预算不带小数、消耗固定两位小数。
func mistralRemaining(vibe, api *mistralBudgetDetail) string {
	segs := make([]string, 0, 2)
	if vibe.usable() {
		segs = append(segs, "Vibe "+mistralBudgetText(vibe))
	}
	if api.usable() {
		segs = append(segs, "API "+mistralBudgetText(api))
	}
	return strings.Join(segs, " · ")
}

// mistralBudgetText 生成单张预算的文本:有预算金额时是 "$0.01/$150",否则退百分比。
func mistralBudgetText(d *mistralBudgetDetail) string {
	if d.InitialBudget != nil && *d.InitialBudget > 0 {
		spent := *d.UsagePercentage / 100 * *d.InitialBudget
		return mistralMoney(spent, d.currency()) + "/" + mistralMoney(*d.InitialBudget, d.currency())
	}
	// 没有预算金额时百分比是唯一可读的量(与 Kimi 的退化展示同口径)。
	return kimiFormatPercent(clampPercent(*d.UsagePercentage)) + "% 已用"
}

// mistralMoney 格式化金额:整数不带小数,其余保留两位;未知币种只显示数值。
func mistralMoney(v float64, currency string) string {
	v = clampNegativeZero(v)
	var num string
	if v == float64(int64(v)) {
		num = strconv.FormatInt(int64(v), 10)
	} else {
		num = strconv.FormatFloat(v, 'f', 2, 64)
	}
	return currencySymbol(currency) + num
}

// clampNegativeZero 把"四舍五入后是 -0"的小负数归零,避免面板出现 "$-0.00"。
func clampNegativeZero(v float64) float64 {
	if v > -0.005 && v < 0 {
		return 0
	}
	return v
}
