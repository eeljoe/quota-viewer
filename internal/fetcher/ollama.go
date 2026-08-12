package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// OllamaFetcher 通过抓取 Ollama Cloud settings 页面获取 Session(5 小时窗口)和
// Weekly(7 天窗口)额度百分比。
// 端点: GET https://ollama.com/settings
// 认证: Cookie 头(包含 wos-session / __Secure-session 等会话 Cookie)
// 备注: Ollama 官方未提供 quota API(https://github.com/ollama/ollama/issues/15132),
//
//	目前只能解析 server-rendered HTML。页面结构变更时需更新本文件。
type OllamaFetcher struct {
	cookie  string
	baseURL string // 可重写,用于测试
}

// NewOllamaFetcher 创建一个新的 OllamaFetcher。
func NewOllamaFetcher(cookie string) *OllamaFetcher {
	cookie = strings.TrimSpace(cookie)
	if cookie != "" && !strings.Contains(cookie, "=") {
		cookie = "__Secure-session=" + cookie
	}
	return &OllamaFetcher{
		cookie:  cookie,
		baseURL: "https://ollama.com",
	}
}

type ollamaUsageWindow struct {
	found   bool
	label   string
	percent float64
	resetAt string
}

var (
	ollamaPercentUsedRe  = regexp.MustCompile(`(?i)([0-9]+(?:\.[0-9]+)?)\s*%\s*used`)
	ollamaWidthPercentRe = regexp.MustCompile(`(?i)width:\s*([0-9]+(?:\.[0-9]+)?)%`)
	ollamaDataTimeRe     = regexp.MustCompile(`data-time=["']([^"']+)["']`)
)

var ollamaUsageLabels = []string{"Session usage", "Hourly usage", "Weekly usage"}

func (f *OllamaFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "Ollama",
		Total:       100,
		LastUpdated: time.Now(),
	}

	if f.cookie == "" {
		result.Error = "未配置 Ollama Cookie"
		return result
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := strings.TrimRight(f.baseURL, "/") + "/settings"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

	req.Header.Set("Cookie", f.cookie)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://ollama.com")
	req.Header.Set("Referer", "https://ollama.com/settings")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		result.Error = "Cookie 已过期或无效,请重新登录 ollama.com 后更新"
		return result
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		result.Error = "凭据无效"
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
	html := string(body)

	session := parseOllamaUsageBlock(html, []string{"Session usage", "Hourly usage"})
	weekly := parseOllamaUsageBlock(html, []string{"Weekly usage"})
	if !session.found {
		if looksLikeOllamaSignIn(html) {
			result.Error = "Cookie 已过期或无效,请重新登录 ollama.com 后更新"
		} else {
			result.Error = "未找到 Ollama 5 小时窗口用量,页面结构可能已变化"
		}
		return result
	}

	// 用户关心的是 Session(5 小时)窗口:它驱动悬浮球颜色和主进度条。
	result.Used = session.percent
	result.Percent = session.percent
	result.ResetAt = session.resetAt
	result.Remaining = fmt.Sprintf("5小时 %.1f%% 已用", session.percent)
	if weekly.found {
		result.Remaining += fmt.Sprintf(" · 周 %.1f%% 已用", weekly.percent)
	}
	return result
}

// parseOllamaUsageBlock 按标题切出最多 4000 字节的窗口,再提取百分比与 data-time。
// 当前页面使用 "X% used";width 百分比用于兼容只保留进度条样式的页面。
func parseOllamaUsageBlock(html string, labels []string) ollamaUsageWindow {
	for _, label := range labels {
		start := strings.Index(html, label)
		if start < 0 {
			continue
		}
		tail := html[start+len(label):]
		window := ollamaUsageBlockWindow(tail, label)
		percent, ok := parseOllamaPercent(window)
		if !ok {
			continue
		}
		return ollamaUsageWindow{
			found:   true,
			label:   label,
			percent: percent,
			resetAt: firstOllamaCapture(window, ollamaDataTimeRe),
		}
	}
	return ollamaUsageWindow{}
}

func ollamaUsageBlockWindow(tail, currentLabel string) string {
	end := len(tail)
	if end > 4000 {
		end = 4000
	}
	for _, label := range ollamaUsageLabels {
		if label == currentLabel {
			continue
		}
		if idx := strings.Index(tail, label); idx >= 0 && idx < end {
			end = idx
		}
	}
	return tail[:end]
}

func parseOllamaPercent(block string) (float64, bool) {
	for _, re := range []*regexp.Regexp{ollamaPercentUsedRe, ollamaWidthPercentRe} {
		raw := firstOllamaCapture(block, re)
		if raw == "" {
			continue
		}
		percent, err := strconv.ParseFloat(raw, 64)
		if err == nil {
			return percent, true
		}
	}
	return 0, false
}

func firstOllamaCapture(text string, re *regexp.Regexp) string {
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func looksLikeOllamaSignIn(html string) bool {
	lower := strings.ToLower(html)
	hasHeading := strings.Contains(lower, "sign in to ollama") || strings.Contains(lower, "log in to ollama")
	hasForm := strings.Contains(lower, "<form")
	hasAuthRoute := strings.Contains(lower, "/api/auth/signin") ||
		strings.Contains(lower, "/auth/signin") ||
		strings.Contains(lower, "action=\"/login\"") ||
		strings.Contains(lower, "action=\"/signin\"")
	hasCredentialField := strings.Contains(lower, "type=\"email\"") || strings.Contains(lower, "type=\"password\"")
	return hasForm && (hasHeading || hasAuthRoute || hasCredentialField)
}
