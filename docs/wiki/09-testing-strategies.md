# 测试策略

**When to read**: 决定改动的验证方式、新增测试时。

---

## 核心内容

### 验证阶梯（由快到慢）

| 阶梯 | 命令 | 覆盖 |
|---|---|---|
| 1. Go 单测 | `go test ./...` | config、fetcher 全部逻辑 |
| 2. 前端构建 | `cd frontend && npm run build` | 前端语法/打包正确性 |
| 3. 完整构建 | `wails build` | 绑定生成 + embed + 打包 |
| 4. 手动冒烟 | 运行 exe | 窗口定位、托盘、真实抓取（需真实凭证） |

### 测试分类

| 类别 | 位置 | 模式 |
|---|---|---|
| 配置持久化 | `internal/config/config_test.go` | Load 默认值 / Save 往返（含 Budget）/ 旧格式迁移（含 mimo_cookie 与 4 平台钳制）/ 新 Provider 追加迁移 / AllProviderIDs 同步守护 / 超限启用钳制（普通 3、扩展 9） |
| 抓取器 | `internal/fetcher/*_test.go` | `net/http/httptest` 假服务 + baseURL 注入；成功/失败/异常 JSON 路径 |
| 注册表 | `internal/fetcher/registry_test.go` | 11 个 Provider、顺序稳定、字段定义完整、Build 可执行不 panic |
| 预算换算 | `internal/fetcher/deepseek_test.go` 的 TestApplyBudget_* | 5 个用例：正常换算/默认预算/超额钳制/用量型 no-op/错误结果 no-op |
| 托盘 | 无测试 | 依赖 GUI，手工验证 |
| 窗口定位 | 无测试 | 依赖真实显示器环境，手工验证（多屏/DPI 需实测） |
| 前端 | 无测试 | 手工验证（当前无测试框架） |

**当前规模（2026-10-13，`go test ./... -count=1 -v` 实测）**：顶层 PASS 150 条 + 子测试 PASS 37 条，0 FAIL；耗时 config ≈0.18s、fetcher ≈4.2s、probe ≈0.11s。包级分布：`internal/fetcher` 113 顶层、`internal/config` 10 顶层、`internal/probe` 27 顶层；`cmd/quotaviewer`、根包与 `internal/tray` 无测试文件。（2026-10-11 的旧口径为 120 条；逐家 Provider 的子分布本次未重算——`opencode_go` 与 `factorydroid` 两家的用例在 2026-10-13 各重写/大幅新增过。）

### 约定

- fetcher 测试通过构造参数 `baseURL`/`apiURL` 指向 httptest server，**不发真实网络请求**
- 新平台抓取器必须带测试（现有 kimi/xfyun/opencode_go/mimo/deepseek/ollama/commandcode/factorydroid/minimax/opendesign/devin 十一家均有）
- 新增 Provider 时 registry_test 自动校验定义完整性
- 配置结构变更必须同步 config_test.go（含迁移用例）
- 余额型 Provider 必须覆盖 ApplyBudget 用例（换算/钳制/no-op 路径）
- 前端无测试框架——UI 改动以构建 + 手动冒烟为准

---

## 关键文件

| 文件 | 职责 |
|---|---|
| `internal/config/config_test.go` | 配置与 Cookie 解析用例 + 清单同步守护 + 超限钳制 |
| `internal/fetcher/kimi_test.go` / `xfyun_test.go` / `opencode_go_test.go` / `mimo_test.go` / `deepseek_test.go` / `ollama_test.go` / `commandcode_test.go` / `factorydroid_test.go` / `minimax_test.go` / `opendesign_test.go` / `devin_test.go` | 抓取器 httptest 用例 |
| `internal/fetcher/registry_test.go` | 注册表完整性（11 家、顺序、字段定义、空凭证 Build 不 panic） |

---

## Must NOT Change

- 测试不得依赖真实网络/真实凭证（CI 与本地均不可行）
- 新增 fetcher 必须遵循 baseURL 注入模式，保证可测性
