package fetcher

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DevinFetcher 通过 Devin Desktop 自带的席位管理接口获取日/周额度。
//
// 为什么走私有接口:Devin 没有可用的公开额度 API。v2/v3 的
// /enterprise/consumption/* 只对 Enterprise 管理员的个人 Key 开放,
// self-serve(Free/Pro/Max/Teams)一律 403,官方文档也写明没有开关可以放开。
// 能拿到日/周两个窗口的只有 Devin Desktop(Windsurf 内核)自己调的
// exa.seat_management_pb.SeatManagementService/GetUserStatus —— CONNECT 协议 POST,
// 返回两侧的剩余百分比与各自的重置时间。
//
// 私有接口无稳定性承诺:字段改名/新增/改形态都可能随时发生。失效时对照
// Devin Desktop 的 resources/app/node_modules/@exa/chat-client/index.js 里
// GetUserStatus / PlanStatus 两个类型定义更新本文件(与官方前端 tgY / tgK 两个
// 换算函数同口径:已用 = 100 − 可见窗口剩余)。
//
// 凭证:api_key 手填优先;留空时读本机 Devin Desktop 的登录态作兜底
// (state.vscdb 里 token 与 apiServerUrl 都是明文),所以登录过桌面版的用户
// 什么都不填也能出数(与 factory-droid 读 ~/.factory/.env 同形)。
type DevinFetcher struct {
	apiKey    string
	serverURL string // 调用方注入(测试注入 httptest);留空 = 走默认端点
	stateFile string // state.vscdb 路径,构造期一次解析
	// readDesktopState 是「是否允许读本机登录态」:api_key 留空时是否去读
	// Devin Desktop 的登录态。默认 true;关掉就等于只认手填 token。
	readDesktopState bool
}

// NewDevinFetcher 创建一个 DevinFetcher。
//
// serverURL 故意不在这里填默认值:默认端点只由 devinStatusURL 兜底,否则票 02
// 从本机读到的 apiServerUrl 永远没有机会生效(字段已被写死)。
//
// stateFile 在这里一次解析完(与 factorydroid 的构造期解析同形):APPDATA 之后
// 再变也不影响本实例读哪份登录态,测试只要在构造前把 APPDATA 指到临时目录即可。
func NewDevinFetcher(apiKey string) *DevinFetcher {
	return &DevinFetcher{
		apiKey:           strings.TrimSpace(apiKey),
		stateFile:        defaultDevinStatePath(),
		readDesktopState: true,
	}
}

// devinMaxStateFileBytes 是本机登录态文件的体积上限(64 MB)。
// 实测 1.5 MB;超出这个数量级说明不是我们认识的 state.vscdb(或已被写坏),
// 与其把整个文件读进内存,不如当作读不到 —— 用户改走手填 token。
const devinMaxStateFileBytes = 64 << 20

// 本机凭据在 state.vscdb 里的两个字面量(实测明文)。
const (
	devinDesktopTokenLiteral  = "devin-session-token$"
	devinDesktopServerLiteral = `"apiServerUrl":"`
)

// defaultDevinStatePath 返回 Devin Desktop 的 state.vscdb 路径。
// APPDATA 为空时返回空串 —— 空串等于「读不到」,让 Fetch 直接走手填/未配置分支。
func defaultDevinStatePath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return ""
	}
	return filepath.Join(appData, "Devin", "User", "globalStorage", "state.vscdb")
}

