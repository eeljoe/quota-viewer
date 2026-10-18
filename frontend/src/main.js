// === Wails 绑定 ===
// window.go.main.App 在运行时由 Wails 注入

// 各视图窗口尺寸,与 Go 侧 ballSize 常量保持一致
const SIZES = {
    ball: [60, 60],
    panel: [340, 310],
    settings: [340, 480],
};

let currentView = "ball"; // ball | panel | settings
let currentResults = [];
let providerCards = []; // [{id, enabled: checkbox, groups: [{key, fields:[{key, input}]}]}]
let currentExtendedMode = false; // 扩展模式:悬浮球最多 16 格(普通模式 9)
let lastPanelH = SIZES.panel[1]; // 面板当前窗口高度(随条目数伸缩)

// 悬浮球可显示的监控对象上限(与后端 config.MaxObjects 同口径):
// 一个监控对象 = 一组凭证 × 一张套餐卡,多 Key / 多套餐各占一格。
function maxEnabled() {
    return currentExtendedMode ? 16 : 9;
}

// === 视图切换(统一入口,负责窗口尺寸与屏幕内定位) ===
function setView(view) {
    currentView = view;
    document.getElementById("ball").classList.toggle("hidden", view !== "ball");
    document.getElementById("panel").classList.toggle("hidden", view !== "panel");
    document.getElementById("settings").classList.toggle("hidden", view !== "settings");

    if (view === "ball") {
        window.go.main.App.CollapseWindow();
    } else {
        const [w, h] = view === "panel" ? panelSize() : SIZES[view];
        lastPanelH = h;
        window.go.main.App.ExpandWindow(w, h);
    }
}

// 面板高度随条目数伸缩:条目 = 监控对象数(多 Key / 多套餐会变多),封顶 640(超出列表内部滚动)
function panelSize() {
    const n = Math.max(currentResults.length, 1);
    const h = Math.min(640, 96 + n * 66);
    return [SIZES.panel[0], Math.max(SIZES.panel[1], h)];
}

// === 事件监听 ===
window.runtime.EventsOn("quota:update", (results) => {
    currentResults = results;
    renderResults(results);
});

// === 球面点击:展开/收起 ===
document.getElementById("ball").addEventListener("click", () => {
    if (currentView !== "ball") return;
    setView("panel");
    refreshIfNeeded(); // 展开时若数据超过 3 分钟则刷新
});

// 收起按钮
document.getElementById("btn-collapse").addEventListener("click", () => {
    setView("ball");
});

// === 刷新 ===
document.getElementById("btn-refresh").addEventListener("click", () => {
    refreshQuota();
});

async function refreshQuota() {
    const btn = document.getElementById("btn-refresh");
    btn.disabled = true;
    btn.classList.add("spinning");
    document.getElementById("last-updated").textContent = "刷新中...";
    try {
        const results = await window.go.main.App.Refresh();
        currentResults = results;
        renderResults(results);
    } catch (e) {
        console.error("refresh error:", e);
        toast("刷新失败: " + e, "error");
    } finally {
        btn.disabled = false;
        btn.classList.remove("spinning");
    }
}

let lastRefreshTime = 0;
async function refreshIfNeeded() {
    if (Date.now() - lastRefreshTime > 3 * 60 * 1000) {
        await refreshQuota();
    }
}

// === 渲染结果 ===

// escapeHtml 转义平台名/套餐名里的 HTML 特殊字符(后端字段直接进 innerHTML)。
function escapeHtml(s) {
    return String(s == null ? "" : s).replace(/[&<>"']/g, (c) => (
        { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]
    ));
}

// 监控对象的展示名:平台名 + 可选后缀(套餐名 / "Key 2")。
function objectLabel(r) {
    return r.platform + (r.sub_label ? " · " + r.sub_label : "");
}

