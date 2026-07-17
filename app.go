package main

import (
	"context"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"quota-viewer/internal/config"
	"quota-viewer/internal/fetcher"
	"quota-viewer/internal/tray"
	"sync/atomic"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	cfg     *config.Config
	mu      sync.Mutex
	cache   []fetcher.QuotaResult
	tray    *tray.TrayHandler
	visible atomic.Bool
}

func NewApp() *App {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		cfg = &config.Config{RefreshIntervalMin: 15, BallX: -1, BallY: -1}
	}
	return &App{
		cfg: cfg,
	}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.visible.Store(true)

	// 恢复悬浮球位置(配置中 BallX/BallY >= 0 时生效)
	if a.cfg.BallX >= 0 && a.cfg.BallY >= 0 {
		wailsruntime.WindowSetPosition(ctx, a.cfg.BallX, a.cfg.BallY)
	}

	// 设置系统托盘菜单(刷新/显示隐藏/打开配置/退出)
	a.tray = tray.New(ctx)
	a.tray.Start()

	// 监听托盘事件并转发到对应行为
	wailsruntime.EventsOn(ctx, "tray:refresh", func(...interface{}) {
		a.Refresh()
	})
	wailsruntime.EventsOn(ctx, "tray:toggle", func(...interface{}) {
		// Wails v2.12.0 无 WindowIsVisible,用本地可见性状态切换
		if a.visible.Load() {
			a.visible.Store(false)
			wailsruntime.WindowHide(ctx)
		} else {
			a.visible.Store(true)
			wailsruntime.WindowShow(ctx)
		}
	})
	wailsruntime.EventsOn(ctx, "tray:settings", func(...interface{}) {
		wailsruntime.EventsEmit(ctx, "ui:show-settings")
	})

	// 启动后台定时刷新
	go a.startAutoRefresh()
}

// OnShutdown 在应用退出时清理托盘图标。
func (a *App) OnShutdown(ctx context.Context) {
	if a.tray != nil {
		a.tray.Quit()
	}
}

// Refresh 并发调用三个 fetcher,返回结果并推送事件到前端。
func (a *App) Refresh() []fetcher.QuotaResult {
	results := a.fetchAll()
	a.mu.Lock()
	a.cache = results
	a.mu.Unlock()

	// 推送事件到前端
	wailsruntime.EventsEmit(a.ctx, "quota:update", results)

	return results
}

// GetConfig 返回当前配置(Cookie/Key 做掩码)。
func (a *App) GetConfig() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	return map[string]interface{}{
		"kimi_api_key":         maskSecret(a.cfg.KimiAPIKey),
		"xfyun_cookie":         maskSecret(a.cfg.XfyunCookie),
		"mimo_cookie":          maskSecret(a.cfg.MimoCookie),
		"refresh_interval_min": a.cfg.RefreshIntervalMin,
		"ball_x":               a.cfg.BallX,
		"ball_y":               a.cfg.BallY,
		"has_kimi_key":         a.cfg.KimiAPIKey != "",
		"has_xfyun_cookie":     a.cfg.XfyunCookie != "",
		"has_mimo_cookie":      a.cfg.MimoCookie != "",
	}
}

// SaveConfig 保存凭证配置。
func (a *App) SaveConfig(kimiKey, xfyunCookie, mimoCookie string, refreshMin int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 空字符串 = 不修改(避免掩码覆盖)
	if kimiKey != "" {
		a.cfg.KimiAPIKey = kimiKey
	}
	if xfyunCookie != "" {
		a.cfg.XfyunCookie = xfyunCookie
	}
	if mimoCookie != "" {
		a.cfg.MimoCookie = mimoCookie
	}
	if refreshMin > 0 {
		a.cfg.RefreshIntervalMin = refreshMin
	}

	return config.Save(a.cfg)
}

// TestConnection 测试单个平台连接是否可用。
func (a *App) TestConnection(platform string) string {
	a.mu.Lock()
	cfg := *a.cfg
	a.mu.Unlock()

	var f fetcher.Fetcher
	switch platform {
	case "kimi":
		f = fetcher.NewKimiFetcher(cfg.KimiAPIKey)
	case "xfyun":
		f = fetcher.NewXfyunFetcher(cfg.XfyunCookie, "")
	case "mimo":
		f = fetcher.NewMiMoFetcher(cfg.MimoCookie, "")
	default:
		return "未知平台"
	}

	result := f.Fetch()
	if result.Error != "" {
		return "失败: " + result.Error
	}
	return "成功: " + result.Remaining
}

// OpenLoginPage 用默认浏览器打开 URL。
func (a *App) OpenLoginPage(url string) {
	openBrowser(url)
}

// SaveBallPosition 保存悬浮球位置。
func (a *App) SaveBallPosition(x, y int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.BallX = x
	a.cfg.BallY = y
	return config.Save(a.cfg)
}

// SetWindowSize 由前端调用,切换收起/展开尺寸。
func (a *App) SetWindowSize(w, h int) {
	wailsruntime.WindowSetSize(a.ctx, w, h)
}

// fetchAll 并发调用三个 fetcher。
func (a *App) fetchAll() []fetcher.QuotaResult {
	a.mu.Lock()
	cfg := *a.cfg
	a.mu.Unlock()

	var wg sync.WaitGroup
	results := make([]fetcher.QuotaResult, 3)

	wg.Add(3)
	go func() {
		defer wg.Done()
		results[0] = fetcher.NewKimiFetcher(cfg.KimiAPIKey).Fetch()
	}()
	go func() {
		defer wg.Done()
		results[1] = fetcher.NewXfyunFetcher(cfg.XfyunCookie, "").Fetch()
	}()
	go func() {
		defer wg.Done()
		results[2] = fetcher.NewMiMoFetcher(cfg.MimoCookie, "").Fetch()
	}()
	wg.Wait()

	return results
}

// startAutoRefresh 定时后台刷新。
func (a *App) startAutoRefresh() {
	for {
		interval := 15
		a.mu.Lock()
		if a.cfg.RefreshIntervalMin > 0 {
			interval = a.cfg.RefreshIntervalMin
		}
		a.mu.Unlock()

		time.Sleep(time.Duration(interval) * time.Minute)
		a.Refresh()
	}
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		if s == "" {
			return ""
		}
		return "****"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}