// readDevinDesktopCredential 从 Devin Desktop 的 state.vscdb 里取出登录 token 与
// apiServerUrl。取不到(路径为空 / 文件不存在 / 空文件 / 没有关键字 / 超限)时
// 返回两个空串 —— 不报错、不打日志:读不到在用户侧的表现就是「未配置 token」,
// 不是「读取失败」,由 Fetch 给出一条可操作的提示。
//
// 为什么按字节扫:state.vscdb 是 VS Code 风格的小型 SQLite 库(键值表 ItemTable),
// 凭证明文存在数据页里。为免为此引一个 SQLite 驱动、也不解析库结构,这里直接在
// 整文件字节流里找两个字面量。如果哪天 Devin Desktop 改成加密存储,本函数会返回
// 两个空串,用户改走手填 token,不会崩、也不会把别人的数据当成 token。
//
// token 的返回值**带** devin-session-token$ 前缀(实测 metadata.api_key 要的就是
// 这个带前缀的整串)。两个结果互相独立:只有一个命中也返回另一个的空串。
func readDevinDesktopCredential(stateFile string) (token string, server string) {
	if stateFile == "" {
		return "", ""
	}
	info, err := os.Stat(stateFile)
	if err != nil || info.IsDir() {
		return "", ""
	}
	// 体积守卫必须在读取之前:超限文件一律不读,否则大文件会把内存顶爆。
	if info.Size() > devinMaxStateFileBytes {
		return "", ""
	}
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return "", ""
	}

	if i := bytes.Index(data, []byte(devinDesktopTokenLiteral)); i >= 0 {
		start := i
		end := start + len(devinDesktopTokenLiteral)
		for end < len(data) && isDevinTokenByte(data[end]) {
			end++
		}
		// 字面量后面必须还有 payload(票面 [A-Za-z0-9_.\-]+ 的 "+"):
		// 只有一个裸前缀不构成一个可用的 token。
		if end > start+len(devinDesktopTokenLiteral) {
			token = string(data[start:end])
		}
	}

	if i := bytes.Index(data, []byte(devinDesktopServerLiteral)); i >= 0 {
		start := i + len(devinDesktopServerLiteral)
		if j := bytes.IndexByte(data[start:], '"'); j >= 0 {
			server = string(data[start : start+j])
		}
	}
	return token, server
}

// isDevinTokenByte 判定 token payload 的合法字节:[A-Za-z0-9_.\-]。
func isDevinTokenByte(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	case c == '_', c == '.', c == '-':
		return true
	}
	return false
}

// devinResolveServer 按 AS-8 的顺序决定打到哪个 server(纯函数):
// 注入值优先 → 本机读到的 apiServerUrl → 空串。空串由 devinStatusURL 兜底成默认端点;
// 默认值只在那一处出现,绝不提前写进 f.serverURL(写进去本机值就再也无法生效)。
func devinResolveServer(injected, local string) string {
	if injected != "" {
		return injected
	}
	return local
}

const (
	// devinDefaultServer 是默认 API 服务器(本机 windsurfAuthStatus.apiServerUrl 即此值)。
	devinDefaultServer = "https://server.codeium.com"
	// devinStatusPath 是席位管理服务里查个人额度的路由。
	devinStatusPath = "/exa.seat_management_pb.SeatManagementService/GetUserStatus"
	// devinUserAgent 是浏览器型固定 UA:该路由挂在官方客户端网关后面,
	// 非浏览器型 UA 有被拦的风险(与 Ollama 的做法一致)。
	devinUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"
	// devinDesktopVersion 是实测可用的 Devin Desktop 版本,接口会校验它。
	devinDesktopVersion = "1.126.0"
)

// devinStatusURL 拼出查询额度的最终 URL(纯函数,便于直接断言字符串)。
// server 为空时走默认端点;非空则去掉尾部斜杠后拼同一路径。
func devinStatusURL(server string) string {
	if server == "" {
		return devinDefaultServer + devinStatusPath
	}
	return strings.TrimRight(server, "/") + devinStatusPath
}

// devinJSONInt64 是容错的 int64:protojson 会把 int64 序列化成字符串
// (实测 dailyQuotaResetAtUnix / weeklyQuotaResetAtUnix / overageBalanceMicros
// 都是 "1793260800" 这种),但不同客户端版本也可能给数字。三种形态都要吃下:
// 数字、数字字符串、null。
type devinJSONInt64 int64

