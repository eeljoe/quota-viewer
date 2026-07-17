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

// MiMoFetcher 通过 Cookie 抓取小米 MiMo 额度页面并解析 HTML DOM。
// 小米 MiMo 平台额度数据为服务端渲染,直接出现在 plan-manage 页面 HTML 中。
type MiMoFetcher struct {
	cookie string
	apiURL string
}

func NewMiMoFetcher(cookie string, apiURL string) *MiMoFetcher {
	if apiURL == "" {
		apiURL = "https://platform.xiaomimimo.com/console/plan-manage"
	}
	return &MiMoFetcher{cookie: cookie, apiURL: apiURL}
}

// "8,239,030,362 / 38,000,000,000" 形式的已用/总量。
var mimoUsedTotalRe = regexp.MustCompile(`([\d,]+)\s*/\s*([\d,]+)`)

// "已使用 22.0%" 形式的百分比。
var mimoPercentRe = regexp.MustCompile(`已使用\s*([\d.]+)%`)

func (m *MiMoFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "小米MiMo",
		LastUpdated: time.Now(),
	}

	if m.cookie == "" {
		result.Error = "未配置 MiMo Cookie"
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

	req, err := http.NewRequest("GET", m.apiURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

	req.Header.Set("Cookie", m.cookie)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", "https://platform.xiaomimimo.com/")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 302 {
		result.Error = "Cookie 已过期,请更新"
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

	// 页面可能有多处 "数字 / 数字" 形式;优先匹配较大的 Credits 数值。
	// 取第一个匹配到的 used/total 组合。
	var used, total float64
	parsed := false
	for _, m := range mimoUsedTotalRe.FindAllStringSubmatch(html, -1) {
		u, errU := parseMimoNumber(m[1])
		t, errT := parseMimoNumber(m[2])
		if errU != nil || errT != nil {
			continue
		}
		// 选取总量 > 0 且数值较大的一组(Credits 量级通常在上亿)
		if t > 0 && t > total {
			used, total = u, t
			parsed = true
		}
	}

	if !parsed {
		// 回退:只匹配百分比
		if pm := mimoPercentRe.FindStringSubmatch(html); len(pm) >= 2 {
			if p, err := strconv.ParseFloat(pm[1], 64); err == nil {
				result.Percent = p
				result.Remaining = fmt.Sprintf("已使用 %.1f%% Credits", p)
				return result
			}
		}
		result.Error = "HTML 中未找到 used/total 或百分比,页面结构可能已变更"
		return result
	}

	result.Used = used
	result.Total = total
	if total > 0 {
		result.Percent = used / total * 100
	}

	// 若 HTML 中也有显式百分比且与计算值偏差较大,优先用显式值
	if pm := mimoPercentRe.FindStringSubmatch(html); len(pm) >= 2 {
		if p, err := strconv.ParseFloat(pm[1], 64); err == nil {
			result.Percent = p
		}
	}

	result.Remaining = fmt.Sprintf("%.0f / %.0f Credits", used, total)

	return result
}

// parseMimoNumber 解析可能含逗号的数字(如 "8,239,030,362")。
func parseMimoNumber(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	return strconv.ParseFloat(s, 64)
}
