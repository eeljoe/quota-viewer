package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Quota Viewer",
		Width:     60,
		Height:    60,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: options.NewRGBA(0, 0, 0, 0),
		// TODO(实验): Windows 透明选项会让 SetWindowSubclass 失败(窗口被撑到 262 宽),
		// 球形改用 SetWindowRgn 裁圆实现,暂不开启。见会话 2026-10-02。
		AlwaysOnTop:       true,
		HideWindowOnClose: true,
		DisableResize:     true,
		WindowStartState:  options.Normal,
		OnStartup:         app.OnStartup,
		OnShutdown:        app.OnShutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
