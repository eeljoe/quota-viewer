package main

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
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
		cfg = config.Default()
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
	// 单击托盘图标 = 唤出到前台:显示 + 激活 + 重申置顶,幂等(不会隐藏)。
	// 前台激活交给 Wails 的 WindowShow:它内部先 ShowWindow(必要时还原最小化),
	// 再调 SetForegroundWindow + SetFocus,因此不需要自己写 Win32 抢焦点。
	wailsruntime.EventsOn(ctx, "tray:activate", func(...interface{}) {
		a.visible.Store(true)
		wailsruntime.WindowShow(ctx)
		wailsruntime.WindowSetAlwaysOnTop(ctx, true)
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

// ProviderInput 是前端提交的 Provider 状态。
// 凭证字段空字符串 = 不修改(避免掩码回写覆盖真实值)。
// Extra 是第 2..N 组凭证,按 Key 对齐到已有组;未提交的组即视为删除。
type ProviderInput struct {
	ID      string             `json:"id"`
	Enabled bool               `json:"enabled"`
	Creds   map[string]string  `json:"creds"`
	Budget  float64            `json:"budget"`
	Extra   []config.CredGroup `json:"extra"`
}

// maskCredSet 对一组凭证逐字段掩码。
func maskCredSet(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = maskSecret(v)
	}
	return out
}

// mergeCredsInto 把前端提交的一组凭证合并进已有值并返回合并结果:
// 非空字段覆盖,空字段 = 不修改(掩码回显时前端提交的是空串)。
func mergeCredsInto(dst map[string]string, in map[string]string, fields []fetcher.CredentialField) map[string]string {
	if dst == nil {
		dst = map[string]string{}
	}
	for _, f := range fields {
		if v, ok := in[f.Key]; ok && v != "" {
			dst[f.Key] = config.NormalizeCookieInput(v)
		}
	}
	return dst
}

// GetConfig 返回当前配置(凭证做掩码)与全部 Provider 元数据。
func (a *App) GetConfig() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 当前配置索引
	cur := map[string]config.ProviderConfig{}
	for _, p := range a.cfg.Providers {
		cur[p.ID] = p
	}

	providers := make([]map[string]interface{}, 0, len(fetcher.GetAll()))
	for _, def := range fetcher.GetAll() {
		pc, ok := cur[def.ID]
		creds := map[string]string{}
		extra := make([]map[string]interface{}, 0, len(pc.Extra))
		if ok {
			creds = maskCredSet(pc.Creds)
			for _, g := range pc.Extra {
				extra = append(extra, map[string]interface{}{"key": g.Key, "creds": maskCredSet(g.Creds)})
			}
		}
		fields := make([]map[string]string, 0, len(def.Fields))
		for _, f := range def.Fields {
			fields = append(fields, map[string]string{
				"key": f.Key, "label": f.Label, "type": f.Type,
			})
		}
		providers = append(providers, map[string]interface{}{
			"id":        def.ID,
			"name":      def.DisplayName,
			"abbr":      def.Abbr,
			"kind":      def.Kind,
			"enabled":   ok && pc.Enabled,
			"login_url": def.LoginURL,
			"fields":    fields,
			"creds":     creds,
			"extra":     extra,
			"budget":    pc.Budget,
		})
	}

	return map[string]interface{}{
		"providers":            providers,
		"refresh_interval_min": a.cfg.RefreshIntervalMin,
		"ball_x":               a.cfg.BallX,
		"ball_y":               a.cfg.BallY,
		"extended_mode":        a.cfg.ExtendedMode,
		"object_limit":         config.MaxObjects(a.cfg.ExtendedMode),
		"object_count":         config.CountObjects(a.cfg.Providers),
	}
}

// SaveConfig 保存 Provider 配置。监控对象数钳制:普通模式最多 9 个,扩展模式 16 个(后端钳制)。
func (a *App) SaveConfig(providers []ProviderInput, refreshMin int, extendedMode bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	limit := config.MaxObjects(extendedMode)

	// 1) 钳制监控对象数:0 个 → 全部启用(再由上限裁掉超出者);超过上限 → 按顺序保留到上限为止。
	//    对象数 = 各启用 Provider 的凭证组数之和(套餐卡数要抓完才知道,展示层再截断)。
	enabledObjects := 0
	for _, p := range providers {
		if p.Enabled {
			enabledObjects += len(config.ProviderConfig{Extra: p.Extra}.CredSets())
		}
	}
	if enabledObjects == 0 {
		for i := range providers {
			providers[i].Enabled = true
		}
		enabledObjects = len(providers)
	}
	if enabledObjects > limit {
		kept := 0
		for i := range providers {
			if !providers[i].Enabled {
				continue
			}
			cost := len(config.ProviderConfig{Extra: providers[i].Extra}.CredSets())
			if kept+cost > limit {
				providers[i].Enabled = false
				continue
			}
			kept += cost
		}
	}
	a.cfg.ExtendedMode = extendedMode

	// 2) 合并到配置(空凭证 = 不修改)
	for _, in := range providers {
		def, ok := fetcher.Get(in.ID)
		if !ok {
			continue // 未知 provider 忽略
		}
		idx := -1
		for j := range a.cfg.Providers {
			if a.cfg.Providers[j].ID == in.ID {
				idx = j
				break
			}
		}
		if idx == -1 {
			a.cfg.Providers = append(a.cfg.Providers, config.ProviderConfig{ID: in.ID})
			idx = len(a.cfg.Providers) - 1
		}
		pc := &a.cfg.Providers[idx]
		pc.Enabled = in.Enabled
		pc.Budget = in.Budget

		pc.Creds = mergeCredsInto(pc.Creds, in.Creds, def.Fields)

		// 附加凭证组:按稳定 Key 对齐到已有组(删中间一行不会让后面的值整体错位),
		// 未提交的组即删除;合并后全空的组丢弃(用户加了行没填内容)。
		byKey := make(map[string]config.CredGroup, len(pc.Extra))
		for _, g := range pc.Extra {
			byKey[g.Key] = g
		}
		merged := make([]config.CredGroup, 0, len(in.Extra))
		usedKeys := map[string]bool{}
		for i, g := range in.Extra {
			key := g.Key
			if key == "" || usedKeys[key] {
				key = fmt.Sprintf("%s#%d", in.ID, i)
			}
			usedKeys[key] = true
			cur := mergeCredsInto(byKey[key].Creds, g.Creds, def.Fields)
			if config.CredGroupEmpty(cur) {
				continue
			}
			merged = append(merged, config.CredGroup{Key: key, Creds: cur})
		}
		pc.Extra = merged
	}

	// 3) 按注册表顺序重排(展示顺序固定)
	ordered := make([]config.ProviderConfig, 0, len(fetcher.GetAll()))
	seen := map[string]bool{}
	for _, def := range fetcher.GetAll() {
		for _, p := range a.cfg.Providers {
			if p.ID == def.ID {
				ordered = append(ordered, p)
				seen[p.ID] = true
				break
			}
		}
	}
	for _, p := range a.cfg.Providers {
		if !seen[p.ID] {
			ordered = append(ordered, p)
		}
	}
	a.cfg.Providers = ordered

	if refreshMin > 0 {
		a.cfg.RefreshIntervalMin = refreshMin
	}

	return config.Save(a.cfg)
}

