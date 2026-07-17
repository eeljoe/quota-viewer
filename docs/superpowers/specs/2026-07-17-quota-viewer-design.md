# Quota Viewer 悬浮球设计文档

- **日期**:2026-07-17
- **状态**:已确认,待实现规划
- **技术栈**:Wails v2 + Go + 原生 HTML/CSS/JS

## 背景与目标

本机同时使用多个 Coding Plan(Kimi 会员、讯飞星辰 MaaS、小米 MiMo),每次查额度都要开浏览器登录控制台,Chrome 常驻 14 个进程占用约 2.2 GB 内存,是本机 RAM 瓶颈的主要来源之一。

目标:做一个轻量悬浮球小工具,不开浏览器即可一眼看到三个平台的额度状态。底线是内存占用远小于开 Chrome(目标 < 150 MB,实测 Chrome 约 2.2 GB)。

## 调研结论

### 悬浮球技术可行性

Electron、Wails、Tauri 均原生支持 `frameless + transparent + alwaysOnTop` 窗口。Windows 上 Wails 和 Tauri 都使用系统 WebView2,不额外捆绑 Chromium;Electron 自带完整 Chromium 副本,内存占用最高。

三方案内存对比(Windows 实测基准):

| 方案 | 空闲 RAM | 包体积 | 透明置顶窗 |
|---|---|---|---|
| Electron 34 | ~168 MB | ~85 MB+ | 稳定 |
| Wails v2 | ~80-120 MB | ~10 MB | v2 文档稳定 |
| Tauri 2 | ~42-80 MB | ~3 MB | beta,有坑 |

实测当前 Chrome 14 进程总工作集 2296 MB(2.24 GB)。Wails v2 预期占用约 80-120 MB,约为 Chrome 现状的 4-5%,底线稳稳满足。

### 三平台额度获取方式

| 平台 | 额度查询路径 | 开源参考 |
|---|---|---|
| Kimi 会员 | Kimi Code API Key(`sk-kimi-xxx`)调接口 | `Golden0Voyager/kimi-code-usage`(Python,逻辑可复用) |
| 讯飞星辰 MaaS | 无公开 API,靠 Cookie 抓 `maas.xfyun.cn/packageSubscription` 背后 XHR | 无,需自行抓包 |
| 小米 MiMo | 无公开 API,靠 Cookie 抓 `platform.xiaomimimo.com/console/plan-manage` 背后 XHR | 无,需自行抓包 |

讯飞与小米无文档化公开额度查询 API,只能以登录态 Cookie 请求网页接口。具体 XHR 地址与响应字段需首次运行时用 F12 Network 实抓确认。

### 参考开源项目

- `projectvelox/claude-usage-widget`:Electron + Windows,悬浮窗显示 claude.ai 用量,结构可借鉴。
- `Golden0Voyager/kimi-code-usage`:Kimi Code 用量查询,CLI + MCP + VSCode 插件,API 调用逻辑可直接参考。

## 架构

### 整体结构

单进程、单窗口双态。Go 后端负责数据抓取与配置管理,前端负责悬浮球 UI 与详情面板渲染。Go 方法通过 Wails binding 暴露给前端,数据更新通过 `runtime.EventsEmit` 推送。

```
┌──────────────────────────────────────────┐
│  主窗口 (frameless, transparent,          │
│  alwaysOnTop, skipTaskbar)                │
│  收起态:60×60 圆形球面                    │
│  展开态:~320×240 详情面板                  │
└──────────────────────────────────────────┘
        │ Wails binding (Go ↔ JS)
        ▼
┌──────────────────────────────────────────┐
│  Go 后端                                 │
│  ├─ fetcher.KimiFetcher   (API Key)      │
│  ├─ fetcher.XfyunFetcher (Cookie 抓 XHR) │
│  └─ fetcher.MiMoFetcher   (Cookie 抓 XHR) │
│  统一返回 QuotaResult                     │
└──────────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────────┐
│  config.json (%APPDATA%/quota-viewer/)   │
│  各平台凭证 + 刷新间隔 + 球位置           │
└──────────────────────────────────────────┘
```

### 数据获取层

统一接口,三个实现,Go 并发调用。

```go
type QuotaResult struct {
    Platform    string  // "Kimi" / "讯飞星辰" / "小米MiMo"
    Used        float64 // 已用量
    Total       float64 // 总量(平台返回则填,否则 0)
    Percent     float64 // Used/Total;无总量时由平台剩余百分比反推
    Remaining   string  // 原始剩余描述(如 "1,200/18,000 次" 或 "无限制")
    ResetAt     string  // 下次重置时间(ISO 8601,空则未知)
    LastUpdated time.Time
    Error       string  // 非空表示失败(如 "Cookie 已过期,请更新")
}
```

**KimiFetcher(API Key)**

- 凭证:`sk-kimi-xxx` API Key,存配置文件。
- 端点:Kimi Code 控制台额度接口,参考 `kimi-code-usage` 项目逻辑(Anthropic 兼容协议)。
- 直接获取剩余额度、套餐用量、刷新时间,无需 Cookie。最稳定的一条。

**XfyunFetcher(Cookie 抓网页接口)**

