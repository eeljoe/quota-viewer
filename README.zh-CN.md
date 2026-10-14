# Quota Viewer

[English](README.md) | **[简体中文](README.zh-CN.md)**

桌面悬浮球 + AI 平台额度监控工具：实时显示各 AI 平台的 API 配额 / 余额剩余。

![悬浮球](docs/screenshots/preview-1.png)
![展开面板](docs/screenshots/preview-2.png)

**十家平台**：[Kimi](https://platform.kimi.com) · [讯飞星辰](https://maas.xfyun.cn) · [OpenCode Go](https://opencode.ai/zen) · [小米 MiMo](https://platform.xiaomimimo.com) · [DeepSeek](https://www.deepseek.com/platform) · [Ollama](https://ollama.com/settings) · [Command Code](https://commandcode.ai) · [Factory Droid](https://factory.ai) · [MiniMax](https://platform.minimaxi.com) · [Open Design](https://open-design.ai)

> 注册表支持 10 家平台，悬浮球**同屏最多显示 9 格**（普通模式 3 格，扩展模式 9 格）。

- **技术栈**: Go 1.24 + Wails v2.12.0 + 原生 HTML/CSS/JS (Vite 打包)
- **平台**: Windows 10+ (WebView2 运行时)
- **体积**: 单个 exe 约 10.9 MB，运行中工作集约 60 MB

## 功能

- **悬浮球**: 60×60 玻璃质感方块，展示启用的 Provider（格字颜色 = 状态色：绿 / 黄 / 红），悬停 tooltip 显示各平台明细
- **自适应球格**: 一至三格单行（1 格占满整球，2 格各半，3 格各三分之一）；扩展模式下 4 格走 2×2 网格，5-9 格按 3 列换行
- **展开面板**: 进度条 + 剩余量明细（千分位数字），自动刷新（默认 15 分钟），刷新倒计时（如 Kimi 5 小时窗口重置时间）
- **预算进度条**: 余额型 Provider（如 DeepSeek）支持配置预算——进度条按 `(预算 - 当前余额) / 预算` 显示消耗，默认预算 ¥300
- **配置面板**: 勾选要展示的 Provider（**普通模式上限 3 个，开启扩展模式后上限 9 个，最少 1 个**），每个 Provider 独立录入自己的凭证（API Key / Cookie），支持直接粘贴浏览器 "Copy as PowerShell" 格式
- **系统托盘**: 刷新 / 显示隐藏 / 打开配置 / 退出；单击托盘图标唤出到前台（显示/隐藏走右键菜单，双击不再切换）
- **窗口定位**: 右下角展开不越屏（四角智能翻转，多显示器 / DPI 缩放适配），收起精确回到原位
- 关闭到托盘，不退出进程

## 支持的 Provider

| Provider | 凭证 | 类型 | 展示内容 |
|----------|------|------|----------|
| [Kimi](https://platform.kimi.com) | API Key (`sk-kimi-xxx`) | usage | 5 小时窗口 / 周额度用量 |
| [讯飞星辰](https://maas.xfyun.cn) | Cookie | usage | 套餐总量用量 |
| [OpenCode Go](https://opencode.ai/zen) | API Key (`oc_sk-...`,留空自动读 `~/.local/share/opencode/auth.json`) | usage | 滚动窗口额度百分比 |
| [小米 MiMo](https://platform.xiaomimimo.com) | Cookie | usage | 套餐 Token 用量 |
| [DeepSeek](https://www.deepseek.com/platform) | API Key (`sk-...`) | balance | 账户余额（余额型） |
| [Ollama](https://ollama.com/settings) | 浏览器会话 Cookie | usage | Cloud 5 小时 Session 用量 / 周用量（设置页） |
| [Command Code](https://commandcode.ai) | API Key (`user_...`,留空自动读 `~/.commandcode/auth.json`) | usage | 5 小时 / 周滚动窗口用量 + 剩余 credits |
| [Factory Droid](https://factory.ai) | API Key (`fk-...`,留空自动读 `~/.factory/.env`) | usage | 5 小时 / 周 / 月窗口 + Core 池用量 |
| [MiniMax](https://platform.minimaxi.com) | 订阅 Key (`sk-cp-...`) | usage | 5 小时 / 周窗口用量（Token Plan） |
| [Open Design](https://open-design.ai) | 浏览器会话 Cookie | usage | Design Plan 5 小时 / 7 天额度用量 |

> 提示：Cookie 类凭证支持在浏览器 F12 → 网络标签 → 复制请求头后，直接粘贴 "Copy as PowerShell" 格式的整段脚本，保存时自动解析为 Cookie 请求头。

## 命令行探针

除图形界面外还有一个无界面探针，方便脚本、状态栏和外部工具对接：

```bash
go build -o build/bin/quota-viewer-cli.exe ./cmd/quotaviewer
```

- `quota-viewer-cli.exe --json` —— 向 stdout 打**一行 JSON**，包含全部平台的额度，退出码 `0`
- 不带参数，或参数无法识别 —— 向 stderr 打 usage，stdout 无输出，退出码 `2`

## Open Design 怎么监控

[Open Design](https://open-design.ai) 是设计 agent 工作台，额度来自 **Design Plan** 订阅，按 **5 小时窗口与 7 天窗口**计。

Quota Viewer 从 Open Design 的控制面读取这两个窗口，而它是唯一**不能**用 API Key 的一家：`sk-amr-` 形式的 key 请求控制面一律返回 `401`，**只能用浏览器会话 Cookie**。

取法：在 `open-design.ai/cloud/dashboard` 登录后，打开 F12 → 网络标签，把请求按 Cookie 复制（或整段用 "Copy as PowerShell" 复制后直接粘贴），粘进配置面板的 Open Design 卡片即可。会话过期后重新登录并替换 Cookie。

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
├── cmd/
│   └── quotaviewer/     # 无界面命令行探针入口（--json）
├── internal/
│   ├── config/          # 配置持久化（动态 Provider 列表 + PowerShell Cookie 解析）
│   ├── fetcher/         # Provider 注册表 + 各平台抓取器
│   ├── probe/           # 命令行探针实现（与 GUI 共用抓取编排）
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

本项目是**非官方**工具，与各平台无任何关联。各平台网页结构与接口可能随时变动导致抓取失效；凭证仅保存在本机配置文件中（`%APPDATA%\quota-viewer\config.json`），不会上传到任何服务器。请自行承担使用风险。

## 许可证

[MIT](LICENSE)
