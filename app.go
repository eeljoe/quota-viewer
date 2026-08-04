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

	// 展开/收起窗口状态:savedX/savedY 记录展开前悬浮球位置
	expanded bool
	savedX   int
	savedY   int
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

	// Windows 对 overlapped 窗口强制默认最小宽度(高 DPI 下实测约 262px 物理),
	// 导致 60px 球窗被撑宽、球体偏左。安装子类覆盖 WM_GETMINMAXINFO,
	// 并设为工具窗口(去任务栏按钮/缩略图),再把球窗规整到精确的 60x60。
	setupWindowStyles("Quota Viewer")
	wailsruntime.WindowSetMinSize(ctx, ballSize, ballSize)
	wailsruntime.WindowSetSize(ctx, ballSize, ballSize)
	// 样式切换后重申置顶,防止球窗被其它窗口盖住
	wailsruntime.WindowSetAlwaysOnTop(ctx, true)

	// 恢复悬浮球位置(配置中 BallX/BallY >= 0 时生效)
	// BallX/BallY 来自 WindowGetPosition 返回的虚拟桌面绝对坐标,
	// 需要 fitToScreen 转为显示器相对坐标后再传给 WindowSetPosition
	if a.cfg.BallX >= 0 && a.cfg.BallY >= 0 {
		if nx, ny, ok := fitToScreen(ctx, a.cfg.BallX, a.cfg.BallY, ballSize, ballSize); ok {
			wailsruntime.WindowSetPosition(ctx, nx, ny)
		}
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
		// 窗口被隐藏时先从托盘唤出,否则配置面板不可见
		a.visible.Store(true)
		wailsruntime.WindowShow(ctx)
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
		a.cfg.XfyunCookie = config.NormalizeCookieInput(xfyunCookie)
	}
	if mimoCookie != "" {
		a.cfg.MimoCookie = config.NormalizeCookieInput(mimoCookie)
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

// ballSize 是悬浮球(收起态)窗口边长,与前端 SIZES.ball 保持一致。
const ballSize = 60

// screenMargin 是展开后面板与屏幕边缘保留的间距。
const screenMargin = 8

// ExpandWindow 调整窗口尺寸并重新定位,保证面板完整落在当前屏幕内:
// 默认从球的位置向右下展开,放不下则翻转到左上,最后整体钳制在屏幕内。
// 首次展开时记录悬浮球位置,供 CollapseWindow 精确恢复。
func (a *App) ExpandWindow(w, h int) {
	a.mu.Lock()
	if !a.expanded {
		a.savedX, a.savedY = wailsruntime.WindowGetPosition(a.ctx)
		a.expanded = true
	}
	x, y := a.savedX, a.savedY
	a.mu.Unlock()

	if nx, ny, ok := fitToScreen(a.ctx, x, y, w, h); ok {
		x, y = nx, ny
	}
	wailsruntime.WindowSetSize(a.ctx, w, h)
	wailsruntime.WindowSetPosition(a.ctx, x, y)
}

// CollapseWindow 收起为悬浮球,并恢复到展开前的位置。
func (a *App) CollapseWindow() {
	a.mu.Lock()
	a.expanded = false
	x, y := a.savedX, a.savedY
	a.mu.Unlock()

	// 将绝对坐标转为显示器相对坐标,避免 WindowSetPosition 叠加工作区原点
	if nx, ny, ok := fitToScreen(a.ctx, x, y, ballSize, ballSize); ok {
		x, y = nx, ny
	}
	wailsruntime.WindowSetSize(a.ctx, ballSize, ballSize)
	wailsruntime.WindowSetPosition(a.ctx, x, y)
}

// fitToScreen 计算让 w×h(逻辑像素)窗口完整落在球所在屏幕内的位置。
// Windows 下 WindowGetPosition 返回物理像素而 Screen.Size 是逻辑像素,
// 直接用 Screen.Size 钳制会在 DPI 缩放 >100% 时失效,因此优先走
// workAreaForPoint 的物理像素精确路径(含任务栏工作区)。
// 返回的坐标为 WindowSetPosition 所需的工作区相对坐标。
func fitToScreen(ctx context.Context, ballX, ballY, w, h int) (int, int, bool) {
	// Windows 精确路径:物理像素,含 DPI 与任务栏
	if wx, wy, ww, wh, dpi, ok := workAreaForPoint(ballX, ballY); ok && dpi > 0 {
		pw := w * dpi / 96
		ph := h * dpi / 96
		ballPhys := ballSize * dpi / 96
		margin := screenMargin * dpi / 96
		ax, ay := anchoredPos(ballX, ballY, pw, ph, ballPhys, wx, wy, ww, wh, margin)
		// WindowSetPosition(SetPos)期望相对工作区原点的坐标
		return ax - wx, ay - wy, true
	}

	// 回退路径(非 Windows 或查询失败):Wails Screen 近似,逻辑坐标
	screens, err := wailsruntime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return 0, 0, false
	}

	// 优先取窗口当前所在屏,回退主屏,再回退第一块屏
	cur := screens[0]
	for _, s := range screens {
		if s.IsCurrent {
			cur = s
			break
		}
		if s.IsPrimary {
			cur = s
		}
	}
	sw, sh := cur.Size.Width, cur.Size.Height
	if sw <= 0 || sh <= 0 {
		return 0, 0, false
	}

	// 球心必须在 (0,0,sw,sh) 坐标系内,否则说明多屏偏移,放弃钳制
	cx, cy := ballX+ballSize/2, ballY+ballSize/2
	if cx < 0 || cx > sw || cy < 0 || cy > sh {
		return 0, 0, false
	}

	x, y := anchoredPos(ballX, ballY, w, h, ballSize, 0, 0, sw, sh, screenMargin)
	return x, y, true
}

// anchoredPos 计算展开位置:默认从球的位置向右下展开,放不下则翻转到左上
// (翻转时球边缘对齐面板边缘),最后整体钳制在矩形 (rx,ry,rw,rh) 内并保留边距。
// 所有参数同一坐标系(调用方保证)。
func anchoredPos(ballX, ballY, w, h, ballPhys, rx, ry, rw, rh, margin int) (int, int) {
	x, y := ballX, ballY
	if x+w > rx+rw-margin {
		x = ballX + ballPhys - w // 向右放不下则向左展开,球右缘对齐面板右缘
	}
	if y+h > ry+rh-margin {
		y = ballY + ballPhys - h // 向下放不下则向上展开
	}
	x = max(rx+margin, min(x, rx+rw-w-margin))
	y = max(ry+margin, min(y, ry+rh-h-margin))
	return x, y
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
