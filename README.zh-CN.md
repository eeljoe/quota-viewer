# Quota Viewer

[English](README.md) | **[简体中文](README.zh-CN.md)**

桌面悬浮球 + AI 平台额度监控工具：实时显示各 AI 平台的 API 配额 / 余额剩余。

![悬浮球](docs/screenshots/preview-1.png)
![展开面板](docs/screenshots/preview-2.png)

- **技术栈**: Go + Wails v2.12.0 + 原生 HTML/CSS/JS (Vite 打包)
- **平台**: Windows 10+ (WebView2 运行时)

## 功能

- **悬浮球**: 60×60 玻璃质感方块，1-3 格展示启用的 Provider（格字颜色 = 状态色：绿 / 黄 / 红），悬停 tooltip 显示各平台明细
- **自适应球格**: 只启用 1 个 Provider 时字母占满整球，2 个时各占 1/2，3 个时各占 1/3
- **展开面板**: 进度条 + 剩余量明细（千分位数字），自动刷新（默认 15 分钟），刷新倒计时（如 Kimi 5 小时窗口重置时间）
- **预算进度条**: 余额型 Provider（如 DeepSeek）支持配置预算——进度条按 `(预算 - 当前余额) / 预算` 显示消耗，默认预算 ¥300
- **配置面板**: 勾选要展示的 Provider（**最多 3 个、最少 1 个**），每个 Provider 独立录入自己的凭证（API Key / Cookie），支持直接粘贴浏览器 "Copy as PowerShell" 格式
- **系统托盘**: 刷新 / 显示隐藏 / 打开配置 / 退出；双击托盘图标切换显示
- **窗口定位**: 右下角展开不越屏（四角智能翻转，多显示器 / DPI 缩放适配），收起精确回到原位
- 关闭到托盘，不退出进程

## 支持的 Provider

| Provider | 凭证 | 展示内容 |
|----------|------|----------|
| Kimi | API Key (`sk-kimi-xxx`) | 5 小时窗口 / 周额度用量 |
| 讯飞星辰 | Cookie | 套餐总量用量 |
| OpenCode Go | Workspace ID + Session Token | 滚动窗口额度百分比 |
| 小米 MiMo | Cookie | 套餐 Token 用量 |
| DeepSeek | API Key (`sk-...`) | 账户余额（余额型） |
| Ollama | 浏览器会话 Cookie | Cloud 5 小时 Session 用量 / 周用量（设置页） |
| Command Code | API Key (`user_...`,留空自动读 `~/.commandcode/auth.json`) | 5 小时 / 周滚动窗口用量 + 剩余 credits |

> 提示：Cookie 类凭证支持在浏览器 F12 → 网络标签 → 复制请求头后，直接粘贴 "Copy as PowerShell" 格式的整段脚本，保存时自动解析为 Cookie 请求头。

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
├── app.go               # Wails 应用主逻辑（窗口定位、配置、刷新编排）
├── main.go              # 入口 + Wails 运行时配置 (frameless / AlwaysOnTop)
├── workarea_windows.go  # Win32 辅助（工作区查询、DPI、最小窗口宽度子类）
├── workarea_other.go    # 非 Windows 桩
├── internal/
│   ├── config/          # 配置持久化（动态 Provider 列表 + PowerShell Cookie 解析）
│   ├── fetcher/         # Provider 注册表 + 各平台抓取器
│   └── tray/            # 系统托盘（energye/systray）
└── frontend/
    ├── src/
    │   ├── index.html   # 悬浮球 + 详情面板 + 配置面板 + Toast
    │   ├── main.js      # 视图状态机 / 动态球格 / 数据渲染 / 交互
    │   └── style.css    # 设计令牌 + 深色玻璃质感
    └── dist/            # Vite 构建产物（Go embed 进 exe）
```

## 如何新增一个 Provider

Quota Viewer 采用注册表驱动架构——新增 Provider 只需实现抓取器 + 注册一行代码，**前端零改动**。

如果你想让 AI agent 新增一个 Provider，复制以下内容给它：

> 阅读 `docs/ADDING_A_PROVIDER.md`，按照文档为 `<平台名>` 新增一个 Provider。端点是 `<URL>`，认证方式是 `<方式>`，展示内容是 `<展示什么>`。

手动步骤详见 [`docs/ADDING_A_PROVIDER.md`](docs/ADDING_A_PROVIDER.md)。

## 免责声明

本项目是**非官方**工具，与各平台无任何关联。各平台网页结构与接口可能随时变动导致抓取失效；凭证仅保存在本机配置文件中（`%APPDATA%/quota-viewer/config.json`），不会上传到任何服务器。请自行承担使用风险。

## 许可证

[MIT](LICENSE)
