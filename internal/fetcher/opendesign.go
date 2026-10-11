package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// OpenDesignFetcher 通过 open-design.ai 的 AMR Link 中转控制面查询编程套餐额度。
//
// 端点(两步,均为 GET,认证 = Cookie 头,即用户浏览器会话 Cookie):
//  1. GET /api/v1/workspaces/current                                → 取 workspaceId
//  2. GET /api/v1/workspaces/{workspaceId}/billing/coding-plan-usage → 取窗口用量
//
// 坑位(实探确认):
//   - 第二步不能拿 current 当 id 用(会 400 workspace_member_required),必须先取真实 id;
//   - 路径前缀是控制面的 /api/v1,不是数据面 amr-link.open-design.ai 的 /v1;
//   - API key(sk-amr-)读不到任何额度,控制面一律 401,只能用 Cookie;
//   - 不要接钱包余额接口,这里只要 5 小时 + 7 天两个窗口。
//
// 响应 windows[] 按 durationSeconds 识别窗口(18000 = 5 小时,604800 = 7 天),
// 不依赖数组下标(窗口集合归后端管,官方前端同样按 durationSeconds 匹配);
// usedCredits / limitCredits / remainingCredits 是十进制字符串,按 ParseFloat 解析成
// 百分比(credits 计数可能超出 float64 精确整数范围,不转 int64)。
//
// 该路由是官方前端的私有接口,无稳定性承诺;失效时对照 open-design.ai 的 web bundle
// 更新本文件。
type OpenDesignFetcher struct {
	cookie  string
	baseURL string // 可重写,用于测试
}

// opendesignWorkspacePath 是 workspace 发现路由。
const opendesignWorkspacePath = "/api/v1/workspaces/current"

// 窗口时长标识(durationSeconds):5 小时主窗口 + 7 天窗口。
const (
	opendesignWindow5hSeconds = 18000
	opendesignWindow7dSeconds = 604800
)

// NewOpenDesignFetcher 创建一个新的 OpenDesignFetcher。
func NewOpenDesignFetcher(cookie string) *OpenDesignFetcher {
	return &OpenDesignFetcher{
		cookie:  strings.TrimSpace(cookie),
		baseURL: "https://amr-api.open-design.ai",
	}
}

// opendesignWorkspaceResp 对应 /workspaces/current 响应(只取用到的字段)。
type opendesignWorkspaceResp struct {
	WorkspaceID string `json:"workspaceId"`
}

// opendesignUsageWindow 是单个额度窗口。credits 为十进制字符串(按百分比使用,
// 不做整数换算);resetsAt 是 RFC3339 UTC,官方给到毫秒精度(如 .000Z)。
type opendesignUsageWindow struct {
	DurationSeconds float64 `json:"durationSeconds"`
	UsedCredits     string  `json:"usedCredits"`
	LimitCredits    string  `json:"limitCredits"`
	ResetsAt        string  `json:"resetsAt"`
}

// opendesignUsageResp 对应 coding-plan-usage 的成功响应。
// eligible / tier 仅为响应结构文档化,不参与任何逻辑(缺字段或取值变化都不影响展示)。
type opendesignUsageResp struct {
	WorkspaceID string                  `json:"workspaceId"`
	Eligible    bool                    `json:"eligible"`
	Tier        string                  `json:"tier"`
	Windows     []opendesignUsageWindow `json:"windows"`
}

