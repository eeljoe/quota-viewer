# Agent Rules

**When to read**: 新 agent 接手本项目时的第一份文件。开工前必读。

---

## 接手约束

- **本项目是一个 Windows 桌面应用**（Wails v2 + WebView2），开发与测试在 Windows 上进行；`workarea_other.go` 仅是非 Windows 编译桩，不要依赖它做验证
- **未提交工作区**：会话开始前遗留的 MiMo → OpenCode Go 替换 + fitToScreen 窗口定位修复尚未提交（见 STATUS.md），动工作区前先与用户确认提交策略
- 修改前端后必须重新构建产物（`frontend/dist/` 被 Go embed，见 10-build-test-baseline.md）
- 修改 Go 绑定的方法签名后必须同步 `frontend/wailsjs/`（`wails generate module` 或 `wails dev/build` 时自动生成）
- 平台抓取器不得存储明文之外的凭证格式变化（配置项 JSON key 变更 = 破坏性变更，需迁移策略）
- 配置文件中可能包含真实 API Key / Cookie —— 任何文档、日志、提交中不得写入真实凭证

## 验证命令

| 命令 | 用途 |
|---|---|
| `go test ./...` | 全部 Go 测试（当前全绿） |
| `cd frontend && npm run build` | 构建前端产物到 `frontend/dist/` |
| `wails build` | 完整构建 exe（产物 `build/bin/quota-viewer.exe`） |
| `wails dev` | 开发模式（热重载） |

## 目录速览

| 路径 | 作用 |
|---|---|
| `app.go` | Wails 主应用（窗口、动态 Provider 编排、配置） |
| `main.go` | 入口 + Wails 运行时选项 |
| `workarea_windows.go` | Win32 原生调用（DPI/工作区/窗口子类） |
| `internal/config/` | 配置持久化（动态 Provider 列表）+ Cookie 解析 |
| `internal/fetcher/` | Provider 注册表 + 五平台抓取器 |
| `internal/tray/` | 系统托盘 |
| `frontend/src/` | 前端源码（index.html / main.js / style.css） |
| `frontend/dist/` | 构建产物（Go embed） |

---

<!-- wiki-meta: {"synced_commit": "e60f844", "synced_at": "2026-08-04T05:31:55Z", "synced_files": 11} -->

> **Wiki 同步状态**：最后同步于 `e60f844`（2026-08-04 05:31），共 11 个文件。⚠️ 内容反映工作区当前状态（含未提交的 OpenCode Go 替换）。

## Wiki 文件索引

| 编号 | 文件 | 主题 |
|---|---|---|
| 00 | agent-rules | 接手约束、验证命令、索引（本文件） |
| 01 | architecture-overview | 架构总览、分层职责、数据流 |
| 02 | module-catalog | 全部模块职责与行数 |
| 03 | core-pipeline | 刷新调用链、事件总线契约 |
| 04 | window-positioning | 窗口定位、DPI、工作区、fitToScreen |
| 05 | fetching-platforms | 五平台抓取器、注册表与 QuotaResult |
| 06 | systray | 系统托盘实现与线程铁律 |
| 07 | config-model | 动态 Provider 配置模型与旧格式迁移 |
| 08 | api-contract | Wails 绑定 API（动态配置契约） |
| 09 | testing-strategies | 验证阶梯与测试分类 |
| 10 | build-test-baseline | 构建与测试基线 |
| 99 | appendix-legacy-status | 旧 STATUS 会话归档（按需创建） |

---

## Must NOT Change

- `ballSize`/`screenMargin` 常量（app.go）与前端 `SIZES.ball` 的一致性——改了必须两边同步
- `tray:refresh` / `tray:toggle` / `tray:settings` 事件名——托盘与 app.go 两端共用契约
- `quota:update` / `ui:show-settings` 事件名——后端推给前端的契约
- `QuotaResult` JSON 字段名（platform/id/abbr/kind/used/total/percent/remaining/reset_at/error）——前端渲染依赖
- Provider id（kimi/xfyun/opencode-go/mimo/deepseek）与注册表字段定义——config 存储、TestConnection 参数、前端绑定共用契约
- 展示上限 3 / 下限 1 个 Provider——前后端都钳制
- 配置 JSON 结构（providers 数组）——已迁移用户配置兼容
