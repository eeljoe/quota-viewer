# 系统托盘

**When to read**: 修改托盘菜单、图标、托盘与主应用交互时。

---

## 核心内容

### 为什么用 energye/systray

Wails v2.12.0 没有内置托盘 API（`options.SystemTray` / `runtime.SetTrayMenu` 在该版本不存在）。选用 `github.com/energye/systray`（支持与其它 UI 工具包共用事件循环）。

### 线程铁律（2026-07-25 fix/tray-unresponsive 的教训）

**Win32: HWND 必须在同一 OS 线程创建和调度消息。** 此前托盘间歇性无响应，因为 systray 内部 Win32 消息泵跑在未锁线程的 goroutine 里。修复：

```go
func (t *TrayHandler) Start() {
    go func() {
        runtime.LockOSThread()      // 关键
        defer runtime.UnlockOSThread()
        systray.Run(t.onReady, t.onExit)
    }()
}
```

### 事件转发模型

托盘不直接调用 app 方法，通过 Wails 事件总线转发：

| 菜单/动作 | 事件 | app.go 处理 |
|---|---|---|
| 刷新 | `tray:refresh` | `a.Refresh()` |
| 显示/隐藏 | `tray:toggle` | visible 标志 + WindowHide/WindowShow |
| 打开配置 | `tray:settings` | 先 Show 窗口再 emit `ui:show-settings` |
| 双击图标 | `tray:toggle` | 同上 |
| 退出 | 直接 `wailsruntime.Quit` | OnShutdown → tray.Quit() |

事件监听在 `app.go OnStartup` 里 `wailsruntime.EventsOn` 注册。

### 图标

- `internal/tray/assets/icon.ico`（Windows）/ `icon.png`（其它平台），`go:embed` 编译进二进制
- `systray.SetTitle("QV")` / `SetTooltip("Quota Viewer")`

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `internal/tray/tray.go` | TrayHandler、菜单构建、事件发射 |
| `internal/tray/assets/` | 图标资源（embed） |
| `app.go` | 事件监听与行为实现（OnStartup / OnShutdown） |

---

## Must NOT Change

- `tray:refresh` / `tray:toggle` / `tray:settings` 事件名（两端契约）
- `systray.Run` 必须包在 `runtime.LockOSThread()` goroutine 中
- 退出路径：托盘"退出"→ `wailsruntime.Quit` → OnShutdown → `tray.Quit()`
