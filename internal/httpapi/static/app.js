const statusEl = document.querySelector("#status");
const loginView = document.querySelector("#login-view");
const appView = document.querySelector("#app-view");
const loginForm = document.querySelector("#login-form");
const loginErrorEl = document.querySelector("#login-error");
const loginUsernameEl = document.querySelector("#login-username");
const loginPasswordEl = document.querySelector("#login-password");
const logoutEl = document.querySelector("#logout");
const metricsEl = document.querySelector("#metrics");
const teamsEl = document.querySelector("#teams");
const usersEl = document.querySelector("#users");
const sourcesEl = document.querySelector("#sources");
const nodesEl = document.querySelector("#nodes");
const virtualNodesEl = document.querySelector("#virtual-nodes");
const policiesEl = document.querySelector("#policies");
const tokensEl = document.querySelector("#tokens");
const refreshEl = document.querySelector("#refresh");
const configCheckEl = document.querySelector("#config-check");
const configPublishEl = document.querySelector("#config-publish");
const configRollbackEl = document.querySelector("#config-rollback");
const configRestartEl = document.querySelector("#config-restart");
const configCheckResultEl = document.querySelector("#config-check-result");
const teamForm = document.querySelector("#team-form");
const userForm = document.querySelector("#user-form");
const sourceForm = document.querySelector("#source-form");
const nodeImportForm = document.querySelector("#node-import-form");
const virtualNodeForm = document.querySelector("#virtual-node-form");
const policyForm = document.querySelector("#policy-form");
const tokenForm = document.querySelector("#token-form");
const userTeamSelect = document.querySelector("#user-team");
const nodeSourceSelect = document.querySelector("#node-source");
const tokenUserSelect = document.querySelector("#token-user");
const tokenResultEl = document.querySelector("#token-result");

refreshEl.addEventListener("click", load);
configCheckEl.addEventListener("click", checkConfig);
configPublishEl.addEventListener("click", publishConfig);
configRollbackEl.addEventListener("click", rollbackConfig);
configRestartEl.addEventListener("click", restartSingBox);
logoutEl.addEventListener("click", logout);
loginForm.addEventListener("submit", login);
teamForm.addEventListener("submit", submitTeam);
userForm.addEventListener("submit", submitUser);
sourceForm.addEventListener("submit", submitSource);
nodeImportForm.addEventListener("submit", submitNodeImport);
virtualNodeForm.addEventListener("submit", submitVirtualNode);
policyForm.addEventListener("submit", submitPolicy);
tokenForm.addEventListener("submit", submitToken);
sourcesEl.addEventListener("click", handleSourceAction);
tokensEl.addEventListener("click", handleTokenAction);
bootstrap();

let appState = {
  teams: [],
  users: [],
  sources: [],
  virtualNodes: [],
};

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
  appState = { teams: [], users: [], sources: [], virtualNodes: [] };
  tokenResultEl.hidden = true;
  tokenResultEl.textContent = "";
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
    const [overview, teams, users, sources, nodes, virtualNodes, policies, tokens] = await Promise.all([
      getJSON("/api/overview"),
      getJSON("/api/teams"),
      getJSON("/api/users"),
      getJSON("/api/sources"),
      getJSON("/api/nodes"),
      getJSON("/api/virtual-nodes"),
      getJSON("/api/policies"),
      getJSON("/api/tokens"),
    ]);
    appState = { teams, users, sources, virtualNodes };
    renderMetrics(overview);
    renderSelectors();
    renderTable(teamsEl, teams, ["id", "name", "description", "status"]);
    renderTable(usersEl, users, ["id", "team_id", "name", "email", "status"]);
    renderSources(sources);
    renderTable(nodesEl, nodes, ["id", "source_name", "raw_name", "display_name", "protocol", "status"]);
    renderTable(virtualNodesEl, virtualNodes, ["id", "name", "listen_protocol", "listen_port", "status"]);
    renderTable(policiesEl, policies, ["id", "name", "scope_type", "scope_id", "allowed_virtual_nodes", "max_nodes", "status"]);
    renderTokens(tokens);
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

