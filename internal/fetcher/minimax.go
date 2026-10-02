package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MiniMaxFetcher 通过 MiniMax 开放平台私有路由查询 Token Plan(订阅套餐)额度。
// 端点: GET {region}/v1/api/openplatform/coding_plan/remains
//   - 国内: https://api.minimaxi.com / https://www.minimaxi.com(sk-cp- 订阅 Key)
//   - 海外: https://api.minimax.io
//
// 三个主机组成回退链(实测 2026-10:api.minimaxi.com 对 sk-cp- Key 可用,
// www.minimaxi.com 偶发 TLS 握手超时,海外主机不认国内 Key 返回业务 1004),
// 逐个尝试直到成功。订阅 Key 在 platform.minimaxi.com/console/plan 生成;
// 按量计费 Key 不能用。请求头带官方 web 端同款 Referer。
//
// 响应为套餐卡数组 model_remains[](general 主套餐在首位,视频/音乐等赠送卡在后),
// 本实现取首位主套餐。窗口语义与 Kimi/Ollama 一致:5 小时 + 周双窗口,任一耗尽都告警。
// 注意:字段是 current_*_remaining_percent(剩余%),展示前必须反转为已用%
// (参照 minimax-status 1.2.5 的语义事故,有回归测试守护);
// 部分套餐不返回绝对量(count 恒 0),此时退化为纯百分比展示。
type MiniMaxFetcher struct {
	apiKey   string
	baseURLs []string // 可重写,用于测试;非空时替换默认域名链
}

// minimaxRemainsPath 是额度查询路由(各区域同一路径)。
const minimaxRemainsPath = "/v1/api/openplatform/coding_plan/remains"

// minimaxBaseURLs 是默认域名回退链(国内优先,海外兜底)。
var minimaxBaseURLs = []string{"https://api.minimaxi.com", "https://www.minimaxi.com", "https://api.minimax.io"}

// NewMiniMaxFetcher 创建一个新的 MiniMaxFetcher。
func NewMiniMaxFetcher(apiKey string) *MiniMaxFetcher {
	return &MiniMaxFetcher{
		apiKey:   strings.TrimSpace(apiKey),
		baseURLs: minimaxBaseURLs,
	}
}

// minimaxBaseResp 是 MiniMax 开放平台的统一业务状态头(1004/2049 = 鉴权失败)。
type minimaxBaseResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

// minimaxModelRemains 是单张套餐卡。百分比为"剩余%"语义,用指针区分缺失与 0;
// *_remains_time 为毫秒级重置倒计时。窗口绝对值为 token 数。
type minimaxModelRemains struct {
	ModelName                   string   `json:"model_name"`
	CurrentIntervalTotalCount   float64  `json:"current_interval_total_count"`
	CurrentIntervalRemainingPct *float64 `json:"current_interval_remaining_percent"`
	RemainsTimeMs               float64  `json:"remains_time"`
	CurrentWeeklyTotalCount     float64  `json:"current_weekly_total_count"`
	CurrentWeeklyRemainingPct   *float64 `json:"current_weekly_remaining_percent"`
	WeeklyRemainsTimeMs         float64  `json:"weekly_remains_time"`
}

// minimaxRemainsResp 对应 coding_plan/remains 的成功响应。
type minimaxRemainsResp struct {
	BaseResp     minimaxBaseResp       `json:"base_resp"`
	ModelRemains []minimaxModelRemains `json:"model_remains"`
}

// minimaxFail 是单次请求的失败信息(域名链上游主循环决定是否换下一个主机)。
type minimaxFail struct {
	msg string
}

func (e *minimaxFail) Error() string { return e.msg }

// Fetch 拉取并解析 MiniMax Token Plan 额度。按域名链逐个尝试(网络不通 / 401/403/404 /
// 业务鉴权码都换下一个),全部失败时报第一个错误;5h 窗口驱动主展示与 ResetAt,
// Percent 取 5h 与周窗口中更紧张的一个(窗口告警契约)。
func (f *MiniMaxFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "MiniMax",
		Kind:        KindUsage,
		LastUpdated: time.Now(),
	}

	if f.apiKey == "" {
		result.Error = "未配置 MiniMax Token Plan Key(sk-cp- 开头的订阅 Key,在 platform.minimaxi.com/console/plan 获取)"
		return result
	}

	var body []byte
	var firstFail *minimaxFail
	for _, base := range f.baseURLs {
		b, fail := f.fetchOne(base)
		if fail == nil {
			body = b
			break
		}
		if firstFail == nil {
			firstFail = fail
		}
	}
	if body == nil {
		result.Error = firstFail.msg
		return result
	}

	var resp minimaxRemainsResp
	if err := json.Unmarshal(body, &resp); err != nil {
		result.Error = "解析响应失败(响应结构可能已变化): " + err.Error()
		return result
	}
	if resp.BaseResp.StatusCode != 0 {
		result.Error = minimaxBizError(resp.BaseResp)
		return result
	}
	if len(resp.ModelRemains) == 0 {
		result.Error = "未找到套餐用量数据(响应结构可能已变化)"
		return result
	}
	return f.parseRemains(result, &resp.ModelRemains[0])
}

