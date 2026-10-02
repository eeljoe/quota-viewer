//go:build !windows

package main

// workAreaForPoint 非 Windows 平台无实现,返回 ok=false,调用方走 Wails Screen 近似。
func workAreaForPoint(px, py int) (x, y, w, h, dpi int, ok bool) {
	return 0, 0, 0, 0, 0, false
}

// setupWindowStyles 非 Windows 平台无需实现。
func setupWindowStyles(title string) bool {
	return false
}