- 凭证:从 `maas.xfyun.cn` 登录态复制的 Cookie 字符串。
- 用 `http.Client` + `cookiejar` 带 Cookie 请求 `packageSubscription` 页面背后的 XHR。
- 具体 XHR 地址与响应字段首次需 F12 实抓确认;MVP 阶段保留可配置的"接口 URL + 响应字段映射",跑通后固化。
- Cookie 失效(401/302 跳登录页):返回 `Error: "Cookie 已过期,请更新"`,前端提示 + 一键打开 `maas.xfyun.cn` 登录页。

**MiMoFetcher(Cookie 抓网页接口)**

- 凭证:从 `platform.xiaomimimo.com` 登录态复制的 Cookie。
- 抓 `console/plan-manage` 背后的 XHR,处理方式与讯飞一致。
- Cookie 失效处理同上。

### 凭证存储

路径:`%APPDATA%/quota-viewer/config.json`,明文存储(本地单用户工具,MVP 不加密)。

```json
{
  "kimi_api_key": "sk-kimi-xxx",
  "xfyun_cookie": "SSID=xxx; ...",
  "mimo_cookie": "session=xxx; ...",
  "refresh_interval_min": 15,
  "ball_position": {"x": 1720, "y": 50}
}
```

### 首次使用引导

启动后检测配置缺失 → 弹出引导面板,分三步填入凭证,每步带"测试连接"按钮验证可达性后再保存。

## 窗口与交互

### 单窗口双态

- **收起态(默认)**:60×60 圆形小球,`frameless + transparent + alwaysOnTop + skipTaskbar`。球面用三段弧或三个小点颜色指示三平台状态:绿=正常有余量,黄=临近上限,红=耗尽或凭证失效。
- **展开态**:点击小球后窗口展开为详情面板(约 320×240),显示三平台各自的进度条 + 用量数字 + 上次刷新时间。再点球或点空白处收起。`alwaysOnTop` 两态保持。
- **拖动**:球面区域 `--wails-draggable: drag`,可拖到任意位置;位置存配置,下次启动恢复。
- **右键菜单**:`刷新` / `打开配置` / `打开各平台登录页(辅助更新 Cookie)` / `退出`。

### 系统托盘

Wails v2 原生支持。托盘图标显示聚合状态色,双击显隐球,右键菜单同上。关闭主窗口即最小化到托盘,不退出进程(`HideWindowOnClose: true`)。

### 刷新策略

- 默认每 15 分钟自动后台拉取(可配置)。
- 展开时若数据超过 3 分钟则触发一次刷新。
- 右键"刷新"立即拉取。
- 所有 HTTP 请求并发执行,带 10s 超时。
- 超时后用上次缓存数据(内存保留),标记"数据可能过期"。

## 项目结构

```
quota-viewer/
├── main.go                    # Wails 入口,窗口选项配置
├── app.go                     # 暴露给前端的方法
├── internal/
│   ├── fetcher/
│   │   ├── types.go           # QuotaResult 接口定义
│   │   ├── kimi.go            # KimiFetcher
│   │   ├── xfyun.go           # XfyunFetcher
│   │   └── mimo.go            # MiMoFetcher
│   ├── config/
│   │   └── config.go          # 配置读写
│   └── tray/
│       └── tray.go            # 系统托盘
├── frontend/
│   ├── dist/                  # 构建产物
│   └── src/
│       ├── main.js            # 入口
│       ├── style.css          # 样式(审美打磨留实现阶段)
│       └── index.html         # 双态切换(球面 / 详情面板)
├── wails.json
└── go.mod
```

## 前端

不引框架,纯原生 HTML/CSS/JS。理由:悬浮球 UI 极简(三个进度条 + 几个状态点),零依赖 = 更小包体积、更低内存。Wails binding 直接暴露 Go 方法,JS 调 `window.go.main.App.Refresh()` 即可。

状态同步:Go 端刷新完成后 `runtime.EventsEmit` 推送数据,前端监听事件更新 UI,无需轮询。

## 错误处理

- 单个 Fetcher 失败不影响其他——三个并发,各自独立返回(含 Error 字段)。
- 前端对失败平台显示"⚠ 点击更新凭证",不阻断整体展示。
- Cookie 失效:球面对应段变红 + 详情面板提示 + 右键"打开登录页"快捷跳转。
- 网络超时:10s,超时后用上次缓存数据,标记"数据可能过期"。

## MVP 范围

必须完成:
1. Wails v2 悬浮球窗口(收起/展开双态 + 拖动 + 位置记忆)
2. 系统托盘(显隐 + 右键菜单)
3. KimiFetcher(API Key 走通)
4. XfyunFetcher(Cookie 抓包走通,含 Cookie 失效提示)
5. MiMoFetcher(同上)
6. 详情面板三平台额度展示
7. 配置文件读写 + 首次引导

后续优化:
- 外观审美打磨
- 凭证加密存储
- 额度告警通知
- 历史 7 天用量图表

## 不在范围

- 不做多账户/多 API Key 管理(MVP 单用户单套凭证)
- 不做自动登录/自动刷新 Cookie(Cookie 过期需手动更新)
- 不做跨平台(macOS/Linux)适配(MVP 仅 Windows)
