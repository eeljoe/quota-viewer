# 架构总览

**When to read**: 理解项目整体结构、分层职责、数据流时。

---

## 核心内容

### 是什么

桌面悬浮球 + 额度监控工具：可配置 1-3 个 AI 平台的 API 配额/余额监控（Kimi / 讯飞星辰 / OpenCode Go / 小米 MiMo / DeepSeek / Ollama）。单窗口（60×60 悬浮球 ⇄ 展开详情面板），常驻系统托盘，关闭到托盘不退出。

### 技术栈

- Go 1.24 + Wails v2.12.0（frameless + AlwaysOnTop + 透明背景）
- 原生 HTML/CSS/JS 前端，Vite 打包，产物 `go:embed` 进 exe
- 系统托盘：`github.com/energye/systray`
- 平台：Windows 10+（WebView2 运行时）

### 分层职责

```
┌─────────────────────────────────────────────────┐
│ frontend/src (悬浮球/面板/配置 UI)               │
│  index.html + main.js + style.css               │
└──────────────────────┬──────────────────────────┘
                       │ Wails 绑定 (wailsjs/go/main/App.*)
                       │ 事件总线 (quota:update / ui:show-settings)
┌──────────────────────▼──────────────────────────┐
│ app.go (编排层)                                 │
│  OnStartup / Refresh / fetchAll / 窗口定位       │
│  GetConfig / SaveConfig / TestConnection        │
└───────┬──────────────┬──────────────┬───────────┘
        │              │              │
┌───────▼──────┐ ┌─────▼──────┐ ┌─────▼──────────┐
│ internal/    │ │ internal/  │ │ internal/tray  │
│ fetcher      │ │ config     │ │ (托盘菜单)      │
│ registry.go  │ │ 动态        │ │ systray.Run    │
│ (注册表)      │ │ Provider   │ │ + LockOSThread │
│ kimi/xfyun/  │ │ 列表持久化  │ └────────────────┘
│ opencode_go/ │ │ + Cookie   │
│ mimo/deepseek│ │ 解析        │
│ ollama       │ │             │
└──────────────┘ └────────────┘
        │
        │ 直接调用方:app.go fetchAll 按启用列表并发抓取
```

- **fetcher 层**：`registry.go` 注册表（ProviderDef：id/显示名/缩写/凭证字段/工厂）+ 每平台一个实现，返回统一 `QuotaResult`
- **config 层**：动态 `Providers []ProviderConfig` 持久化到 `%APPDATA%/quota-viewer/config.json`，旧扁平格式自动迁移
- **tray 层**：托盘图标/菜单，点击事件通过 Wails 事件总线转给 app.go
- **app.go**：唯一编排者——持有 ctx/cfg/cache，按启用 Provider 并发抓取，事件推送前端

### 数据流（一次刷新）

```
startAutoRefresh (每 N 分钟) ──┐
tray:refresh 事件 ─────────────┤→ App.Refresh()
前端手动刷新 ──────────────────┘
   → fetchAll(): 启用列表(1-3)并发 Fetch()
   → 写 a.cache → EventsEmit("quota:update", results)
   → 前端 updateBall()(动态重建格子) / updatePanel() 渲染
```

### 窗口两种形态

| 形态 | 尺寸 | 触发 |
|---|---|---|
| 悬浮球（收起） | 60×60 | 启动 / CollapseWindow / 点击球体收起 |
| 详情面板（展开） | 动态 (w,h) | ExpandWindow（前端请求） |

展开时记录球位置（savedX/savedY），收起时精确还原；展开定位做四角智能翻转 + 屏幕钳制（见 04-window-positioning.md）。

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `main.go` | 入口，Wails 运行时选项（frameless/透明/置顶/隐藏关闭） |
| `app.go` | 编排层：启动流程、刷新编排、配置读写、窗口定位、事件转发 |
| `workarea_windows.go` | Win32：显示器工作区查询、DPI、窗口样式/子类（Windows only） |
| `internal/fetcher/*.go` | 注册表 + 五平台抓取器 + 统一类型 |
| `internal/config/*.go` | 配置持久化 + Cookie 解析 |
| `internal/tray/tray.go` | 系统托盘 |
| `frontend/src/*` | 全部 UI |

---

## Must NOT Change

- 单一编排层模式：抓取/配置/托盘不得互相直接调用，必须经 app.go
- 事件名契约（见 00-agent-rules.md）
- 前端产物必须构建后提交（Go embed 依赖 dist 内容）
