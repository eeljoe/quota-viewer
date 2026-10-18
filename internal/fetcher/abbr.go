package fetcher

import (
	"fmt"
	"strings"
	"unicode"
)

// 球格缩写派生工具:多套餐 / 多 Key 场景下,一个 Provider 会产出多个监控对象,
// 每个对象都要有自己的球格缩写,否则同屏出现几个一模一样的格子,用户无从分辨。
//
// 派生规则(刻意保守,优先保证同一平台内可区分):
//   - CJK 名称取首字("视频赠送卡" → "视","音乐卡" → "音")
//   - ASCII 名称取前两个字母或数字并大写("MiniMax-M3" → "M3","video_pack" → "VI")
//   - 无法派生时由调用方退回序号缩写("P1" / "P2")

// SettleIdentity 归一一个监控对象的身份字段(app.go 编排层与 probe 探针共用,
// 两处必须同口径,否则球面与 --json 输出对不上)。
//
// 参数:
//   - credNo   该对象属于第几组凭证(0 = 主凭证)
//   - planIdx  该对象是该凭证返回的第几张套餐卡(0 = 主套餐)
//   - planCount 这组凭证共产出几张套餐卡
//   - credCount 该 Provider 共几组凭证
//
// 规则:
//   - 单对象(1 组凭证 × 1 张卡):清空后缀、沿用 Provider 缩写 —— 与旧观感完全一致
//   - 多对象:主对象保留 Provider 缩写;套餐卡用卡名派生缩写;第 2..N 组凭证
//     用「Provider 首字母 + Key 序号」(如 M1 / M2)区分
func SettleIdentity(r *QuotaResult, def ProviderDef, credNo, planIdx, planCount, credCount int) {
	r.Slot = fmt.Sprintf("%s/%d", def.ID, credNo)
	if planCount > 1 {
		r.Slot = fmt.Sprintf("%s/%d/p%d", def.ID, credNo, planIdx)
	}

	// 单对象:后缀是噪音,清掉。
	if credCount <= 1 && planCount <= 1 {
		r.SubLabel = ""
		r.Abbr = def.Abbr
		return
	}

	// 面板后缀:多 Key 时标出是第几把;多套餐时保留 fetcher 给的套餐名。
	switch {
	case credCount > 1 && planCount > 1:
		key := fmt.Sprintf("Key %d", credNo+1)
		if r.SubLabel != "" {
			r.SubLabel = key + " · " + r.SubLabel
		} else {
			r.SubLabel = key
		}
	case credCount > 1:
		r.SubLabel = fmt.Sprintf("Key %d", credNo+1)
	}

	// 球格缩写:每格最多两个字符(球只有 60px)。
	switch {
	case r.Abbr != "":
		// fetcher 已按套餐卡名派生出缩写(如 "视" / "M3"),直接用。
	case credNo > 0 && planCount > 1:
		// 多 Key 且多套餐:两维都要区分,退回 Provider 首字母 + Key 序号。
		r.Abbr = keyAbbr(def.Abbr, credNo)
	case credNo > 0:
		r.Abbr = keyAbbr(def.Abbr, credNo)
	default:
		r.Abbr = def.Abbr
	}
}

// keyAbbr 用 Provider 缩写的首字母 + Key 序号组成两字符缩写("MiniMax" 的第 2 把 Key → "M2")。
func keyAbbr(base string, credNo int) string {
	r := []rune(strings.TrimSpace(base))
	head := "?"
	if len(r) > 0 {
		head = string(r[0])
	}
	if head == "?" {
		return fmt.Sprintf("K%d", credNo+1)
	}
	return fmt.Sprintf("%s%d", head, credNo+1)
}

// shortPlanAbbr 从套餐卡名派生两字以内的缩写;派生不出来时返回 ""。
func shortPlanAbbr(name string, _ int) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	// CJK 首字:汉字直接可读,取一个字最省球格空间。
	for _, r := range name {
		if unicode.Is(unicode.Han, r) {
			return string(r)
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			break // 非 CJK 的字母数字走 ASCII 分支
		}
	}
	// ASCII:取前两个字母或数字。
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToUpper(r))
			if b.Len() >= 2 {
				break
			}
		}
	}
	return b.String()
}

// planAbbrSetter 为一个 Provider 的多张套餐卡分配互不重复的球格缩写。
type planAbbrSetter struct {
	used map[string]bool
}

// next 为第 idx 张卡(0-based)取缩写:能用卡名派生且未被占用就用派生值,
// 否则退回 "P{idx+1}" 序号缩写,确保同屏不会出现两个相同的格子。
func (s *planAbbrSetter) next(idx int, name string) string {
	if s.used == nil {
		s.used = map[string]bool{}
	}
	if idx == 0 {
		s.used[""] = true // 主套餐保留 Provider 缩写,先占位避免后续卡片撞上
		return ""
	}
	if a := shortPlanAbbr(name, idx); a != "" && !s.used[a] {
		s.used[a] = true
		return a
	}
	a := fmt.Sprintf("P%d", idx+1)
	for s.used[a] {
		idx++
		a = fmt.Sprintf("P%d", idx+1)
	}
	s.used[a] = true
	return a
}