// Fetch 先发现 workspaceId,再查两个窗口的用量。
// 5 小时窗口驱动 Used / ResetAt / Remaining 首项;
// Percent 取 5 小时与 7 天窗口中更紧张的一个(窗口告警契约:任一耗尽都要报警)。
func (f *OpenDesignFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "Open Design",
		Kind:        KindUsage,
		Total:       100,
		LastUpdated: time.Now(),
	}

	if f.cookie == "" {
		result.Error = "未配置 Open Design Cookie"
		return result
	}

	discoveryBody, err := f.get(opendesignWorkspacePath)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	var ws opendesignWorkspaceResp
	if err := json.Unmarshal(discoveryBody, &ws); err != nil {
		result.Error = "解析 workspace 响应失败(响应结构可能已变化): " + err.Error()
		return result
	}
	workspaceID := strings.TrimSpace(ws.WorkspaceID)
	if workspaceID == "" {
		result.Error = "未取到 workspaceId,Cookie 可能已失效,请重新登录 open-design.ai 后更新"
		return result
	}

	usageBody, err := f.get("/api/v1/workspaces/" + workspaceID + "/billing/coding-plan-usage")
	if err != nil {
		result.Error = err.Error()
		return result
	}

	var usage opendesignUsageResp
	if err := json.Unmarshal(usageBody, &usage); err != nil {
		result.Error = "解析响应失败(响应结构可能已变化): " + err.Error()
		return result
	}

	return opendesignParseWindows(result, &usage)
}

// opendesignParseWindows 按 durationSeconds 挑出 5 小时与 7 天窗口并填充结果。
// 缺 7 天窗口(或它不可解析)时只写 5 小时段;缺 5 小时窗口则无法驱动主展示,报错。
func opendesignParseWindows(result QuotaResult, usage *opendesignUsageResp) QuotaResult {
	var fiveHour, weekly *opendesignUsageWindow
	for i := range usage.Windows {
		switch usage.Windows[i].DurationSeconds {
		case opendesignWindow5hSeconds:
			if fiveHour == nil {
				fiveHour = &usage.Windows[i]
			}
		case opendesignWindow7dSeconds:
			if weekly == nil {
				weekly = &usage.Windows[i]
			}
		}
	}

	if fiveHour == nil {
		result.Error = "未找到 5 小时窗口用量(响应结构可能已变化)"
		return result
	}
	fiveHourUsed, ok := opendesignUsedPercent(fiveHour)
	if !ok {
		result.Error = "5 小时窗口用量字段不可解析(响应结构可能已变化)"
		return result
	}

	result.Used = fiveHourUsed
	result.Percent = fiveHourUsed
	result.ResetAt = opendesignResetAt(fiveHour.ResetsAt)
	result.Remaining = fmt.Sprintf("5小时 %.1f%% 已用", fiveHourUsed)

	if weekly != nil {
		if weeklyUsed, ok := opendesignUsedPercent(weekly); ok {
			if weeklyUsed > result.Percent {
				result.Percent = weeklyUsed
			}
			result.Remaining += fmt.Sprintf(" · 周 %.1f%% 已用", weeklyUsed)
		}
	}
	return result
}

// opendesignUsedPercent 由窗口的已用/总量 credits 字符串算出已用百分比。
func opendesignUsedPercent(w *opendesignUsageWindow) (float64, bool) {
	used, err := strconv.ParseFloat(strings.TrimSpace(w.UsedCredits), 64)
	if err != nil {
		return 0, false
	}
	limit, err := strconv.ParseFloat(strings.TrimSpace(w.LimitCredits), 64)
	if err != nil || limit <= 0 {
		return 0, false
	}
	return clampPercent(used / limit * 100), true
}

// opendesignResetAt 把窗口的 resetsAt 规范化为项目惯用的 UTC 秒精度 RFC3339
// (官方给到毫秒精度,如 2026-10-11T09:35:11.000Z);无法解析时原样返回。
func opendesignResetAt(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	ts, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}
	return ts.UTC().Format(time.RFC3339)
}

// get 发起一次带 Cookie 头的 GET,返回 200 的响应体。
// Cookie 失效时控制面可能返回 401(API key 或过期 Cookie)或 302 跳转登录页,
// 两者统一按 Cookie 失效提示;因此必须禁止自动重定向,否则 302 会被跟随成
// 200 登录页 HTML,报成「解析响应失败」而丢掉「请重新登录」指引。
func (f *OpenDesignFetcher) get(path string) ([]byte, error) {
	url := strings.TrimRight(f.baseURL, "/") + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Cookie", f.cookie)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 302:
		return nil, fmt.Errorf("Cookie 已过期或无效,请重新登录 open-design.ai/cloud/dashboard 后更新")
	case resp.StatusCode != 200:
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	return data, nil
}