async function submitTeam(event) {
  event.preventDefault();
  const form = new FormData(teamForm);
  await postAndReload("/api/teams", {
    name: textField(form, "name"),
    description: textField(form, "description"),
  });
  teamForm.reset();
}

async function submitUser(event) {
  event.preventDefault();
  const form = new FormData(userForm);
  const teamID = numberField(form, "team_id");
  await postAndReload("/api/users", {
    team_id: teamID > 0 ? teamID : null,
    name: textField(form, "name"),
    email: textField(form, "email"),
  });
  userForm.reset();
}

async function submitSource(event) {
  event.preventDefault();
  const form = new FormData(sourceForm);
  await postAndReload("/api/sources", {
    name: textField(form, "name"),
    type: textField(form, "type") || "manual",
    url: textField(form, "url"),
    refresh_interval_minutes: numberField(form, "refresh_interval_minutes"),
  });
  sourceForm.reset();
}

async function submitNodeImport(event) {
  event.preventDefault();
  const form = new FormData(nodeImportForm);
  await postAndReload("/api/nodes/import", {
    source_id: numberField(form, "source_id"),
    content: textField(form, "content"),
  });
  nodeImportForm.reset();
}

async function submitVirtualNode(event) {
  event.preventDefault();
  const form = new FormData(virtualNodeForm);
  await postAndReload("/api/virtual-nodes", {
    name: textField(form, "name"),
    listen_protocol: "vless",
    listen_port: numberField(form, "listen_port"),
  });
  virtualNodeForm.reset();
}

async function submitPolicy(event) {
  event.preventDefault();
  const form = new FormData(policyForm);
  const scopeID = numberField(form, "scope_id");
  await postAndReload("/api/policies", {
    name: textField(form, "name"),
    scope_type: textField(form, "scope_type") || "team",
    scope_id: scopeID > 0 ? scopeID : null,
    allowed_virtual_nodes: textField(form, "allowed_virtual_nodes"),
    max_nodes: numberField(form, "max_nodes"),
  });
  policyForm.reset();
}

async function submitToken(event) {
  event.preventDefault();
  const form = new FormData(tokenForm);
  const quotaMiB = numberField(form, "quota_mib");
  const result = await postAndReload("/api/tokens", {
    user_id: numberField(form, "user_id"),
    name: textField(form, "name"),
    expire_days: numberField(form, "expire_days"),
    quota_bytes: quotaMiB * 1024 * 1024,
  });
  tokenResultEl.hidden = false;
  tokenResultEl.innerHTML = `
    <strong>订阅地址</strong>
    <code>${escapeHTML(result.subscription)}</code>
  `;
  tokenForm.reset();
}

async function handleSourceAction(event) {
  const button = event.target.closest("button[data-action='refresh-source']");
  if (!button) return;
  button.disabled = true;
  statusEl.textContent = "刷新来源中";
  try {
    await postJSON(`/api/sources/${button.dataset.sourceId}/refresh`, {});
    await load();
  } catch (error) {
    statusEl.textContent = "刷新失败";
    button.disabled = false;
  }
}

async function handleTokenAction(event) {
  const button = event.target.closest("button[data-token-action]");
  if (!button) return;
  button.disabled = true;
  const id = button.dataset.tokenId;
  const action = button.dataset.tokenAction;
  statusEl.textContent = "更新 Token 中";
  try {
    if (action === "extend") {
      await postJSON(`/api/tokens/${id}/extend`, { extend_days: 30 });
    } else if (action === "quota") {
      await postJSON(`/api/tokens/${id}/quota`, { quota_bytes: 1024 * 1024 * 1024 });
    } else if (action === "revoke") {
      await postJSON(`/api/tokens/${id}/revoke`, {});
    } else if (action === "restore") {
      await postJSON(`/api/tokens/${id}/restore`, {});
    }
    await load();
  } catch (error) {
    statusEl.textContent = "更新失败";
    button.disabled = false;
  }
}

