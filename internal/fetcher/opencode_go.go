package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OpenCodeGoFetcher 通过官方 JSON API 获取 OpenCode Go 套餐的窗口额度百分比。
// 端点: GET https://opencode.ai/zen/go/v1/usage
// 认证: Authorization: Bearer <API Key>
type OpenCodeGoFetcher struct {
	apiKey       string
	baseURL      string // 可重写,用于测试
	readAuthFile bool   // apiKey 为空时是否尝试读取 OpenCode CLI 的 auth.json
	authFile     string // 覆盖 auth.json 默认路径(测试注入)
}

// NewOpenCodeGoFetcher 创建一个新的 OpenCodeGoFetcher。
// apiKey 为空时,凭证会在 Fetch 里回退到 OpenCode CLI 保存的 auth.json。
func NewOpenCodeGoFetcher(apiKey string) *OpenCodeGoFetcher {
	return &OpenCodeGoFetcher{
		apiKey:       apiKey,
		baseURL:      "https://opencode.ai",
		readAuthFile: true,
	}
}

// windowInfo 描述一个配额窗口。
type windowInfo struct {
	windowType   string  // "rolling", "weekly", "monthly"
	usagePercent float64 // API 直接返回的百分比(0-100)
	resetsAt     string  // 原始 resetsAt(可能为空)
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

// openCodeGoWindow 是 /zen/go/v1/usage 响应里单个窗口的结构。
// percent 是 0-100 的百分比单位(1 = 1%),resetsAt 是绝对时间(可能为 null/缺失)。
type openCodeGoWindow struct {
	Status   string   `json:"status"` // 语义未知,不用于过滤窗口
	Percent  *float64 `json:"percent"`
	ResetsAt *string  `json:"resetsAt"`
}

type openCodeGoUsageResp struct {
	Usage *struct {
		Rolling *openCodeGoWindow `json:"rolling"`
		Weekly  *openCodeGoWindow `json:"weekly"`
		Monthly *openCodeGoWindow `json:"monthly"`
	} `json:"usage"`
}

// openCodeGoErrorResp 是错误响应里的 error.type 结构(用于区分未订阅)。
type openCodeGoErrorResp struct {
	Error *struct {
		Type string `json:"type"`
	} `json:"error"`
}

// openCodeGoAuthFile 是 OpenCode CLI 的 ~/.local/share/opencode/auth.json 结构:
// {"opencode-go":{"type":"api","key":"..."}} 之类。为容错只取 key 字段。
type openCodeGoAuthFile map[string]struct {
	Key string `json:"key"`
}

// Fetch 拉取并解析 OpenCode Go 额度。
// 选窗:Percent 取所有有值窗口中最大的(最紧张),同值时 rolling > weekly > monthly。
func (f *OpenCodeGoFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "OpenCode Go",
		Total:       100,
		LastUpdated: time.Now(),
	}

	key := f.apiKey
	if key == "" && f.readAuthFile {
		key = f.readAuthKey()
	}
	if key == "" {
		result.Error = "未配置 OpenCode Go API Key(在 opencode.ai 订阅后复制 API Key 填入,或在 opencode CLI 里 /connect 后自动读取)"
		return result
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", f.baseURL+"/zen/go/v1/usage", nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("User-Agent", "quota-viewer/1.0")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result
	}

	switch resp.StatusCode {
	case 401:
		result.Error = "API Key 无效"
		return result
	case 403:
		var e openCodeGoErrorResp
		if json.Unmarshal(body, &e) == nil && e.Error != nil && e.Error.Type == "EntitlementError" {
			result.Error = "未订阅 OpenCode Go 套餐(或订阅已失效)"
			return result
		}
		result.Error = "API Key 无权限"
		return result
	case 200:
		// 继续解析
	default:
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return result
	}

	var parsed openCodeGoUsageResp
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Usage == nil {
		result.Error = "无法解析额度数据(响应结构可能已变化)"
		return result
	}

	windows := []windowInfo{}
	for _, w := range []struct {
		wt     string
		window *openCodeGoWindow
	}{
		{"rolling", parsed.Usage.Rolling},
		{"weekly", parsed.Usage.Weekly},
		{"monthly", parsed.Usage.Monthly},
	} {
		if w.window == nil || w.window.Percent == nil {
			continue
		}
		info := windowInfo{windowType: w.wt, usagePercent: *w.window.Percent}
		if w.window.ResetsAt != nil {
			info.resetsAt = *w.window.ResetsAt
		}
		windows = append(windows, info)
	}

	if len(windows) == 0 {
		result.Error = "未找到有效配额窗口"
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

	result.Percent = best.usagePercent
	result.Remaining = fmt.Sprintf("%s · 已用 %s · 剩余 %s", windowLabel(best.windowType),
		formatUsagePercent(best.usagePercent), formatUsagePercent(100-best.usagePercent))
	if t, err := time.Parse(time.RFC3339Nano, best.resetsAt); err == nil {
		result.ResetAt = t.UTC().Format(time.RFC3339)
	}

	return result
}

// readAuthKey 从 OpenCode CLI 的凭证文件读取 opencode-go 的 key。
// 文件缺失、结构不符或没有该 provider 时返回空字符串(调用方回退到"未配置"错误)。
func (f *OpenCodeGoFetcher) readAuthKey() string {
	path := f.authFile
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		path = filepath.Join(home, ".local", "share", "opencode", "auth.json")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var auth openCodeGoAuthFile
	if json.Unmarshal(data, &auth) != nil {
		return ""
	}
	return strings.TrimSpace(auth["opencode-go"].Key)
}

// formatUsagePercent 把浮点百分比格式化为紧凑文本:整数值去小数(43 → "43%"),
// 非整数保留一位小数(3.1 → "3.1%")。
func formatUsagePercent(p float64) string {
	if p == float64(int(p)) {
		return fmt.Sprintf("%d%%", int(p))
	}
	return fmt.Sprintf("%.1f%%", p)
}
