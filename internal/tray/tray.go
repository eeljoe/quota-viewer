// Package tray 提供系统托盘图标与右键菜单(刷新/显示隐藏/打开配置/退出)。
//
// 注意:Wails v2.12.0 没有内置的系统托盘运行时 API
// (options.SystemTray / runtime.SetTrayMenu 在该版本中均不存在),
// 因此这里使用第三方库 github.com/energye/systray(支持与其它 UI
// 工具包共用事件循环,适合与 Wails 集成)。菜单点击通过 Wails 事件总线
// 发送给前端(与 task brief 的事件名保持一致):tray:refresh、
// tray:toggle、tray:settings。
package tray

import (
	"context"
	_ "embed"
	"runtime"

	"github.com/energye/systray"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed assets/icon.ico
var iconICO []byte

//go:embed assets/icon.png
var iconPNG []byte

// TrayHandler 持有 Wails 上下文,负责构建托盘菜单并转发点击事件。
type TrayHandler struct {
	ctx context.Context
}

// New 创建一个 TrayHandler。ctx 应为 Wails OnStartup 传入的上下文。
func New(ctx context.Context) *TrayHandler {
	return &TrayHandler{ctx: ctx}
}

// iconBytes 按平台返回合适的图标字节:Windows 使用 .ico,
// 其它平台使用 .png。
func iconBytes() []byte {
	if runtime.GOOS == "windows" {
		return iconICO
	}
	return iconPNG
}

// Start 注册系统托盘并构建右键菜单。
// 使用 RunWithExternalLoop,使托盘事件循环与 Wails 主循环共存:
// 返回的 start 在托盘就绪后被调用以启动(在独立 goroutine 中的)消息泵。
func (t *TrayHandler) Start() {
	start, _ := systray.RunWithExternalLoop(t.onReady, t.onExit)
	// 启动托盘原生消息泵(在 systray 内部的 goroutine 中运行,
	// 不会阻塞 Wails 主线程)。
	start()
}

// Quit 退出托盘。应在应用关闭时调用,以便清理托盘图标。
func (t *TrayHandler) Quit() {
	systray.Quit()
}

func (t *TrayHandler) onReady() {
	systray.SetIcon(iconBytes())
	systray.SetTitle("QV")
	systray.SetTooltip("Quota Viewer")

	// 双击托盘图标:切换悬浮球显示/隐藏
	systray.SetOnDClick(func(menu systray.IMenu) {
		wailsruntime.EventsEmit(t.ctx, "tray:toggle")
	})

	// 刷新
	mRefresh := systray.AddMenuItem("刷新", "刷新配额")
	mRefresh.Click(func() {
		wailsruntime.EventsEmit(t.ctx, "tray:refresh")
	})

	// 显示/隐藏
	mToggle := systray.AddMenuItem("显示/隐藏", "显示或隐藏悬浮球")
	mToggle.Click(func() {
		wailsruntime.EventsEmit(t.ctx, "tray:toggle")
	})

	systray.AddSeparator()

	// 打开配置
	mSettings := systray.AddMenuItem("打开配置", "打开配置面板")
	mSettings.Click(func() {
		wailsruntime.EventsEmit(t.ctx, "tray:settings")
	})

	systray.AddSeparator()

	// 退出
	mQuit := systray.AddMenuItem("退出", "退出 Quota Viewer")
	mQuit.Click(func() {
		wailsruntime.Quit(t.ctx)
	})
}

func (t *TrayHandler) onExit() {
	// 托盘原生资源已由 systray 释放,无需额外处理。
}