function renderResults(results) {
    // 更新详情面板
    const list = document.getElementById("quota-list");
    list.innerHTML = "";
    results.forEach((r, idx) => {
        const color = getStatusColor(r);
        const percent = r.percent || 0;
        const item = document.createElement("div");
        item.className = "quota-item" + (r.error ? " error" : "");
        item.style.animationDelay = idx * 45 + "ms";
        const resetHtml = (r.reset_at && !r.error)
            ? `<div class="quota-reset" data-reset-at="${r.reset_at}">${formatCountdown(r.reset_at)}</div>`
            : "";
        item.innerHTML = `
            <div class="quota-item-header">
                <span class="quota-platform"><i class="status-dot ${color}"></i>${escapeHtml(r.platform)}${r.sub_label ? `<em class="quota-sub">${escapeHtml(r.sub_label)}</em>` : ""}</span>
                <span class="quota-remaining">${r.error || r.remaining || "-"}</span>
            </div>
            <div class="progress-bar">
                <div class="progress-fill ${color}" style="width: ${r.error ? 100 : percent}%"></div>
            </div>
            ${resetHtml}
        `;
        list.appendChild(item);
    });

    // 更新球面格子
    updateBall(results);

    // 面板已展开且条目数变化时,同步窗口高度
    syncPanelSize();

    // 更新时间
    const now = new Date();
    lastRefreshTime = now.getTime();
    document.getElementById("last-updated").textContent = "更新于 " + now.toLocaleTimeString("zh-CN");
}

function syncPanelSize() {
    if (currentView !== "panel") return;
    const [, h] = panelSize();
    if (h !== lastPanelH) {
        lastPanelH = h;
        window.go.main.App.ExpandWindow(SIZES.panel[0], h);
    }
}

// === 倒计时 ===
function formatCountdown(isoStr) {
    const target = new Date(isoStr);
    if (isNaN(target.getTime())) return "";
    const diff = target - Date.now();
    if (diff <= 0) return "已过期";
    const hours = Math.floor(diff / 3600000);
    const mins = Math.floor((diff % 3600000) / 60000);
    if (hours > 0) return `距下次刷新: ${hours}时${mins}分`;
    return `距下次刷新: ${mins}分`;
}

function updateCountdowns() {
    document.querySelectorAll(".quota-reset").forEach((el) => {
        const resetAt = el.getAttribute("data-reset-at");
        const text = formatCountdown(resetAt);
        if (text) el.textContent = text;
    });
}

setInterval(updateCountdowns, 30000);

// 状态色阈值:≥60% 黄,≥80% 红,≥100% 熄灭(额度耗尽)
function getStatusColor(r) {
    if (r.error) return "red";
    // 余额型(如 DeepSeek):未设预算恒绿,设了预算后按消耗百分比走同一套阈值
    if (r.kind === "balance" && !(r.percent > 0)) return "green";
    if (r.percent >= 100) return "off";
    if (r.percent >= 80) return "red";
    if (r.percent >= 60) return "yellow";
    return "green";
}

// 球面格子 = 监控对象数量(1-16):一个监控对象 = 一组凭证 × 一张套餐卡。
// 1-3 单行条带,4 → 2x2,5-9 → 3 列,10-16 → 4 列;超过上限时末格显示 "+N"。
// 格字颜色=状态;悬停 tooltip 显示各对象明细。
function updateBall(results) {
    const ball = document.getElementById("ball");
    ball.querySelectorAll(".ball-cell").forEach((c) => c.remove());

    const cap = maxEnabled();
    let shown = results;
    let overflow = 0;
    if (results.length > cap) {
        // 宁可少显示一格也要把 "+N" 说清楚——静默丢弃会让用户以为那格不存在
        // (同屏 9 格上限踩过的坑:门面写"十家"却不说上限,用户勾满就消失)。
        overflow = results.length - (cap - 1);
        shown = results.slice(0, cap - 1);
    }

    const n = shown.length + (overflow > 0 ? 1 : 0);
    // 列数:1-3 单行 n 列;4 → 2x2;5-9 → 3 列;10-16 → 4 列
    const cols = n <= 3 ? n : (n <= 4 ? 2 : (n <= 9 ? 3 : 4));
    ball.classList.toggle("grid", n >= 4);
    // 列数交给 CSS(.ball.grid 按 --cols 等宽换行 + 末行居中)
    if (n >= 4) ball.style.setProperty("--cols", cols);
    // 窄格场景(3 格单行 / 5+ 格多列):双字符缩写要降字号,否则顶到格边、相邻两格粘一起
    ball.classList.toggle("narrow", n === 3 || n >= 5);

    shown.forEach((r) => {
        const cell = document.createElement("span");
        const abbr = r.abbr || r.platform.slice(0, 1);
        // wide 标记多字符缩写(MX / OD / Go / K2):窄格下由 CSS 降字号换净空
        cell.className = "ball-cell " + getStatusColor(r) + (abbr.length > 1 ? " wide" : "");
        cell.textContent = abbr;
        ball.appendChild(cell);
    });

    if (overflow > 0) {
        const cell = document.createElement("span");
        cell.className = "ball-cell more";
        cell.textContent = "+" + overflow;
        ball.appendChild(cell);
    }

    // 数量档位驱动字号(4+ 网格才需要缩字号,1-3 沿用默认)
    ball.classList.remove("cells-4-6", "cells-7-9", "cells-10-16");
    if (n >= 10) ball.classList.add("cells-10-16");
    else if (n >= 7) ball.classList.add("cells-7-9");
    else if (n >= 4) ball.classList.add("cells-4-6");

    // 单格时放大字母占满整个球
    ball.classList.toggle("single-cell", n === 1);

    const lines = results.map((r) => objectLabel(r) + ": " + (r.error || r.remaining || "未知"));
    if (overflow > 0) {
        lines.push(`另有 ${overflow} 个监控对象未上球(上限 ${cap} 格,可在扩展模式放到 ${16} 格),详情见展开面板`);
    }
    ball.title = lines.join("\n");
}

