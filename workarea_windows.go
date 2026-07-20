//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// Windows 显示器信息(物理像素坐标)
type winRect struct {
	left, top, right, bottom int32
}

type monitorInfo struct {
	cbSize    uint32
	rcMonitor winRect
	rcWork    winRect
	dwFlags   uint32
}

var (
	winUser32            = syscall.NewLazyDLL("user32.dll")
	winShcore            = syscall.NewLazyDLL("shcore.dll")
	winComctl32          = syscall.NewLazyDLL("comctl32.dll")
	procMonitorFromPoint = winUser32.NewProc("MonitorFromPoint")
	procGetMonitorInfoW  = winUser32.NewProc("GetMonitorInfoW")
	procGetDpiForMonitor = winShcore.NewProc("GetDpiForMonitor")
	procFindWindowW      = winUser32.NewProc("FindWindowW")
	procGetDpiForWindow  = winUser32.NewProc("GetDpiForWindow")
	procSetWindowSub     = winComctl32.NewProc("SetWindowSubclass")
	procDefSubclassProc  = winComctl32.NewProc("DefSubclassProc")

	subclassCB uintptr // 持有回调引用,防止被 GC 回收
)

const (
	monitorDefaultToNearest = 2
	mdtEffectiveDPI         = 0
	wmGetMinMaxInfo         = 0x0024
)

// minTrackSubclassProc 拦截 WM_GETMINMAXINFO:overlapped 窗口有系统默认最小宽度
// (高 DPI 下约 262px 物理,会把 60px 球窗撑宽),这里把最小拖动尺寸压到
// ballSize 对应的物理像素。其余消息走默认子类过程。
func minTrackSubclassProc(hwnd uintptr, msg uint32, wparam uintptr, lparam unsafe.Pointer, uIDSubclass, dwRefData uintptr) uintptr {
	if msg == wmGetMinMaxInfo {
		dpi, _, _ := procGetDpiForWindow.Call(hwnd)
		if dpi == 0 {
			dpi = 96
		}
		minPx := int32(uint32(ballSize) * uint32(dpi) / 96)
		// MINMAXINFO.PtMinTrackSize 位于偏移 24(ptReserved/ptMaxSize/ptMaxPosition 各 8 字节)
		p := (*[2]int32)(unsafe.Add(lparam, 24))
		p[0] = minPx
		p[1] = minPx
		return 0
	}
	ret, _, _ := procDefSubclassProc.Call(hwnd, uintptr(msg), wparam, uintptr(lparam))
	return ret
}

// setupWindowStyles 安装子类覆盖系统默认最小窗口宽度。
// 返回 false 表示未找到窗口(调用方仅记录)。
func setupWindowStyles(title string) bool {
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return false
	}
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		return false
	}

	subclassCB = syscall.NewCallback(minTrackSubclassProc)
	ret, _, _ := procSetWindowSub.Call(hwnd, subclassCB, 1, 0)
	return ret != 0
}

// workAreaForPoint 返回包含点 (px,py)(物理像素)的显示器工作区(不含任务栏)
// 及该屏 DPI,结果均为物理像素。ok=false 表示查询失败,调用方走回退逻辑。
func workAreaForPoint(px, py int) (x, y, w, h, dpi int, ok bool) {
	// POINT 按值传参:64 位下将 x/y 打包进一个 uintptr(低 4 字节 x,高 4 字节 y)
	pt := uintptr(uint64(uint32(int32(px))) | (uint64(uint32(int32(py))) << 32))
	hmon, _, _ := procMonitorFromPoint.Call(pt, monitorDefaultToNearest)
	if hmon == 0 {
		return 0, 0, 0, 0, 0, false
	}

	var mi monitorInfo
	mi.cbSize = uint32(unsafe.Sizeof(mi))
	ret, _, _ := procGetMonitorInfoW.Call(hmon, uintptr(unsafe.Pointer(&mi)))
	if ret == 0 {
		return 0, 0, 0, 0, 0, false
	}

	var dpiX, dpiY uint32
	hr, _, _ := procGetDpiForMonitor.Call(hmon, mdtEffectiveDPI,
		uintptr(unsafe.Pointer(&dpiX)), uintptr(unsafe.Pointer(&dpiY)))
	if hr != 0 || dpiX == 0 {
		dpiX = 96 // 获取失败按 100% 处理
	}

	return int(mi.rcWork.left), int(mi.rcWork.top),
		int(mi.rcWork.right - mi.rcWork.left), int(mi.rcWork.bottom - mi.rcWork.top),
		int(dpiX), true
}
