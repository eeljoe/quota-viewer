// === Wails 绑定 ===
// window.go.main.App 在运行时由 Wails 注入

let isExpanded = false;
let currentResults = [];

// === 事件监听 ===
window.runtime.EventsOn("quota:update", (results) => {
    currentResults = results;
    renderResults(results);
});

// === 球面点击:展开/收起 ===
document.getElementById("ball").addEventListener("click", () => {
    togglePanel();
});

function togglePanel() {
    isExpanded = !isExpanded;
    const ball = document.getElementById("ball");
    const panel = document.getElementById("panel");
    if (isExpanded) {
        ball.classList.add("hidden");
        panel.classList.remove("hidden");
        // 展开时若数据超过 3 分钟则刷新
        refreshIfNeeded();
    } else {
        panel.classList.add("hidden");
        ball.classList.remove("hidden");
    }
}

// === 刷新 ===
document.getElementById("btn-refresh").addEventListener("click", () => {
    refreshQuota();
});

async function refreshQuota() {
    document.getElementById("last-updated").textContent = "刷新中...";
    try {
        const results = await window.go.main.App.Refresh();
        currentResults = results;
        renderResults(results);
    } catch (e) {
        console.error("refresh error:", e);
    }
}

let lastRefreshTime = 0;
async function refreshIfNeeded() {
    if (Date.now() - lastRefreshTime > 3 * 60 * 1000) {
        await refreshQuota();
    }
}

// === 渲染结果 ===
function renderResults(results) {
    // 更新详情面板
    const list = document.getElementById("quota-list");
    list.innerHTML = "";
    results.forEach((r) => {
        const color = getDotColor(r);
        const percent = r.Percent || 0;
        const item = document.createElement("div");
        item.className = "quota-item";
        item.innerHTML = `
            <div class="quota-item-header">
                <span class="quota-platform">${r.Platform}</span>
                <span class="quota-remaining">${r.Error ? "⚠ " + r.Error : r.Remaining || "-"}</span>
            </div>
            <div class="progress-bar">
                <div class="progress-fill ${color}" style="width: ${r.Error ? 100 : percent}%"></div>
            </div>
        `;
        list.appendChild(item);
    });

    // 更新球面指示点
    updateDots(results);

    // 更新时间
    const now = new Date();
    lastRefreshTime = now.getTime();
    document.getElementById("last-updated").textContent = "更新于 " + now.toLocaleTimeString("zh-CN");
}

function getDotColor(r) {
    if (r.Error) return "red";
    if (r.Percent >= 90) return "red";
    if (r.Percent >= 75) return "yellow";
    return "green";
}

function updateDots(results) {
    const names = ["Kimi", "讯飞星辰", "小米MiMo"];
    const ids = ["dot-kimi", "dot-xfyun", "dot-mimo"];
    results.forEach((r, i) => {
        const dot = document.getElementById(ids[i]);
        if (!dot) return;
        dot.className = "dot " + getDotColor(r);
    });
}

// === 配置面板 ===
document.getElementById("btn-settings").addEventListener("click", () => {
    document.getElementById("panel").classList.add("hidden");
    document.getElementById("settings").classList.remove("hidden");
    loadConfig();
});

document.getElementById("btn-close-settings").addEventListener("click", () => {
    document.getElementById("settings").classList.add("hidden");
    document.getElementById("ball").classList.remove("hidden");
    isExpanded = false;
});

async function loadConfig() {
    const cfg = await window.go.main.App.GetConfig();
    document.getElementById("input-kimi").placeholder = cfg.kimi_api_key || "sk-kimi-xxx";
    document.getElementById("input-xfyun").placeholder = cfg.xfyun_cookie || "从浏览器 F12 复制 Cookie";
    document.getElementById("input-mimo").placeholder = cfg.mimo_cookie || "从浏览器 F12 复制 Cookie";
    document.getElementById("input-interval").value = cfg.refresh_interval_min || 15;
}

document.getElementById("btn-save-config").addEventListener("click", async () => {
    const kimi = document.getElementById("input-kimi").value;
    const xfyun = document.getElementById("input-xfyun").value;
    const mimo = document.getElementById("input-mimo").value;
    const interval = parseInt(document.getElementById("input-interval").value) || 15;
    await window.go.main.App.SaveConfig(kimi, xfyun, mimo, interval);
    // 清空输入框(已保存)
    document.getElementById("input-kimi").value = "";
    document.getElementById("input-xfyun").value = "";
    document.getElementById("input-mimo").value = "";
    alert("已保存");
});

// 测试连接按钮
document.querySelectorAll("[data-test]").forEach((btn) => {
    btn.addEventListener("click", async () => {
        const platform = btn.getAttribute("data-test");
        // 先保存当前输入(如果有)
        const kimi = document.getElementById("input-kimi").value;
        const xfyun = document.getElementById("input-xfyun").value;
        const mimo = document.getElementById("input-mimo").value;
        if (kimi || xfyun || mimo) {
            await window.go.main.App.SaveConfig(kimi, xfyun, mimo, 0);
        }
        const result = await window.go.main.App.TestConnection(platform);
        alert(result);
    });
});

// 打开登录页按钮
document.querySelectorAll("[data-open]").forEach((btn) => {
    btn.addEventListener("click", () => {
        const url = btn.getAttribute("data-open");
        window.go.main.App.OpenLoginPage(url);
    });
});

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