// === 配置面板 ===
document.getElementById("btn-settings").addEventListener("click", () => {
    setView("settings");
    loadConfig();
});

document.getElementById("btn-close-settings").addEventListener("click", () => {
    setView("ball");
});

async function loadConfig() {
    try {
        const cfg = await window.go.main.App.GetConfig();
        currentExtendedMode = !!cfg.extended_mode;
        document.getElementById("input-extended").checked = currentExtendedMode;
        renderProviderList(cfg.providers || []);
        document.getElementById("input-interval").value = cfg.refresh_interval_min || 15;
        updateObjectHint();
    } catch (e) {
        console.error("loadConfig error:", e);
        toast("加载配置失败: " + e, "error");
    }
}

// 一组凭证里是否算"已填":输入框有内容,或已有保存的凭证(掩码显示在 placeholder 上)。
// 注意不能只看 value —— GetConfig 回显的是掩码占位符,用户没重打时 value 是空的,
// 但后端 CredSets() 仍然认这组(判的是存下来的凭证),两边口径必须一致。
function groupFilled(group) {
    return group.fields.some((f) => f.input.value.trim() !== "" || (f.input.placeholder || "").trim() !== "");
}

// 该 Provider 实际占用的监控对象数:主组恒算 1,附加组只有填了内容才算
// ——与后端 config.CredSets() 的跳过规则保持一致,否则前端提示会虚高。
function cardObjectCount(card) {
    if (!card.enabled.checked) return 0;
    let n = 1;
    card.groups.slice(1).forEach((g) => {
        if (groupFilled(g)) n += 1;
    });
    return n;
}

// 当前勾选共占用多少格(面板提示用)。
function currentObjectCount() {
    return providerCards.reduce((sum, c) => sum + cardObjectCount(c), 0);
}

function updateObjectHint() {
    const el = document.getElementById("object-hint");
    if (!el) return;
    const cap = maxEnabled();
    const n = currentObjectCount();
    el.textContent = n > cap
        ? `已勾选 ${n} 个监控对象,超出悬浮球上限 ${cap} 格(扩展模式 16 格);展开面板仍会全部列出`
        : `悬浮球占用 ${n} / ${cap} 格(普通模式 9,扩展模式 16)`;
}

// 收集当前 UI 上的 Provider 状态(保存/测试共用)。
// creds = 主组;extra = 附加组(带稳定 key,后端据此对齐,删中间一行不会错位)。
function collectProviders() {
    return providerCards.map((c) => {
        const read = (group) => {
            const creds = {};
            group.fields.forEach((f) => {
                creds[f.key] = f.input.value;
            });
            return creds;
        };
        const extra = c.groups.slice(1).map((g) => ({ key: g.key, creds: read(g) }));
        const budget = c.budget ? parseFloat(c.budget.value) || 0 : 0;
        return { id: c.id, enabled: c.enabled.checked, creds: read(c.groups[0]), extra, budget };
    });
}

