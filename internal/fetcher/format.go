package fetcher

import "strconv"

// formatNum 以千分位分隔格式化数字,用于额度展示(如 8,335,815,690)。
func formatNum(v float64) string {
	n := v + 0.5 // 四舍五入到整数(与原 %.0f 展示语义一致)
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(int64(n), 10)
	out := make([]byte, 0, len(s)+len(s)/3+1)
	if neg {
		out = append(out, '-')
	}
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return string(out)
}
