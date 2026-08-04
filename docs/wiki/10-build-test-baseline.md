# 构建与测试基线

**When to read**: 构建项目、诊断构建问题、确认基线状态时。

---

## 核心内容

### 当前基线（2026-08-04，工作区状态）

- **Go 测试**：`go test ./...` 全绿
  - `quota-viewer`（无测试文件）
  - `internal/config` ok
  - `internal/fetcher` ok
  - `internal/tray`（无测试文件）
- **工作区**：dirty（15 项变更，MiMo→OpenCode 替换 + fitToScreen 修复未提交），未验证 `wails build` 产物

### 构建流程

```bash
# 1. Go 依赖
go mod tidy

# 2. 前端产物（必须！Go embed 编译进 exe）
cd frontend && npm install && npm run build && cd ..

# 3. 完整构建
wails build        # 产物: build/bin/quota-viewer.exe

# 开发模式（热重载）
wails dev
```

### 产物与资源

| 项 | 位置 | 说明 |
|---|---|---|
| exe | `build/bin/quota-viewer.exe` | Wails 构建产物 |
| 前端产物 | `frontend/dist/` | `main.go` 里 `go:embed all:frontend/dist` |
| wailsjs 绑定 | `frontend/wailsjs/` | 构建时生成，勿手改 |
| 托盘图标 | `internal/tray/assets/` | embed 进二进制 |
| 应用图标 | `build/appicon.png` | 打包用 |

### 构建注意事项

- 忘构建前端直接 `wails build` → exe 里是旧 UI（embed 的是 dist 内容）
- `wails.json` 指定 `frontend:build: npm run build`，`wails build` 会自动跑前端构建
- 前端构建警告（dist/wailsjs LF→CRLF）不影响产物
- 版本：Wails v2.12.0（go.mod），WebView2 运行时（目标机需安装）

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `wails.json` | Wails 配置（名称/产物名/前端命令） |
| `main.go` | embed 与运行时选项 |
| `frontend/vite.config.js` | Vite 构建配置 |
| `go.mod` | Go 依赖（wails/v2 v2.12.0、energye/systray 等） |

---

## Must NOT Change

- `frontend/dist/` 与 `frontend/wailsjs/` 必须与源码同步提交（生成物也是交付物）
- embed 指令 `all:frontend/dist` 不可去掉（否则 UI 丢失）