// 渲染 Provider 卡片列表(勾选 + 多组凭证 + 余额预算)。
function renderProviderList(providers) {
    const container = document.getElementById("provider-list");
    container.innerHTML = "";
    providerCards = [];
    let uid = 0;

    providers.forEach((p) => {
        const card = document.createElement("div");
        card.className = "provider-card" + (p.enabled ? "" : " disabled");
        card.dataset.id = p.id;

        // 头部:勾选框 + 名称 + 动作按钮
        const head = document.createElement("div");
        head.className = "provider-head";

        const toggle = document.createElement("label");
        toggle.className = "provider-toggle";
        const cb = document.createElement("input");
        cb.type = "checkbox";
        cb.className = "provider-check";
        cb.checked = !!p.enabled;
        const name = document.createElement("span");
        name.className = "provider-name";
        name.textContent = p.name;
        toggle.append(cb, name);

        const actions = document.createElement("div");
        actions.className = "provider-actions";
        const testBtn = document.createElement("button");
        testBtn.className = "btn-sm";
        testBtn.dataset.test = p.id;
        testBtn.dataset.cred = "0";
        testBtn.textContent = "测试";
        actions.appendChild(testBtn);
        if (p.login_url) {
            const loginBtn = document.createElement("button");
            loginBtn.className = "btn-sm";
            loginBtn.dataset.open = p.login_url;
            loginBtn.textContent = "打开登录页";
            actions.appendChild(loginBtn);
        }

        head.append(toggle, actions);
        card.appendChild(head);

        // 凭证区:主组 + 附加组,每组一行输入,组间可增删
        const fieldsWrap = document.createElement("div");
        fieldsWrap.className = "provider-fields";
        card.appendChild(fieldsWrap); // 必须先挂到卡片上:后面 buildGroup 只是往里填

        const meta = { id: p.id, enabled: cb, groups: [], budget: null };

        // 建一组凭证的 DOM(返回 {key, fields})
        const buildGroup = (groupKey, creds, index, removable) => {
            const block = document.createElement("div");
            block.className = "cred-group";

            const ghead = document.createElement("div");
            ghead.className = "cred-group-head";
            const glabel = document.createElement("span");
            glabel.className = "cred-group-title";
            glabel.textContent = index === 0 ? "凭证" : `凭证 ${index + 1}`;
            ghead.appendChild(glabel);

            const gtest = document.createElement("button");
            gtest.className = "btn-sm";
            gtest.dataset.test = p.id;
            gtest.dataset.cred = String(index);
            gtest.textContent = "测试";
            ghead.appendChild(gtest);

            if (removable) {
                const del = document.createElement("button");
                del.className = "btn-sm danger";
                del.textContent = "删除";
                del.addEventListener("click", () => {
                    const idx = meta.groups.indexOf(group);
                    if (idx > 0) {
                        meta.groups.splice(idx, 1);
                        block.remove();
                        renumberGroups(meta);
                        updateObjectHint();
                    }
                });
                ghead.appendChild(del);
            }
            block.appendChild(ghead);

            const fields = [];
            (p.fields || []).forEach((f) => {
                const group = document.createElement("div");
                group.className = "form-group";
                const label = document.createElement("label");
                label.textContent = f.label;
                const input = document.createElement(f.type === "textarea" ? "textarea" : "input");
                if (f.type === "password") input.type = "password";
                if (f.type === "text") input.type = "text";
                if (f.type === "textarea") input.rows = 2;
                input.placeholder = (creds && creds[f.key]) || "";
                input.addEventListener("input", updateObjectHint);
                group.append(label, input);
                block.appendChild(group);
                fields.push({ key: f.key, input });
            });
            fieldsWrap.appendChild(block);
            return { key: groupKey, fields, block };
        };

        const primaryKey = p.id + "#" + (uid++);
        meta.groups.push(buildGroup(primaryKey, p.creds || {}, 0, false));
        (p.extra || []).forEach((g) => {
            meta.groups.push(buildGroup(g.key || (p.id + "#" + uid++), g.creds || {}, meta.groups.length, true));
        });

        // 余额型 Provider 增加预算输入框
        if (p.kind === "balance") {
            const budgetGroup = document.createElement("div");
            budgetGroup.className = "form-group";
            const budgetLabel = document.createElement("label");
            budgetLabel.textContent = "预算(用于进度条计算)";
            const budgetInput = document.createElement("input");
            budgetInput.type = "number";
            budgetInput.id = "budget-" + p.id;
            budgetInput.min = "0";
            budgetInput.step = "0.01";
            budgetInput.placeholder = "设为 0 则不计算进度条";
            if (p.budget > 0) budgetInput.value = p.budget;
            budgetGroup.append(budgetLabel, budgetInput);
            fieldsWrap.appendChild(budgetGroup);
            meta.budget = budgetInput;
        }

        // 「再加一组凭证」按钮:同平台多 Key / 多账号
        const addBtn = document.createElement("button");
        addBtn.className = "btn-sm add-cred";
        addBtn.textContent = "+ 再加一组凭证(多 Key / 多账号)";
        addBtn.addEventListener("click", () => {
            const idx = meta.groups.length;
            meta.groups.push(buildGroup(p.id + "#" + (uid++), {}, idx, true));
            renumberGroups(meta);
            updateObjectHint();
        });
        fieldsWrap.appendChild(addBtn);

        // 勾选限制:按监控对象数算上限(普通 9 / 扩展 16),最少 1 家
        cb.addEventListener("change", () => {
            const total = currentObjectCount();
            if (!cb.checked && total < 1) {
                cb.checked = true;
                toast("至少保留 1 个 Provider", "error");
                return;
            }
            card.classList.toggle("disabled", !cb.checked);
            updateObjectHint();
        });

        container.appendChild(card);
        providerCards.push(meta);
    });
}

