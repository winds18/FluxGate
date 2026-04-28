const statusEl = document.querySelector("#status");
const loginView = document.querySelector("#login-view");
const appView = document.querySelector("#app-view");
const loginForm = document.querySelector("#login-form");
const loginErrorEl = document.querySelector("#login-error");
const loginUsernameEl = document.querySelector("#login-username");
const loginPasswordEl = document.querySelector("#login-password");
const logoutEl = document.querySelector("#logout");
const metricsEl = document.querySelector("#metrics");
const sourcesEl = document.querySelector("#sources");
const nodesEl = document.querySelector("#nodes");
const tokensEl = document.querySelector("#tokens");
const refreshEl = document.querySelector("#refresh");

refreshEl.addEventListener("click", load);
logoutEl.addEventListener("click", logout);
loginForm.addEventListener("submit", login);
bootstrap();

async function bootstrap() {
  statusEl.textContent = "连接中";
  try {
    await getJSON("/api/auth/session");
    showApp();
    await load();
  } catch (error) {
    showLogin();
  }
}

async function login(event) {
  event.preventDefault();
  loginErrorEl.textContent = "";
  statusEl.textContent = "登录中";
  try {
    const username = loginUsernameEl.value.trim();
    const password = loginPasswordEl.value;
    await postJSON("/api/auth/login", { username, password });
    loginPasswordEl.value = "";
    showApp();
    await load();
  } catch (error) {
    statusEl.textContent = "未登录";
    loginErrorEl.textContent = "账号或密码不正确";
  }
}

async function logout() {
  await fetch("/api/auth/logout", { method: "POST" });
  showLogin();
}

function showLogin() {
  appView.hidden = true;
  loginView.hidden = false;
  logoutEl.hidden = true;
  statusEl.textContent = "未登录";
  loginUsernameEl.focus();
}

function showApp() {
  loginView.hidden = true;
  appView.hidden = false;
  logoutEl.hidden = false;
  statusEl.textContent = "已登录";
}

async function load() {
  statusEl.textContent = "刷新中";
  try {
    const [overview, sources, nodes, tokens] = await Promise.all([
      getJSON("/api/overview"),
      getJSON("/api/sources"),
      getJSON("/api/nodes"),
      getJSON("/api/tokens"),
    ]);
    renderMetrics(overview);
    renderTable(sourcesEl, sources, ["id", "name", "type", "display_prefix", "status"]);
    renderTable(nodesEl, nodes, ["id", "source_name", "raw_name", "display_name", "protocol", "status"]);
    renderTable(tokensEl, tokens, ["id", "user_id", "token_prefix", "name", "status", "quota_bytes"]);
    statusEl.textContent = "已连接";
  } catch (error) {
    if (error.status === 401) {
      showLogin();
      return;
    }
    statusEl.textContent = "异常";
    metricsEl.innerHTML = `<div class="empty">${escapeHTML(error.message)}</div>`;
  }
}

async function getJSON(path) {
  const response = await fetch(path, { credentials: "same-origin" });
  if (!response.ok) {
    const error = new Error(`${path} ${response.status}`);
    error.status = response.status;
    throw error;
  }
  return response.json();
}

async function postJSON(path, payload) {
  const response = await fetch(path, {
    method: "POST",
    credentials: "same-origin",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    const error = new Error(`${path} ${response.status}`);
    error.status = response.status;
    throw error;
  }
  return response.json();
}

function renderMetrics(data) {
  const items = [
    ["团队", data.teams],
    ["用户", data.users],
    ["Token", data.tokens],
    ["来源", data.sources],
    ["节点", data.nodes],
    ["虚拟节点", data.virtual_nodes],
  ];
  metricsEl.innerHTML = items
    .map(([label, value]) => `<div class="metric"><span>${label}</span><strong>${value ?? 0}</strong></div>`)
    .join("");
}

function renderTable(target, rows, columns) {
  if (!rows || rows.length === 0) {
    target.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  target.innerHTML = `
    <table>
      <thead>
        <tr>${columns.map((column) => `<th>${escapeHTML(column)}</th>`).join("")}</tr>
      </thead>
      <tbody>
        ${rows
          .map(
            (row) =>
              `<tr>${columns
                .map((column) => `<td>${formatCell(row[column])}</td>`)
                .join("")}</tr>`,
          )
          .join("")}
      </tbody>
    </table>
  `;
}

function formatCell(value) {
  if (value === null || value === undefined || value === "") return "";
  if (typeof value === "string" && value.length > 38) {
    return `<code>${escapeHTML(value.slice(0, 38))}...</code>`;
  }
  return escapeHTML(String(value));
}

function escapeHTML(value) {
  return value.replace(/[&<>"']/g, (char) => {
    switch (char) {
      case "&":
        return "&amp;";
      case "<":
        return "&lt;";
      case ">":
        return "&gt;";
      case '"':
        return "&quot;";
      default:
        return "&#039;";
    }
  });
}