async function checkConfig() {
  configCheckEl.disabled = true;
  statusEl.textContent = "检查配置中";
  try {
    const result = await postJSON("/api/sing-box/config/check", {});
    configCheckResultEl.hidden = false;
    configCheckResultEl.innerHTML = `
      <strong>${result.valid ? "检查通过" : "检查失败"}</strong>
      <code>hash=${escapeHTML(String(result.config_hash || "").slice(0, 12))} in=${formatCell(result.inbound_count)} out=${formatCell(result.outbound_count)} upstream=${formatCell(result.upstream_outbound_count)} users=${formatCell(result.user_count)}</code>
    `;
    statusEl.textContent = result.valid ? "配置可用" : "配置异常";
  } catch (error) {
    configCheckResultEl.hidden = false;
    configCheckResultEl.innerHTML = `<strong>检查失败</strong><code>${escapeHTML(error.message)}</code>`;
    statusEl.textContent = "配置异常";
  } finally {
    configCheckEl.disabled = false;
  }
}

async function publishConfig() {
  configPublishEl.disabled = true;
  statusEl.textContent = "发布配置中";
  try {
    const result = await postJSON("/api/sing-box/config/publish", {});
    configCheckResultEl.hidden = false;
    const restartText = result.restart ? ` restart=${formatRestartResult(result.restart)}` : ` restart=${result.restart_required ? "需要" : "无需"}`;
    configCheckResultEl.innerHTML = `
      <strong>${result.published ? "发布完成" : "发布失败"}</strong>
      <code>hash=${escapeHTML(String(result.config_hash || "").slice(0, 12))} previous=${result.previous_saved ? "已保存" : "无"}${restartText} out=${formatCell(result.outbound_count)} users=${formatCell(result.user_count)}</code>
    `;
    statusEl.textContent = result.published && !result.restart_required ? "配置已发布并生效" : result.published ? "配置已发布，需重启 sing-box" : "发布失败";
  } catch (error) {
    configCheckResultEl.hidden = false;
    configCheckResultEl.innerHTML = `<strong>发布失败</strong><code>${escapeHTML(error.message)}</code>`;
    statusEl.textContent = "发布失败";
  } finally {
    configPublishEl.disabled = false;
  }
}

async function rollbackConfig() {
  configRollbackEl.disabled = true;
  statusEl.textContent = "回滚配置中";
  try {
    const result = await postJSON("/api/sing-box/config/rollback", {});
    configCheckResultEl.hidden = false;
    const restartText = result.restart ? ` restart=${formatRestartResult(result.restart)}` : ` restart=${result.restart_required ? "需要" : "无需"}`;
    configCheckResultEl.innerHTML = `
      <strong>${result.rolled_back ? "回滚完成" : "回滚失败"}</strong>
      <code>hash=${escapeHTML(String(result.config_hash || "").slice(0, 12))}${restartText} out=${formatCell(result.outbound_count)} users=${formatCell(result.user_count)}</code>
    `;
    statusEl.textContent = result.rolled_back && !result.restart_required ? "配置已回滚并生效" : result.rolled_back ? "配置已回滚，需重启 sing-box" : "回滚失败";
  } catch (error) {
    configCheckResultEl.hidden = false;
    configCheckResultEl.innerHTML = `<strong>回滚失败</strong><code>${escapeHTML(error.message)}</code>`;
    statusEl.textContent = "回滚失败";
  } finally {
    configRollbackEl.disabled = false;
  }
}

async function restartSingBox() {
  configRestartEl.disabled = true;
  statusEl.textContent = "重启服务中";
  try {
    const result = await postJSON("/api/sing-box/restart", {});
    configCheckResultEl.hidden = false;
    configCheckResultEl.innerHTML = `
      <strong>${result.success ? "重启已执行" : result.skipped ? "重启未启用" : "重启失败"}</strong>
      <code>enabled=${result.enabled ? "true" : "false"} executed=${result.executed ? "true" : "false"} duration_ms=${formatCell(result.duration_ms)} message=${escapeHTML(result.message || "")}</code>
    `;
    statusEl.textContent = result.success ? "服务已重启" : result.skipped ? "重启未启用" : "重启失败";
  } catch (error) {
    configCheckResultEl.hidden = false;
    configCheckResultEl.innerHTML = `<strong>重启失败</strong><code>${escapeHTML(error.message)}</code>`;
    statusEl.textContent = "重启失败";
  } finally {
    configRestartEl.disabled = false;
  }
}

