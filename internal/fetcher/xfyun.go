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

// XfyunFetcher 通过 Cookie 抓取讯飞星辰 MaaS 额度页面并解析 HTML DOM。
// 讯飞平台额度数据为服务端渲染,直接出现在 packageSubscription 页面 HTML 中。
type XfyunFetcher struct {
	cookie string
	apiURL string // HTML 页面地址
}

func NewXfyunFetcher(cookie string, apiURL string) *XfyunFetcher {
	if apiURL == "" {
		apiURL = "https://maas.xfyun.cn/packageSubscription"
	}
	return &XfyunFetcher{cookie: cookie, apiURL: apiURL}
}

// usage-used / usage-total / usage-unit 的取值用非贪婪匹配,避免跨标签贪婪。
var (
	xfyunUsedRe   = regexp.MustCompile(`class="usage-used"[^>]*>\s*([\d,]+)\s*<`)
	xfyunTotalRe  = regexp.MustCompile(`class="usage-total"[^>]*>\s*([\d,]+)\s*<`)
	xfyunUnitRe   = regexp.MustCompile(`class="usage-unit"[^>]*>\s*([^<]+?)\s*<`)
	xfyunNameRe   = regexp.MustCompile(`class="package-name"[^>]*>\s*([^<]+?)\s*<`)
	xfyunWidthRe  = regexp.MustCompile(`width:\s*([\d.]+)`)
)

func (x *XfyunFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "讯飞星辰",
		LastUpdated: time.Now(),
	}

	if x.cookie == "" {
		result.Error = "未配置讯飞 Cookie"
		return result
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		// Cookie 过期时讯飞会 302 跳转到登录页,不自动跟随
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", x.apiURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

	req.Header.Set("Cookie", x.cookie)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", "https://maas.xfyun.cn/")

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

	usedMatch := xfyunUsedRe.FindStringSubmatch(html)
	totalMatch := xfyunTotalRe.FindStringSubmatch(html)

	if len(usedMatch) < 2 || len(totalMatch) < 2 {
		result.Error = "HTML 中未找到 usage-used/usage-total,页面结构可能已变更"
		return result
	}

	used, err := parseXfyunNumber(usedMatch[1])
	if err != nil {
		result.Error = fmt.Sprintf("解析 used 失败: %v", err)
		return result
	}
	total, err := parseXfyunNumber(totalMatch[1])
	if err != nil {
		result.Error = fmt.Sprintf("解析 total 失败: %v", err)
		return result
	}

	result.Used = used
	result.Total = total
	if total > 0 {
		result.Percent = used / total * 100
	}

	// 单位与套餐名(可选)
	unit := "次"
	if m := xfyunUnitRe.FindStringSubmatch(html); len(m) >= 2 {
		unit = strings.TrimSpace(m[1])
	}
	name := ""
	if m := xfyunNameRe.FindStringSubmatch(html); len(m) >= 2 {
		name = strings.TrimSpace(m[1])
	}

	// 进度条宽度作为百分比交叉校验(可选,不覆盖主结果)
	if m := xfyunWidthRe.FindStringSubmatch(html); len(m) >= 2 {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil && v > 0 && v <= 1 && result.Percent == 0 {
			result.Percent = v * 100
		}
	}

	if name != "" {
		result.Remaining = fmt.Sprintf("%s: %.0f / %.0f %s", name, used, total, unit)
	} else {
		result.Remaining = fmt.Sprintf("%.0f / %.0f %s", used, total, unit)
	}

	return result
}

// parseXfyunNumber 解析可能含逗号的数字(如 "1,017")。
func parseXfyunNumber(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	return strconv.ParseFloat(s, 64)
}
