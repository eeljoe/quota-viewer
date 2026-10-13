# 构建与测试基线

**When to read**: 构建项目、诊断构建问题、确认基线状态时。

---

## 核心内容

### 当前基线（2026-10-13，新增无界面探针后 `6f51b2b` 工作区状态）

- **Go 测试**：`go test ./... -count=1` 全绿（exit 0；顶层 PASS 145 条 + 子测试 PASS 32 条、0 FAIL；config ≈0.1s、fetcher ≈3.9s、probe ≈0.1s。数字取自本票返工 r2 后的工作区实测；`internal/fetcher` 侧另有并发提交，条数会随其变动）
  - `quota-viewer`（无测试文件）
  - `cmd/quotaviewer`（入口包，逻辑在 `internal/probe`，无测试文件）
  - `internal/config` ok
  - `internal/fetcher` ok
  - `internal/probe` ok（23 条：契约单测 + `reset_at` 归一 + 入口 `Run`）
  - `internal/tray`（无测试文件）
- **构建**：`go build ./...` exit 0；`go build -o build/bin/quota-viewer-cli.exe ./cmd/quotaviewer` exit 0（console 子系统）
- **工作区**：Go/前端源码无未提交改动（同步当次 `git status` 只有 `docs/` 文档与未跟踪目录的变动）
- **本次未跑 `wails build`**：这一轮只做 wiki 同步，构建产物未重新验证（上一次构建验证见附录 99 的会话记录）

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

# 4. 全包编译校验（含 cmd/ quotaviewer 与 internal/*,不产 GUI 产物）
go build ./...
```

### 无界面探针（`--json`）

```bash
# 构建:console 子系统产物,名称与命令行开关固定
go build -o build/bin/quota-viewer-cli.exe ./cmd/quotaviewer

# 运行(正常路径):一行合法 JSON 打到 stdout 并退出 0
build/bin/quota-viewer-cli.exe --json

# 无参 / 未知参数:stderr 打 usage 并退出 2(不进 GUI、不抓取)
build/bin/quota-viewer-cli.exe
```

- `go build ./...` 编译全部包（含 `cmd/quotaviewer`）exit 0；不加 `-H windowsgui`，产物就是 `Subsystem=3`（console），stdout / stderr / 退出码在 `Start-Process -Wait -PassThru -Redirect*` 与 `node spawnSync` 下均可读。
- 退出码（R1 口径）：**无参 / 未知参数 → stderr 打 usage + 退出 `2`**（不进 GUI、不抓取）；**`--json` 调用内**的某家抓取失败或配置读取失败 → 失败信封（`providers: []` + 顶层 `error` 带文本）+ **退出 `0`**，业务失败不体现在退出码上（spec 约束 8：绝不允许空 stdout + exit 0）；唯一的非 0 例外是**写 stdout 失败 → 退出 `1`**。
- `--json` 输出一行合法 JSON（内部无换行）。
- 契约与字段定义见 `internal/probe/probe.go`（顶层 7 键 + 每条记录 8 键）。

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
