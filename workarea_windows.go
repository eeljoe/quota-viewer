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
	winGdi32             = syscall.NewLazyDLL("gdi32.dll")
	procMonitorFromPoint = winUser32.NewProc("MonitorFromPoint")
	procGetMonitorInfoW  = winUser32.NewProc("GetMonitorInfoW")
	procGetDpiForMonitor = winShcore.NewProc("GetDpiForMonitor")
	procFindWindowW      = winUser32.NewProc("FindWindowW")
	procGetWindowLongW   = winUser32.NewProc("GetWindowLongPtrW")
	procSetWindowLongW   = winUser32.NewProc("SetWindowLongPtrW")
	procSetWindowPos     = winUser32.NewProc("SetWindowPos")
	procGetDpiForWindow  = winUser32.NewProc("GetDpiForWindow")
	procSetWindowSub     = winComctl32.NewProc("SetWindowSubclass")
	procDefSubclassProc  = winComctl32.NewProc("DefSubclassProc")
	procGetWindowRect    = winUser32.NewProc("GetWindowRect")
	procSetWindowRgn     = winUser32.NewProc("SetWindowRgn")
	procCreateElliptic   = winGdi32.NewProc("CreateEllipticRgn")

	subclassCB   uintptr // 持有回调引用,防止被 GC 回收
	mainBallHwnd uintptr // 球窗句柄(region/尺寸强制用)
)

const (
	monitorDefaultToNearest = 2
	mdtEffectiveDPI         = 0

	gwlExStyle = -20

	// 工具窗口:不进任务栏、不出现在 Alt+Tab、没有任务栏缩略图(托盘应用标准做法)
	wsExToolWindow = 0x00000080

	wmGetMinMaxInfo = 0x0024
	wmSize          = 0x0005

	swpNoSize       = 0x0001
	swpNoMove       = 0x0002
	swpNoZOrder     = 0x0004
	swpNoActivate   = 0x0010
	swpFrameChanged = 0x0020
)

// minTrackSubclassProc 拦截 WM_GETMINMAXINFO:overlapped 窗口有系统默认最小宽度
// (高 DPI 下约 262px 物理,会把 60px 球窗撑宽),这里把最小拖动尺寸压到
// ballSize 对应的物理像素。其余消息走默认子类过程。
func minTrackSubclassProc(hwnd uintptr, msg uint32, wparam uintptr, lparam unsafe.Pointer, uIDSubclass, dwRefData uintptr) uintptr {
	switch msg {
	case wmGetMinMaxInfo:
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
	case wmSize:
		// 球态(60x60 物理 = ballPx)时把窗口裁成圆形:圆外逐像素透明由
		// NOREDIRECTIONBITMAP 负责,region 兜底裁掉 DWM 边框/背景板的方形残影,
		// 圆外区域同时天然点击穿透。展开面板(非正方形或非球尺寸)时恢复矩形。
		if wparam != 1 { // 1 = SIZE_MINIMIZED,尺寸无效
			w := int32(int16(uint16(uintptr(lparam) & 0xFFFF)))
			h := int32(int16(uint16(uintptr(lparam) >> 16)))
			applyBallRegion(hwnd, w, h)
		}
	}

	ret, _, _ := procDefSubclassProc.Call(hwnd, uintptr(msg), wparam, uintptr(lparam))
	return ret
}

// applyBallRegion 在窗口为球尺寸(边长 = ballSize 物理像素)时套椭圆 region,
// 其余尺寸清 region 恢复矩形(面板/配置窗不裁角)。region 交给系统托管,不释放。
func applyBallRegion(hwnd uintptr, w, h int32) {
	dpi, _, _ := procGetDpiForWindow.Call(hwnd)
	if dpi == 0 {
		dpi = 96
	}
	ballPx := int32(uint32(ballSize) * uint32(dpi) / 96)

	if w == ballPx && h == ballPx {
		hrgn, _, _ := procCreateElliptic.Call(0, 0, uintptr(w), uintptr(h))
		if hrgn != 0 {
			procSetWindowRgn.Call(hwnd, hrgn, 1)
		}
		return
	}
	procSetWindowRgn.Call(hwnd, 0, 1)
}

// setupWindowStyles 设置工具窗口样式(去任务栏/缩略图),并安装子类
// 覆盖系统默认最小窗口宽度。返回 false 表示未找到窗口(调用方仅记录)。
func setupWindowStyles(title string) bool {
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return false
	}
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	mainBallHwnd = hwnd
	if hwnd == 0 {
		return false
	}

	exStyleIdxV := int32(gwlExStyle) // 经变量转换,避免常量负数溢出 uintptr
	exStyleIdx := uintptr(exStyleIdxV)
	exStyle, _, _ := procGetWindowLongW.Call(hwnd, exStyleIdx)
	if newEx := exStyle | wsExToolWindow; newEx != exStyle {
		procSetWindowLongW.Call(hwnd, exStyleIdx, newEx)
		procSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0,
			swpNoMove|swpNoSize|swpNoZOrder|swpNoActivate|swpFrameChanged)
	}

	subclassCB = syscall.NewCallback(minTrackSubclassProc)
	ret, _, _ := procSetWindowSub.Call(hwnd, subclassCB, 1, 0)

	// 说明:Wails 的 OnStartup 与窗口不在同一线程,SetWindowSubclass 会失败
	// (comctl32 要求窗口所属线程调用),所以不能依赖子类做尺寸/region,
	// 统一走下面的物理像素直调,region 切换挂在 Expand/Collapse 流程里。

	// 直接以物理像素强制球窗尺寸:overlapped 窗口的系统最小宽度(高 DPI 下约
	// 262px 物理)会在创建期把 60px 逻辑宽的球窗撑宽,程序化 SetWindowPos 绕开。
	ForceBallWindowSize()

	// 应用球态 region(圆形裁切):窗口创建期的尺寸变化早于子类安装,启动补一次
	var wr winRect
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&wr)))
	applyBallRegion(hwnd, wr.right-wr.left, wr.bottom-wr.top)

	return ret != 0
}

// ForceBallWindowSize 以物理像素把球窗强制规整为 ballSize 逻辑边长
// (绕开 overlapped 窗口系统最小宽度对创建期尺寸的钳制)。
func ForceBallWindowSize() {
	if mainBallHwnd == 0 {
		return
	}
	dpi, _, _ := procGetDpiForWindow.Call(mainBallHwnd)
	if dpi == 0 {
		dpi = 96
	}
	ballPx := int32(uint32(ballSize) * uint32(dpi) / 96)
	procSetWindowPos.Call(mainBallHwnd, 0, 0, 0, uintptr(ballPx), uintptr(ballPx),
		swpNoZOrder|swpNoActivate|swpNoMove)
}

// SyncBallRegion 在展开/收起切换后同步窗口 region:球态裁圆,面板态恢复矩形。
func SyncBallRegion() {
	if mainBallHwnd == 0 {
		return
	}
	var wr winRect
	if r, _, _ := procGetWindowRect.Call(mainBallHwnd, uintptr(unsafe.Pointer(&wr))); r != 0 {
		applyBallRegion(mainBallHwnd, wr.right-wr.left, wr.bottom-wr.top)
	}
}

// syncBallRegion 是 app.go 用的包内别名(非 Windows 平台为空实现)。
func syncBallRegion() { SyncBallRegion() }

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