// 附加组增删后重新编号(标题、测试按钮的凭证序号都要跟上)。
function renumberGroups(meta) {
    meta.groups.forEach((g, i) => {
        const title = g.block.querySelector(".cred-group-title");
        if (title) title.textContent = i === 0 ? "凭证" : `凭证 ${i + 1}`;
        const t = g.block.querySelector("[data-test]");
        if (t) t.dataset.cred = String(i);
    });
}

document.getElementById("btn-save-config").addEventListener("click", async () => {
    const providers = collectProviders();
    const interval = parseInt(document.getElementById("input-interval").value) || 15;
    currentExtendedMode = document.getElementById("input-extended").checked;
    try {
        await window.go.main.App.SaveConfig(providers, interval, currentExtendedMode);
        // 清空输入框(已保存;placeholder 会显示新的掩码)
        providerCards.forEach((c) => c.groups.forEach((g) => g.fields.forEach((f) => {
            f.input.value = "";
        })));
        toast("已保存", "success");
        await loadConfig(); // 重新拉取(placeholder 显示新掩码)
        await refreshQuota(); // 立即刷新展示
    } catch (e) {
        console.error("saveConfig error:", e);
        toast("保存配置失败: " + e, "error");
    }
});

// 扩展模式开关:上限变了,提示文案要跟着变
document.getElementById("input-extended").addEventListener("change", (e) => {
    currentExtendedMode = e.target.checked;
    updateObjectHint();
    updateBall(currentResults); // 已勾选的对象可能超过新上限,立刻重排球格
});

// 测试/打开登录页按钮(事件委托,兼容动态生成的按钮)
document.addEventListener("click", async (e) => {
    const testBtn = e.target.closest("[data-test]");
    if (testBtn) {
        const platform = testBtn.getAttribute("data-test");
        const credIndex = parseInt(testBtn.getAttribute("data-cred") || "0", 10) || 0;
        // 先保存当前输入(刷新间隔不动),再测试
        try {
            await window.go.main.App.SaveConfig(collectProviders(), 0, currentExtendedMode);
            const result = await window.go.main.App.TestConnection(platform, credIndex);
            toast(result, result.startsWith("成功") ? "success" : "error");
        } catch (err) {
            console.error("testConnection error:", err);
            toast("测试连接失败: " + err, "error");
        }
        return;
    }
    const openBtn = e.target.closest("[data-open]");
    if (openBtn) {
        const url = openBtn.getAttribute("data-open");
        try {
            window.go.main.App.OpenLoginPage(url);
        } catch (err) {
            console.error("openLoginPage error:", err);
            toast("打开登录页失败: " + err, "error");
        }
    }
});

// === Toast(替代 alert) ===
let toastTimer = null;
function toast(msg, type) {
    const el = document.getElementById("toast");
    el.textContent = msg;
    el.className = "toast show " + (type || "");
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => {
        el.classList.remove("show");
    }, 2500);
}

// === 球位置记忆(拖动结束时保存)===
let dragTimer = null;
document.getElementById("ball").addEventListener("mouseup", () => {
    clearTimeout(dragTimer);
    dragTimer = setTimeout(() => {
        // Wails 获取窗口位置
        window.runtime.WindowGetPosition().then((pos) => {
            window.go.main.App.SaveBallPosition(pos.x, pos.y);
        });
    }, 500);
});

// === 启动:加载初始数据 ===
window.go.main.App.Refresh();

// === 托盘事件 ===
window.runtime.EventsOn("tray:refresh", () => {
    refreshQuota();
});

window.runtime.EventsOn("ui:show-settings", () => {
    setView("settings");
    loadConfig();
});