func (v *devinJSONInt64) UnmarshalJSON(raw []byte) error {
	s := strings.TrimSpace(string(raw))
	if s == "null" {
		*v = 0
		return nil
	}
	s = strings.Trim(s, `"`)
	if s == "" {
		*v = 0
		return nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		*v = devinJSONInt64(n)
		return nil
	}
	// 兜底:整数值被写成 1.7932608e9 之类的浮点形态。
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("不是合法的整数: %q", s)
	}
	*v = devinJSONInt64(int64(f))
	return nil
}

// devinJSONFloat 是容错的浮点:百分比正常是 JSON 数字,但被序列化成字符串
// ("35.5")时也要能解析,不缺字段时不制造错误。
type devinJSONFloat float64

func (v *devinJSONFloat) UnmarshalJSON(raw []byte) error {
	s := strings.TrimSpace(string(raw))
	if s == "null" {
		return nil
	}
	s = strings.Trim(s, `"`)
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("不是合法的数字: %q", s)
	}
	*v = devinJSONFloat(f)
	return nil
}

// devinPlanStatus 是解析后的额度快照。
// 两个 Remaining 字段语义是**剩余**百分比(与接口字段名一致),
// 换算成"已用"由 applyPlanStatus 负责,避免解析层埋反向语义。
type devinPlanStatus struct {
	DailyRemaining  *float64 // dailyQuotaRemainingPercent,语义 = 剩余
	WeeklyRemaining *float64
	DailyResetAt    int64 // dailyQuotaResetAtUnix,epoch 秒,缺失 = 0
	WeeklyResetAt   int64
	HideDaily       bool // planInfo.hideDailyQuota
	HideWeekly      bool
	OverageMicros   *int64 // overageBalanceMicros,缺失 = nil
	PlanName        string // planInfo.planName,只解析不展示
}

// devinRawPlanStatus 是线上响应的直接映射(用容错类型接字段),由 devinParsePlanStatus 翻译成 devinPlanStatus。
type devinRawPlanStatus struct {
	DailyQuotaRemainingPercent  *devinJSONFloat `json:"dailyQuotaRemainingPercent"`
	WeeklyQuotaRemainingPercent *devinJSONFloat `json:"weeklyQuotaRemainingPercent"`
	DailyQuotaResetAtUnix       devinJSONInt64  `json:"dailyQuotaResetAtUnix"`
	WeeklyQuotaResetAtUnix      devinJSONInt64  `json:"weeklyQuotaResetAtUnix"`
	OverageBalanceMicros        *devinJSONInt64 `json:"overageBalanceMicros"`
	PlanInfo                    *struct {
		PlanName        string `json:"planName"`
		HideDailyQuota  bool   `json:"hideDailyQuota"`
		HideWeeklyQuota bool   `json:"hideWeeklyQuota"`
	} `json:"planInfo"`
}

// devinGetUserStatusResp 是 GetUserStatus 的响应外壳:额度都在 userStatus.planStatus 里。
type devinGetUserStatusResp struct {
	UserStatus *struct {
		PlanStatus *devinRawPlanStatus `json:"planStatus"`
	} `json:"userStatus"`
}

// devinParsePlanStatus 把响应体翻译成 devinPlanStatus(纯函数,供单测直接喂字节)。
// 缺 userStatus / 缺 planStatus / 两个窗口的剩余百分比都缺 ⇒ 明确报错,
// 不 panic、也不当作 0% 已用(那样会静默渲染成"额度全新")。
func devinParsePlanStatus(raw []byte) (*devinPlanStatus, error) {
	const unparsable = "无法解析额度数据（响应结构可能已变化）"

	var resp devinGetUserStatusResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, errors.New(unparsable)
	}
	if resp.UserStatus == nil || resp.UserStatus.PlanStatus == nil {
		return nil, errors.New(unparsable)
	}
	ps := resp.UserStatus.PlanStatus
	if ps.DailyQuotaRemainingPercent == nil && ps.WeeklyQuotaRemainingPercent == nil {
		return nil, errors.New(unparsable)
	}

	st := &devinPlanStatus{
		DailyResetAt:  int64(ps.DailyQuotaResetAtUnix),
		WeeklyResetAt: int64(ps.WeeklyQuotaResetAtUnix),
	}
	if ps.DailyQuotaRemainingPercent != nil {
		v := float64(*ps.DailyQuotaRemainingPercent)
		st.DailyRemaining = &v
	}
	if ps.WeeklyQuotaRemainingPercent != nil {
		v := float64(*ps.WeeklyQuotaRemainingPercent)
		st.WeeklyRemaining = &v
	}
	if ps.OverageBalanceMicros != nil {
		v := int64(*ps.OverageBalanceMicros)
		st.OverageMicros = &v
	}
	if ps.PlanInfo != nil {
		st.PlanName = ps.PlanInfo.PlanName
		st.HideDaily = ps.PlanInfo.HideDailyQuota
		st.HideWeekly = ps.PlanInfo.HideWeeklyQuota
	}
	return st, nil
}

