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
        const percent = r.percent || 0;
        const item = document.createElement("div");
        item.className = "quota-item";
        item.innerHTML = `
            <div class="quota-item-header">
                <span class="quota-platform">${r.platform}</span>
                <span class="quota-remaining">${r.error ? "⚠ " + r.error : r.remaining || "-"}</span>
            </div>
            <div class="progress-bar">
                <div class="progress-fill ${color}" style="width: ${r.error ? 100 : percent}%"></div>
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
    if (r.error) return "red";
    if (r.percent >= 90) return "red";
    if (r.percent >= 75) return "yellow";
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
    try {
        const cfg = await window.go.main.App.GetConfig();
        document.getElementById("input-kimi").placeholder = cfg.kimi_api_key || "sk-kimi-xxx";
        document.getElementById("input-xfyun").placeholder = cfg.xfyun_cookie || "从浏览器 F12 复制 Cookie";
        document.getElementById("input-mimo").placeholder = cfg.mimo_cookie || "从浏览器 F12 复制 Cookie";
        document.getElementById("input-interval").value = cfg.refresh_interval_min || 15;
    } catch (e) {
        console.error("loadConfig error:", e);
        alert("加载配置失败: " + e);
    }
}

document.getElementById("btn-save-config").addEventListener("click", async () => {
    const kimi = document.getElementById("input-kimi").value;
    const xfyun = document.getElementById("input-xfyun").value;
    const mimo = document.getElementById("input-mimo").value;
    const interval = parseInt(document.getElementById("input-interval").value) || 15;
    try {
        await window.go.main.App.SaveConfig(kimi, xfyun, mimo, interval);
        // 清空输入框(已保存)
        document.getElementById("input-kimi").value = "";
        document.getElementById("input-xfyun").value = "";
        document.getElementById("input-mimo").value = "";
        alert("已保存");
    } catch (e) {
        console.error("saveConfig error:", e);
        alert("保存配置失败: " + e);
    }
});

// 测试连接按钮
document.querySelectorAll("[data-test]").forEach((btn) => {
    btn.addEventListener("click", async () => {
        const platform = btn.getAttribute("data-test");
        // 先保存当前输入(如果有)
        const kimi = document.getElementById("input-kimi").value;
        const xfyun = document.getElementById("input-xfyun").value;
        const mimo = document.getElementById("input-mimo").value;
        try {
            if (kimi || xfyun || mimo) {
                await window.go.main.App.SaveConfig(kimi, xfyun, mimo, 0);
            }
            const result = await window.go.main.App.TestConnection(platform);
            alert(result);
        } catch (e) {
            console.error("testConnection error:", e);
            alert("测试连接失败: " + e);
        }
    });
});

// 打开登录页按钮
document.querySelectorAll("[data-open]").forEach((btn) => {
    btn.addEventListener("click", () => {
        const url = btn.getAttribute("data-open");
        try {
            window.go.main.App.OpenLoginPage(url);
        } catch (e) {
            console.error("openLoginPage error:", e);
            alert("打开登录页失败: " + e);
        }
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

// === 托盘事件 ===
window.runtime.EventsOn("tray:refresh", () => {
    refreshQuota();
});

window.runtime.EventsOn("ui:show-settings", () => {
    document.getElementById("ball").classList.add("hidden");
    document.getElementById("panel").classList.add("hidden");
    document.getElementById("settings").classList.remove("hidden");
    loadConfig();
});
