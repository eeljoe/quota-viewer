# Quota Viewer

[English](README.md) | **[简体中文](README.zh-CN.md)**

桌面悬浮球 + AI 平台额度监控工具：实时显示各 AI 平台的 API 配额 / 余额剩余。

- **技术栈**: Go + Wails v2.12.0 + 原生 HTML/CSS/JS (Vite 打包)
- **平台**: Windows 10+ (WebView2 运行时)

## 功能

- **悬浮球**: 60×60 玻璃质感方块，1-3 格展示启用的 Provider（格字颜色 = 状态色：绿 / 黄 / 红），悬停 tooltip 显示各平台明细
- **自适应球格**: 只启用 1 个 Provider 时字母占满整球，2 个时各占 1/2，3 个时各占 1/3
- **展开面板**: 进度条 + 剩余量明细（千分位数字），自动刷新（默认 15 分钟）
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

1. 在 `internal/fetcher/` 新增抓取器（实现 `Fetcher` 接口，`Fetch() QuotaResult`），带 httptest 测试
2. 在 `internal/fetcher/registry.go` 注册：ID、显示名、球格缩写、凭证字段定义、`Build` 工厂
3. 前端配置面板与球格自动适配，无需改动 UI 代码

## 免责声明

本项目是**非官方**工具，与各平台无任何关联。各平台网页结构与接口可能随时变动导致抓取失效；凭证仅保存在本机配置文件中（`%APPDATA%/quota-viewer/config.json`），不会上传到任何服务器。请自行承担使用风险。

## 许可证

[MIT](LICENSE)