// devinMetadata 是 CONNECT 请求的 metadata 块。官方桌面端就是这么自报身份的,
// 字段值照抄实测可用的那套;session_id 只用占位值,不引第三方包生成 UUID。
type devinMetadata struct {
	APIKey           string `json:"api_key"`
	IDEName          string `json:"ide_name"`
	IDEVersion       string `json:"ide_version"`
	ExtensionName    string `json:"extension_name"`
	ExtensionVersion string `json:"extension_version"`
	SessionID        string `json:"session_id"`
	Locale           string `json:"locale"`
	OS               string `json:"os"`
}

// fetchPlanStatus 发起一次 GetUserStatus 并解析。状态码映射:
// 401/403 = token 失效,其余非 200 透出状态码(AS-9)。
// server 是调用方已按 AS-8 顺序解析好的值:空串 ⇒ devinStatusURL 兜底到默认端点。
func (f *DevinFetcher) fetchPlanStatus(client *http.Client, token, server string) (*devinPlanStatus, error) {
	payload := struct {
		Metadata devinMetadata `json:"metadata"`
	}{Metadata: devinMetadata{
		APIKey:           token,
		IDEName:          "windsurf",
		IDEVersion:       devinDesktopVersion,
		ExtensionName:    "windsurf",
		ExtensionVersion: devinDesktopVersion,
		SessionID:        strconv.FormatInt(time.Now().UnixMilli(), 10),
		Locale:           "en",
		OS:               "win32",
	}}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", devinStatusURL(server), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connect-Protocol-Version", "1")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", devinUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return nil, errors.New("Devin token 已失效，请重新登录 Devin Desktop 或更新 token")
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	return devinParsePlanStatus(raw)
}

// devinWindow 是一个可见额度窗口的展示素材:名称、已用百分比与重置 epoch 秒。
type devinWindow struct {
	label   string
	used    float64
	resetAt int64
}

