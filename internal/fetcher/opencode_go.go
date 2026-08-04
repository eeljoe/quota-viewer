package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// OpenCodeGoFetcher 通过抓取 OpenCode Dashboard 页面获取 Go 套餐的滚动窗口额度百分比。
// 端点: GET https://opencode.ai/workspace/{workspaceID}/go
// 认证: auth Cookie(与 Xfyun/MiMo 相同通过 Cookie 头传递)
type OpenCodeGoFetcher struct {
	workspaceID  string
	sessionToken string
	baseURL      string // 可重写,用于测试
}

// NewOpenCodeGoFetcher 创建一个新的 OpenCodeGoFetcher。
func NewOpenCodeGoFetcher(workspaceID string, sessionToken string) *OpenCodeGoFetcher {
	return &OpenCodeGoFetcher{
		workspaceID:  workspaceID,
		sessionToken: sessionToken,
		baseURL:      "https://opencode.ai",
	}
}

// windowInfo 描述一个配额窗口。
type windowInfo struct {
	windowType   string // "rolling", "weekly", "monthly"
	usagePercent int
	resetInSec   int
}

// windowTypeOrder 用于 tiebreaker: 相同百分比时 rolling > weekly > monthly。
var windowTypeOrder = map[string]int{
	"rolling": 0,
	"weekly":  1,
	"monthly": 2,
}

// windowLabel 窗口类型的中文描述。
func windowLabel(wt string) string {
	switch wt {
	case "rolling":
		return "5小时窗口"
	case "weekly":
		return "周窗口"
	case "monthly":
		return "月窗口"
	}
	return wt
}

// ssrPattern 匹配内嵌在 HTML 中的 SSR hydration 配额数据。
// 格式: rollingUsage:$R[10]={usagePercent:7,resetInSec:18000}
var ssrPattern = regexp.MustCompile(`(rolling|weekly|monthly)Usage:\$R\[\d+\]=\{([^}]+)\}`)

// slotPattern 是 data-slot HTML 备选解析模式。
var slotPattern = regexp.MustCompile(`<div data-slot="usage-item">\s*<span data-slot="usage-label">(Rolling Usage|Weekly Usage|Monthly Usage)</span>\s*<span data-slot="usage-value"><!--\$-->(\d+)<!--/-->%</span>\s*</div>`)

// slotNameMap 将 data-slot label 映射到内部窗口类型。
var slotNameMap = map[string]string{
	"Rolling Usage": "rolling",
	"Weekly Usage":  "weekly",
	"Monthly Usage": "monthly",
}

func (f *OpenCodeGoFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "OpenCode Go",
		Total:       100,
		LastUpdated: time.Now(),
	}

	if f.sessionToken == "" {
		result.Error = "未配置 OpenCode Go Cookie"
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

	url := fmt.Sprintf("%s/workspace/%s/go", f.baseURL, f.workspaceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

	req.Header.Set("Cookie", "auth="+f.sessionToken)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	// 错误分类
	switch resp.StatusCode {
	case 302, 303:
		result.Error = "会话已过期"
		return result
	case 401, 403:
		result.Error = "凭据无效"
		return result
	case 404:
		result.Error = "Workspace 不存在"
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

	// 1) 尝试 SSR hydration 数据
	windows := parseSSRWindows(string(body))
	if len(windows) == 0 {
		// SSR 模式匹配成功但所有窗口 usagePercent=0 → 无有效窗口
		if ssrPattern.MatchString(string(body)) {
			result.Error = "未找到有效配额窗口"
			return result
		}
		// 2) 回退到 data-slot HTML 解析
		windows = parseSlotWindows(string(body))
	}

	if len(windows) == 0 {
		result.Error = "页面结构已变化"
		return result
	}

	// 选取 usagePercent 最高的窗口; 相同则 rolling > weekly > monthly
	best := windows[0]
	for _, w := range windows[1:] {
		if w.usagePercent > best.usagePercent ||
			(w.usagePercent == best.usagePercent && windowTypeOrder[w.windowType] < windowTypeOrder[best.windowType]) {
			best = w
		}
	}

	result.Percent = float64(best.usagePercent)
	result.Remaining = fmt.Sprintf("%s · 已用 %d%% · 剩余 %d%%",
		windowLabel(best.windowType), best.usagePercent, 100-best.usagePercent)
	result.ResetAt = time.Now().Add(time.Duration(best.resetInSec) * time.Second).Format(time.RFC3339)

	return result
}

// parseSSRWindows 从 HTML 中解析 SSR hydration 配额窗口数据。
func parseSSRWindows(html string) []windowInfo {
	matches := ssrPattern.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil
	}

	var windows []windowInfo
	for _, m := range matches {
		wt := m[1]
		inner := m[2]
		percent, resetSec := parseSSRFields(inner)
		if percent > 0 {
			windows = append(windows, windowInfo{
				windowType:   wt,
				usagePercent: percent,
				resetInSec:   resetSec,
			})
		}
	}
	return windows
}

// parseSSRFields 解析 SSR 窗口数据内部的 key=value 对。
// 内部格式: usagePercent:7,resetInSec:18000 (顺序不固定)
func parseSSRFields(inner string) (percent int, resetSec int) {
	for _, part := range strings.Split(inner, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "usagePercent:") {
			percent, _ = strconv.Atoi(strings.TrimPrefix(part, "usagePercent:"))
		} else if strings.HasPrefix(part, "resetInSec:") {
			resetSec, _ = strconv.Atoi(strings.TrimPrefix(part, "resetInSec:"))
		}
	}
	return
}

// parseSlotWindows 从 HTML data-slot 结构中解析配额窗口数据(SSR 解析失败时的备选)。
func parseSlotWindows(html string) []windowInfo {
	matches := slotPattern.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil
	}

	var windows []windowInfo
	for _, m := range matches {
		label := m[1]
		percentStr := m[2]
		percent, _ := strconv.Atoi(percentStr)
		if percent <= 0 {
			continue
		}
		wt, ok := slotNameMap[label]
		if !ok {
			continue
		}
		// data-slot 没有 resetInSec, 默认 0 表示未知
		windows = append(windows, windowInfo{
			windowType:   wt,
			usagePercent: percent,
			resetInSec:   0,
		})
	}
	return windows
}