// TestConnection 测试单个平台连接是否可用。
// credIndex 选第几组凭证(0 = 主凭证);越界回落到主凭证。
// 多套餐平台会一次返回多张卡,逐张校验后汇总回报。
func (a *App) TestConnection(platform string, credIndex int) string {
	def, ok := fetcher.Get(platform)
	if !ok {
		return "未知平台"
	}

	a.mu.Lock()
	var sets []map[string]string
	for _, p := range a.cfg.Providers {
		if p.ID == platform {
			sets = p.CredSets()
			break
		}
	}
	a.mu.Unlock()

	if credIndex < 0 || credIndex >= len(sets) {
		credIndex = 0
	}
	var creds map[string]string
	if len(sets) > 0 {
		creds = sets[credIndex]
	}

	results := fetcher.FetchAll(def.Build(creds))
	for _, r := range results {
		if r.Error != "" {
			return "失败: " + r.Error
		}
	}
	if len(results) == 0 {
		return "失败: 无返回数据"
	}
	if len(results) == 1 {
		return "成功: " + results[0].Remaining
	}
	parts := make([]string, 0, len(results))
	for _, r := range results {
		name := r.SubLabel
		if name == "" {
			name = "套餐"
		}
		parts = append(parts, name+" "+r.Remaining)
	}
	return fmt.Sprintf("成功: %d 个监控对象 —— %s", len(results), strings.Join(parts, " / "))
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

// fetchAll 并发抓取所有已启用的监控对象。
//
// 一个监控对象 = 一组凭证 × 一张套餐卡:每家 Provider 按凭证组展开成若干任务,
// 每组凭证再按 fetcher 返回的套餐卡展开成若干结果。结果顺序 =
// 注册表顺序 → 凭证组顺序 → 套餐卡顺序(与 probe 探针同口径)。
//
// 这里不截断:悬浮球的格数上限由前端按 enabled_mode 计算并在溢出时显示 "+N",
// 展开面板永远列全部对象。截断交给展示层,抓取层保持全量。
func (a *App) fetchAll() []fetcher.QuotaResult {
	a.mu.Lock()
	cfg := *a.cfg
	a.mu.Unlock()

	type job struct {
		p      config.ProviderConfig
		def    fetcher.ProviderDef
		creds  map[string]string
		credNo int
		known  bool
	}

	var jobs []job
	for _, p := range cfg.Providers {
		if !p.Enabled {
			continue
		}
		def, ok := fetcher.Get(p.ID)
		sets := p.CredSets()
		for i, creds := range sets {
			jobs = append(jobs, job{p: p, def: def, creds: creds, credNo: i, known: ok})
		}
	}

	perJob := make([][]fetcher.QuotaResult, len(jobs))
	var wg sync.WaitGroup
	for i := range jobs {
		wg.Add(1)
		go func(i int, j job) {
			defer wg.Done()
			if !j.known {
				perJob[i] = []fetcher.QuotaResult{{Platform: j.p.ID, Error: "未知平台"}}
				return
			}
			rs := fetcher.FetchAll(j.def.Build(j.creds))
			credCount := len(j.p.CredSets())
			for k := range rs {
				rs[k].ID = j.def.ID
				if rs[k].Kind == "" {
					rs[k].Kind = fetcher.KindUsage
				}
				// 身份归一(缩写/后缀/slot)与 probe 共用 SettleIdentity。
				fetcher.SettleIdentity(&rs[k], j.def, j.credNo, k, len(rs), credCount)
				fetcher.ApplyBudget(&rs[k], j.p.Budget)
			}
			perJob[i] = rs
		}(i, jobs[i])
	}
	wg.Wait()

	results := make([]fetcher.QuotaResult, 0, len(jobs))
	for _, rs := range perJob {
		results = append(results, rs...)
	}
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
