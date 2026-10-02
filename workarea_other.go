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

// syncBallRegion 非 Windows 平台无需实现(窗口形状由系统管理)。
func syncBallRegion() {}

// ForceBallWindowSize 非 Windows 平台无需实现(无系统最小宽度问题)。
func ForceBallWindowSize() {}
