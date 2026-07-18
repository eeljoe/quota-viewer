# Quota Viewer

桌面悬浮球 + 额度监控工具，实时显示 Kimi / 讯飞星辰 / 小米 MiMo 三大 AI 平台的 API 配额剩余。

- **技术栈**: Go + Wails v2.12.0 + 原生 HTML/CSS/JS (Vite 打包)
- **平台**: Windows 10+ (WebView2 运行时)

## 功能

- **悬浮球**: 60×60 玻璃质感方块，三格显示 K/讯/M 状态色（绿 / 黄 / 红），悬停 tooltip 显示各平台明细
- **展开面板**: 进度条 + 剩余量明细（千分位数字），自动刷新（默认 15 分钟）
- **配置面板**: 录入各平台 API Key / Cookie，支持直接粘贴浏览器 "Copy as PowerShell" 格式
- **系统托盘**: 刷新 / 显示隐藏 / 打开配置 / 退出；双击托盘图标切换显示
- **窗口定位**: 右下角展开不越屏（四角智能翻转），收起精确回到原位
- 关闭到托盘，不退出进程

## 开发

```bash
# 安装依赖
go mod tidy
cd frontend && npm install && cd ..

# 开发模式（热重载）
wails dev

# 构建
wails build
```

产物: `build/bin/quota-viewer.exe`

## 项目结构

```
├── app.go               # Wails 应用主逻辑（窗口定位、配置、刷新）
├── main.go              # 入口 + Wails 运行时配置 (frameless / AlwaysOnTop)
├── workarea_windows.go   # Win32 辅助（工作区查询、最小窗口宽度子类）
├── workarea_other.go     # 非 Windows 桩
├── internal/
│   ├── config/          # 配置持久化（含 PowerShell Cookie 解析）
│   ├── fetcher/         # Kimi / 讯飞 / MiMo 三个平台的 HTTP 抓取器
│   └── tray/            # 系统托盘（energye/systray）
└── frontend/
    ├── src/
    │   ├── index.html   # 悬浮球 + 详情面板 + 配置面板 + Toast
    │   ├── main.js      # 视图状态机 / 数据渲染 / 交互
    │   └── style.css    # 设计令牌 + 深色玻璃质感
    └── dist/            # Vite 构建产物（Go embed 进 exe）
```

## 配置说明

| 平台 | 配置项 | 获取方式 |
|------|--------|----------|
| Kimi | API Key (`sk-kimi-xxx`) | Kimi 开放平台 → API Key 管理 |
| 讯飞星辰 | Cookie | 浏览器 F12 → 网络标签 → 复制请求头 |
| 小米 MiMo | Cookie | 浏览器 F12 → 网络标签 → 复制请求头（或 "Copy as PowerShell" 一键粘贴） |

MiMo / 讯飞的 Cookie 输入框支持直接粘贴浏览器 "Copy as PowerShell" 格式的整段脚本，保存时自动解析为 Cookie 请求头。

## 变更记录

| 日期 | 版本/分支 | 内容 |
|------|-----------|------|
| 2026-07-18 | feat/ui-polish | UI 大修：玻璃质感深色主题、三格悬浮球、展开定位修复（200% DPI 适配）、PowerShell Cookie 解析、Toast 替代 alert、千分位数字 |
| 2026-07-17 | master | 三个平台抓取器切换到真实 API 端点、系统托盘、悬浮球位置记忆 |