// applyPlanStatus 把额度快照映射成 QuotaResult(票面 AS-3..AS-7)。
//
// 口径:已用 = 100 − 剩余,夹取复用同包的 clampPercent;可见窗口里已用更大者驱动
// 百分比与倒计时(并列取日);驱动窗口没有重置时间时 ResetAt/ResetLabel 都留空,
// 不借另一个窗口的时间(借了会把"日额度 3 小时后重置"错标成周的时间)。
func (f *DevinFetcher) applyPlanStatus(st *devinPlanStatus) QuotaResult {
	result := QuotaResult{
		Platform:    "Devin",
		Kind:        KindUsage,
		LastUpdated: time.Now(),
	}

	// 可见窗口 = 未被套餐隐藏 且 本次响应确实给了剩余百分比。
	var windows []devinWindow
	if !st.HideDaily && st.DailyRemaining != nil {
		windows = append(windows, devinWindow{
			label:   "日",
			used:    clampPercent(100 - *st.DailyRemaining),
			resetAt: st.DailyResetAt,
		})
	}
	if !st.HideWeekly && st.WeeklyRemaining != nil {
		windows = append(windows, devinWindow{
			label:   "周",
			used:    clampPercent(100 - *st.WeeklyRemaining),
			resetAt: st.WeeklyResetAt,
		})
	}
	if len(windows) == 0 {
		result.Error = "当前套餐不提供可读的额度窗口"
		return result
	}

	// 驱动窗口:已用更大者;并列时保留排头(日在前)。
	driver := windows[0]
	for _, w := range windows[1:] {
		if w.used > driver.used {
			driver = w
		}
	}
	result.Total = 100
	result.Used = driver.used
	result.Percent = driver.used

	// 倒计时只跟驱动窗口走:它才是决定球色与进度条的那个窗口。
	if driver.resetAt > 0 {
		result.ResetAt = time.Unix(driver.resetAt, 0).UTC().Format(time.RFC3339)
		result.ResetLabel = driver.label + "额度"
	}

	// Remaining 每段一个可见窗口,按已用降序(并列日在前):头部只有 210px 宽,
	// 最紧张的数字排在后面会被截掉,球红了却看不到原因(与 factory-droid 同口径)。
	segs := make([]devinWindow, len(windows))
	copy(segs, windows)
	sort.SliceStable(segs, func(i, j int) bool { return segs[i].used > segs[j].used })
	texts := make([]string, 0, len(segs)+1)
	for _, s := range segs {
		texts = append(texts, fmt.Sprintf("%s %.1f%% 已用", s.label, s.used))
	}
	// 超额余额是 micros(百万分之一美元);只 > 0 才追加,套餐名一律不出现。
	if st.OverageMicros != nil && *st.OverageMicros > 0 {
		texts = append(texts, fmt.Sprintf("超额金额 $%.2f", float64(*st.OverageMicros)/1e6))
	}
	result.Remaining = strings.Join(texts, " · ")
	return result
}

// Fetch 拉取 Devin 的日/周额度。
//
// 开头就构造 QuotaResult 的三项身份(Platform/Kind/LastUpdated):错误路径只补 Error,
// 那三项必须保留 —— 上层按 Platform 对齐球格,零值会让错误提示无处可落。
//
// 凭证与端点按 AS-8 的顺序解析(票 02 的增量):
//   - token:手填(TrimSpace 后)优先 → 本机 Devin Desktop 登录态 → 报「未配置」且不发请求;
//   - server:注入值优先 → 本机读到的 apiServerUrl → 默认端点(只由 devinStatusURL 兜底)。
//
// 本机 token 与 server 来自同一次 readDevinDesktopCredential 调用,不重复读文件。
// 读取条件是「readDesktopState 为真 且(手填 token 为空 或 未注入 server)」:生产构造点
// 只传 api_key、永不注入 serverURL(registry.go 的 NewDevinFetcher(creds["api_key"]))⇒
// 该条件在生产上恒真,每次 Fetch 都会把整份 state.vscdb(实测约 1.5 MB)读一遍。
// 也就是说手填 token 只省掉「token 取值」这一步,并不省掉读文件;只有注入 server 的
// 调用方(测试)在 token 也非空时才真的不碰用户的文件。
func (f *DevinFetcher) Fetch() QuotaResult {
	result := QuotaResult{
		Platform:    "Devin",
		Kind:        KindUsage,
		LastUpdated: time.Now(),
	}

	token := strings.TrimSpace(f.apiKey)
	var localToken, localServer string
	if f.readDesktopState && (token == "" || f.serverURL == "") {
		localToken, localServer = readDevinDesktopCredential(f.stateFile)
	}
	if token == "" {
		token = localToken
	}
	if token == "" {
		result.Error = "未配置 Devin token（登录 Devin Desktop 后留空即可，或粘贴 token）"
		return result
	}
	// 注意:本机 server 只作为本次请求的参数,绝不回写 f.serverURL
	// (写进去等于把它固化,换机或手填 server 就再也不生效)。
	server := devinResolveServer(f.serverURL, localServer)

	client := &http.Client{Timeout: 10 * time.Second}
	st, err := f.fetchPlanStatus(client, token, server)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	return f.applyPlanStatus(st)
}
