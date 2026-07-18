package config

import (
	"regexp"
	"strings"
)

// psCookieLineRe 匹配 Chrome/Edge "复制为 PowerShell" 里的 Cookie 定义:
// $session.Cookies.Add((New-Object System.Net.Cookie("name", "value", "/", "domain")))
// value 是 PowerShell 双引号字符串,内部可能含反引号转义(`" 表示字面引号)。
var psCookieLineRe = regexp.MustCompile("System\\.Net\\.Cookie\\(\\s*\"([^\"]+)\"\\s*,\\s*\"((?:`.|[^\"`])*)\"\\s*,")

// NormalizeCookieInput 规范化设置页粘贴的 Cookie:
//  1. 浏览器 "Copy as PowerShell" 整段脚本 -> 提取为 Cookie 请求头格式 "k1=v1; k2=v2"
//  2. 已是 Cookie 头格式 -> 原样返回(仅去首尾空白)
func NormalizeCookieInput(s string) string {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "System.Net.Cookie") {
		return s
	}
	matches := psCookieLineRe.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return s
	}
	seen := make(map[string]bool, len(matches))
	pairs := make([]string, 0, len(matches))
	for _, m := range matches {
		name := m[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		pairs = append(pairs, name+"="+unescapePSString(m[2]))
	}
	return strings.Join(pairs, "; ")
}

// unescapePSString 还原 PowerShell 双引号字符串内的反引号转义。
func unescapePSString(s string) string {
	const sentinel = "\x00"
	s = strings.ReplaceAll(s, "``", sentinel)
	s = strings.ReplaceAll(s, "`\"", "\"")
	s = strings.ReplaceAll(s, "`n", "\n")
	s = strings.ReplaceAll(s, "`r", "")
	s = strings.ReplaceAll(s, "`t", "\t")
	return strings.ReplaceAll(s, sentinel, "`")
}