// parseRemains 解析主套餐卡(取 model_remains 首位,与官方 CLI 生态一致)。
// 展示与 Kimi 对齐:5h 绝对值为主项,周窗口按已用百分比追加;周无限制明确标出。
func (f *MiniMaxFetcher) parseRemains(result QuotaResult, m *minimaxModelRemains) QuotaResult {
	if m.CurrentIntervalRemainingPct == nil {
		result.Error = "未找到 5 小时窗口用量(响应结构可能已变化)"
		return result
	}

	fiveUsed := clampPercent(100 - *m.CurrentIntervalRemainingPct)
	// "无周限"判定(minimax-status 同款):周总额=0 且未返回周剩余百分比
	weeklyUnlimited := m.CurrentWeeklyTotalCount == 0 && m.CurrentWeeklyRemainingPct == nil

	percent := fiveUsed
	weeklyUsed := -1.0
	if !weeklyUnlimited && m.CurrentWeeklyRemainingPct != nil {
		weeklyUsed = clampPercent(100 - *m.CurrentWeeklyRemainingPct)
		if weeklyUsed > percent {
			percent = weeklyUsed
		}
	}

	result.Percent = percent
	result.ResetAt = minimaxResetAt(m.RemainsTimeMs)

	if m.CurrentIntervalTotalCount > 0 {
		result.Used = m.CurrentIntervalTotalCount * fiveUsed / 100
		result.Total = m.CurrentIntervalTotalCount
		result.Remaining = fmt.Sprintf("%s / %s (5小时)",
			formatNum(result.Used), formatNum(m.CurrentIntervalTotalCount))
	} else {
		// 总量为 0(如年付套餐未正确 provision 的已知问题)→ 退化为纯百分比展示
		result.Used = fiveUsed
		result.Total = 100
		result.Remaining = fmt.Sprintf("5小时 %.1f%% 已用", fiveUsed)
	}
	switch {
	case weeklyUnlimited:
		result.Remaining += " · 周无限制"
	case weeklyUsed >= 0:
		result.Remaining += fmt.Sprintf(" · 周 %.1f%% 已用", weeklyUsed)
	}
	return result
}

// minimaxBizError 把业务错误码转成用户可读提示(鉴权失败提示 Key 类型与区域)。
func minimaxBizError(base minimaxBaseResp) string {
	if base.StatusCode == 1004 || base.StatusCode == 2049 {
		return fmt.Sprintf("Key 无效或区域不匹配(%d: %s)——请确认使用 sk-cp- 开头的 Token Plan 订阅 Key,而非按量计费 Key",
			base.StatusCode, base.StatusMsg)
	}
	return fmt.Sprintf("API 错误(%d): %s", base.StatusCode, base.StatusMsg)
}

// minimaxResetAt 由 5h 窗口重置倒计时(毫秒)计算重置时间。
func minimaxResetAt(ms float64) string {
	if ms <= 0 {
		return ""
	}
	return time.Now().Add(time.Duration(ms) * time.Millisecond).UTC().Format(time.RFC3339)
}

// fetchOne 调用域名链中一个主机的 remains 端点,返回 200 响应体。
// 失败不区分类型,由 Fetch 主循环统一换下一个主机。
func (f *MiniMaxFetcher) fetchOne(baseURL string) ([]byte, *minimaxFail) {
	url := strings.TrimRight(baseURL, "/") + minimaxRemainsPath
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &minimaxFail{msg: fmt.Sprintf("创建请求失败: %v", err)}
	}
	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://platform.minimaxi.com/")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &minimaxFail{msg: fmt.Sprintf("请求失败: %v", err)}
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 404:
		return nil, &minimaxFail{msg: fmt.Sprintf("Key 无效或端点不可用(HTTP %d)", resp.StatusCode)}
	case resp.StatusCode != 200:
		return nil, &minimaxFail{msg: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &minimaxFail{msg: fmt.Sprintf("读取响应失败: %v", err)}
	}

	// HTTP 200 也可能携带业务鉴权错误(如 Key 类型/区域不对)→ 同样换下一个主机
	var probe struct {
		BaseResp minimaxBaseResp `json:"base_resp"`
	}
	if json.Unmarshal(data, &probe) == nil {
		if probe.BaseResp.StatusCode == 1004 || probe.BaseResp.StatusCode == 2049 {
			return nil, &minimaxFail{msg: minimaxBizError(probe.BaseResp)}
		}
	}
	return data, nil
}