function formatRestartResult(result) {
  if (result.success) {
    return "完成";
  }
  if (result.skipped) {
    return "未启用";
  }
  return "失败";
}

async function postAndReload(path, payload) {
  statusEl.textContent = "保存中";
  try {
    const result = await postJSON(path, payload);
    await load();
    return result;
  } catch (error) {
    statusEl.textContent = "保存失败";
    throw error;
  }
}

function renderSelectors() {
  renderOptions(userTeamSelect, appState.teams, "不绑定团队");
  renderOptions(nodeSourceSelect, appState.sources, "选择来源");
  renderOptions(tokenUserSelect, appState.users, "选择成员");
}

function renderOptions(target, rows, emptyLabel) {
  const options = [`<option value="">${escapeHTML(emptyLabel)}</option>`].concat(
    (rows || []).map((row) => `<option value="${row.id}">${escapeHTML(row.name)}</option>`),
  );
  target.innerHTML = options.join("");
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

function textField(form, name) {
  return String(form.get(name) || "").trim();
}

function numberField(form, name) {
  const value = Number.parseInt(String(form.get(name) || "0"), 10);
  return Number.isFinite(value) ? value : 0;
}

function renderMetrics(data) {
  const items = [
    ["团队", data.teams],
    ["用户", data.users],
    ["Token", data.tokens],
    ["来源", data.sources],
    ["节点", data.nodes],
    ["虚拟节点", data.virtual_nodes],
    ["策略", data.policies],
  ];
  metricsEl.innerHTML = items
    .map(([label, value]) => `<div class="metric"><span>${label}</span><strong>${value ?? 0}</strong></div>`)
    .join("");
}

function renderSources(rows) {
  if (!rows || rows.length === 0) {
    sourcesEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  sourcesEl.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>id</th>
          <th>name</th>
          <th>type</th>
          <th>display_prefix</th>
          <th>refresh_min</th>
          <th>last_sync_at</th>
          <th>last_error</th>
          <th>actions</th>
        </tr>
      </thead>
      <tbody>
        ${rows
          .map(
            (row) => `
              <tr>
                <td>${formatCell(row.id)}</td>
                <td>${formatCell(row.name)}</td>
                <td>${formatCell(row.type)}</td>
                <td>${formatCell(row.display_prefix)}</td>
                <td>${formatCell(row.refresh_interval_minutes)}</td>
                <td>${formatCell(row.last_sync_at)}</td>
                <td>${formatCell(row.last_error)}</td>
                <td>
                  <button class="table-button" data-action="refresh-source" data-source-id="${row.id}">刷新</button>
                </td>
              </tr>
            `,
          )
          .join("")}
      </tbody>
    </table>
  `;
}

function renderTokens(rows) {
  if (!rows || rows.length === 0) {
    tokensEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  tokensEl.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>id</th>
          <th>user_id</th>
          <th>token_prefix</th>
          <th>name</th>
          <th>status</th>
          <th>expire_at</th>
          <th>quota_bytes</th>
          <th>actions</th>
        </tr>
      </thead>
      <tbody>
        ${rows
          .map(
            (row) => `
              <tr>
                <td>${formatCell(row.id)}</td>
                <td>${formatCell(row.user_id)}</td>
                <td>${formatCell(row.token_prefix)}</td>
                <td>${formatCell(row.name)}</td>
                <td>${formatCell(row.status)}</td>
                <td>${formatCell(row.expire_at)}</td>
                <td>${formatCell(row.quota_bytes)}</td>
                <td class="table-actions">
                  <button class="table-button" data-token-action="extend" data-token-id="${row.id}">续期30天</button>
                  <button class="table-button" data-token-action="quota" data-token-id="${row.id}">+1024MiB</button>
                  <button class="table-button" data-token-action="restore" data-token-id="${row.id}">恢复</button>
                  <button class="table-button danger-button" data-token-action="revoke" data-token-id="${row.id}">撤销</button>
                </td>
              </tr>
            `,
          )
          .join("")}
      </tbody>
    </table>
  `;
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
