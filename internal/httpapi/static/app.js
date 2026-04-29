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
const trafficHourlyEl = document.querySelector("#traffic-hourly");
const trafficDailyEl = document.querySelector("#traffic-daily");
const trafficOutboundsEl = document.querySelector("#traffic-outbounds");
const trafficTokensEl = document.querySelector("#traffic-tokens");
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
nodesEl.addEventListener("click", handleNodeAction);
nodesEl.addEventListener("submit", handleNodeEditSubmit);
tokensEl.addEventListener("click", handleTokenAction);
bootstrap();

let appState = {
  teams: [],
  users: [],
  sources: [],
  nodes: [],
  virtualNodes: [],
  editingSourceID: null,
  editingNodeID: null,
};

const columnLabels = {
  id: "ID",
  name: "名称",
  description: "备注",
  status: "状态",
  team_id: "团队 ID",
  user_id: "成员 ID",
  token_id: "Token ID",
  gateway_account_id: "网关账号 ID",
  source_name: "来源",
  url: "URL",
  raw_name: "原始名称",
  display_name: "展示名称",
  name_mode: "命名模式",
  protocol: "协议",
  tags: "标签",
  listen_protocol: "监听协议",
  listen_port: "端口",
  tag_selector: "标签选择器",
  scope_type: "范围",
  scope_id: "范围 ID",
  include_tags: "包含标签",
  exclude_tags: "排除标签",
  allowed_virtual_nodes: "允许虚拟节点",
  max_nodes: "最大节点数",
  token_prefix: "Token 前缀",
  expire_at: "到期时间",
  quota_bytes: "额度",
  used_total: "已用",
  auth_user: "网关用户",
  token_status: "Token 状态",
  today_total_bytes: "今日",
  month_total_bytes: "本月",
  used_upload_bytes: "累计上传",
  used_download_bytes: "累计下载",
  used_total_bytes: "累计总量",
  outbound_tag: "出口标签",
  upstream_node_id: "节点 ID",
  node_name: "节点",
  upload_bytes: "上传",
  download_bytes: "下载",
  total_bytes: "总量",
  updated_at: "更新时间",
  type: "类型",
  display_prefix: "前缀",
  default_tags: "默认标签",
  refresh_interval_minutes: "刷新分钟",
  last_sync_at: "上次同步",
  last_error: "错误",
  actions: "操作",
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
  appState = { teams: [], users: [], sources: [], nodes: [], virtualNodes: [], editingSourceID: null, editingNodeID: null };
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
    const [overview, teams, users, sources, nodes, virtualNodes, policies, tokens, trafficHourly, trafficDaily, trafficOutbounds, trafficTokens] = await Promise.all([
      getJSON("/api/overview"),
      getJSON("/api/teams"),
      getJSON("/api/users"),
      getJSON("/api/sources"),
      getJSON("/api/nodes"),
      getJSON("/api/virtual-nodes"),
      getJSON("/api/policies"),
      getJSON("/api/tokens"),
      getJSON("/api/traffic/hourly?hours=24"),
      getJSON("/api/traffic/daily?days=14"),
      getJSON("/api/traffic/outbounds?days=14"),
      getJSON("/api/traffic/tokens"),
    ]);
    appState = { ...appState, teams, users, sources, nodes, virtualNodes };
    renderMetrics(overview);
    renderSelectors();
    renderTable(teamsEl, teams, ["id", "name", "description", "status"]);
    renderTable(usersEl, users, ["id", "team_id", "name", "email", "status"]);
    renderSources(sources);
    renderNodes(nodes);
    renderTable(virtualNodesEl, virtualNodes, ["id", "name", "listen_protocol", "listen_port", "tag_selector", "status"]);
    renderTable(policiesEl, policies, ["id", "name", "scope_type", "scope_id", "include_tags", "exclude_tags", "allowed_virtual_nodes", "max_nodes", "status"]);
    renderTokens(tokens);
    renderTrafficHourly(trafficHourly);
    renderTrafficDaily(trafficDaily);
    renderTrafficOutbounds(trafficOutbounds);
    renderTrafficTokens(trafficTokens);
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
    default_tags: textField(form, "default_tags"),
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
    tag_selector: textField(form, "tag_selector"),
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
    include_tags: textField(form, "include_tags"),
    exclude_tags: textField(form, "exclude_tags"),
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
  const button = event.target.closest("button[data-source-action], button[data-action='refresh-source']");
  if (!button) return;
  const id = button.dataset.sourceId;
  const action = button.dataset.sourceAction || "refresh";
  if (action === "edit") {
    appState.editingSourceID = Number.parseInt(id || "0", 10);
    renderSources(appState.sources);
    sourcesEl.querySelector(`tr[data-source-id="${id}"] input[data-source-field="name"]`)?.focus();
    return;
  }
  if (action === "cancel") {
    appState.editingSourceID = null;
    renderSources(appState.sources);
    return;
  }
  if (action === "save") {
    await saveSource(button);
    return;
  }
  if (action === "regenerate") {
    await regenerateSourceNames(button);
    return;
  }
  if (action !== "refresh") return;
  button.disabled = true;
  statusEl.textContent = "刷新来源中";
  try {
    await postJSON(`/api/sources/${id}/refresh`, {});
    await load();
  } catch (error) {
    statusEl.textContent = "刷新失败";
    button.disabled = false;
  }
}

async function saveSource(button) {
  const row = button.closest("tr[data-source-id]");
  if (!row) return;
  const id = row.dataset.sourceId;
  const refreshInterval = Number.parseInt(row.querySelector('[data-source-field="refresh_interval_minutes"]')?.value || "0", 10);
  const payload = {
    name: sourceFieldValue(row, "name"),
    type: sourceFieldValue(row, "type"),
    url: sourceFieldValue(row, "url"),
    display_prefix: sourceFieldValue(row, "display_prefix"),
    default_tags: sourceFieldValue(row, "default_tags"),
    refresh_interval_minutes: Number.isFinite(refreshInterval) ? refreshInterval : 0,
  };
  row.querySelectorAll("button").forEach((item) => {
    item.disabled = true;
  });
  statusEl.textContent = "保存来源中";
  try {
    await patchJSON(`/api/sources/${id}`, payload);
    appState.editingSourceID = null;
    await load();
  } catch (error) {
    statusEl.textContent = "保存失败";
    row.querySelectorAll("button").forEach((item) => {
      item.disabled = false;
    });
  }
}

async function regenerateSourceNames(button) {
  const id = button.dataset.sourceId;
  button.disabled = true;
  statusEl.textContent = "同步节点命名中";
  try {
    await postJSON(`/api/sources/${id}/regenerate-node-names`, {});
    await load();
  } catch (error) {
    statusEl.textContent = "同步失败";
    button.disabled = false;
  }
}

function sourceFieldValue(row, name) {
  return String(row.querySelector(`[data-source-field="${name}"]`)?.value || "").trim();
}

async function handleNodeAction(event) {
  const button = event.target.closest("button[data-node-action]");
  if (!button) return;
  const id = Number.parseInt(button.dataset.nodeId || "0", 10);
  if (!id) return;
  const action = button.dataset.nodeAction;
  if (action === "edit") {
    appState.editingNodeID = id;
    renderNodes(appState.nodes);
    nodesEl.querySelector(`form[data-node-id="${id}"] input[name="display_name"]`)?.focus();
    return;
  }
  if (action === "cancel") {
    appState.editingNodeID = null;
    renderNodes(appState.nodes);
    return;
  }
  if (action !== "reset-name") return;
  button.disabled = true;
  statusEl.textContent = "恢复节点命名中";
  try {
    await postJSON(`/api/nodes/${id}/reset-display-name`, {});
    appState.editingNodeID = null;
    await load();
  } catch (error) {
    statusEl.textContent = "恢复失败";
    button.disabled = false;
  }
}

async function handleNodeEditSubmit(event) {
  const form = event.target.closest("form[data-node-edit-form]");
  if (!form) return;
  event.preventDefault();
  const id = Number.parseInt(form.dataset.nodeId || "0", 10);
  if (!id) return;
  const displayName = textField(new FormData(form), "display_name");
  form.querySelectorAll("button").forEach((button) => {
    button.disabled = true;
  });
  statusEl.textContent = "保存节点中";
  try {
    await patchJSON(`/api/nodes/${id}`, { display_name: displayName });
    appState.editingNodeID = null;
    await load();
  } catch (error) {
    statusEl.textContent = "保存失败";
    form.querySelectorAll("button").forEach((button) => {
      button.disabled = false;
    });
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
  return sendJSON("POST", path, payload);
}

async function patchJSON(path, payload) {
  return sendJSON("PATCH", path, payload);
}

async function sendJSON(method, path, payload) {
  const response = await fetch(path, {
    method,
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
  appState.sources = rows || [];
  if (!rows || rows.length === 0) {
    sourcesEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  sourcesEl.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>${labelForColumn("id")}</th>
          <th>${labelForColumn("name")}</th>
          <th>${labelForColumn("type")}</th>
          <th>${labelForColumn("url")}</th>
          <th>${labelForColumn("display_prefix")}</th>
          <th>${labelForColumn("default_tags")}</th>
          <th>${labelForColumn("refresh_interval_minutes")}</th>
          <th>${labelForColumn("last_sync_at")}</th>
          <th>${labelForColumn("last_error")}</th>
          <th>${labelForColumn("actions")}</th>
        </tr>
      </thead>
      <tbody>
        ${rows.map((row) => renderSourceRow(row)).join("")}
      </tbody>
    </table>
  `;
}

function renderSourceRow(row) {
  const isEditing = appState.editingSourceID === row.id;
  return `
    <tr data-source-id="${row.id}" class="${isEditing ? "source-edit-row" : ""}">
      <td>${formatCell(row.id, "id")}</td>
      <td>${isEditing ? sourceTextInput(row, "name") : formatCell(row.name, "name")}</td>
      <td>${isEditing ? sourceTypeSelect(row.type) : formatCell(row.type, "type")}</td>
      <td>${isEditing ? sourceTextInput(row, "url", "table-edit-input-wide") : formatCell(row.url, "url")}</td>
      <td>${isEditing ? sourceTextInput(row, "display_prefix", "table-edit-input", "留空自动") : formatCell(row.display_prefix, "display_prefix")}</td>
      <td>${isEditing ? sourceTextInput(row, "default_tags", "table-edit-input", "HK, Premium") : formatCell(row.default_tags, "default_tags")}</td>
      <td>${isEditing ? sourceNumberInput(row, "refresh_interval_minutes") : formatCell(row.refresh_interval_minutes, "refresh_interval_minutes")}</td>
      <td>${formatCell(row.last_sync_at, "last_sync_at")}</td>
      <td>${formatCell(row.last_error, "last_error")}</td>
      <td class="table-actions source-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-source-action="save" data-source-id="${row.id}">保存</button>
               <button class="table-button ghost-button" type="button" data-source-action="cancel" data-source-id="${row.id}">取消</button>`
            : `<button class="table-button" type="button" data-source-action="edit" data-source-id="${row.id}">编辑</button>`
        }
        <button class="table-button" type="button" data-source-action="refresh" data-source-id="${row.id}">刷新</button>
        <button class="table-button ghost-button" type="button" data-source-action="regenerate" data-source-id="${row.id}">同步命名</button>
      </td>
    </tr>
  `;
}

function sourceTextInput(row, field, className = "table-edit-input", placeholder = "") {
  return `<input class="${className}" data-source-field="${field}" value="${escapeHTML(String(row[field] || ""))}" placeholder="${escapeHTML(placeholder)}" />`;
}

function sourceNumberInput(row, field) {
  return `<input class="table-edit-input table-edit-number" data-source-field="${field}" type="number" min="0" step="1" value="${escapeHTML(String(row[field] || 0))}" />`;
}

function sourceTypeSelect(value) {
  const current = String(value || "manual");
  return `
    <select class="table-edit-input" data-source-field="type">
      <option value="manual" ${current === "manual" ? "selected" : ""}>manual</option>
      <option value="subscription" ${current === "subscription" ? "selected" : ""}>subscription</option>
    </select>
  `;
}

function renderNodes(rows) {
  appState.nodes = rows || [];
  if (!rows || rows.length === 0) {
    nodesEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  nodesEl.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>${labelForColumn("id")}</th>
          <th>${labelForColumn("source_name")}</th>
          <th>${labelForColumn("raw_name")}</th>
          <th>${labelForColumn("display_name")}</th>
          <th>${labelForColumn("name_mode")}</th>
          <th>${labelForColumn("protocol")}</th>
          <th>${labelForColumn("tags")}</th>
          <th>${labelForColumn("status")}</th>
          <th>${labelForColumn("actions")}</th>
        </tr>
      </thead>
      <tbody>
        ${rows.map((row) => renderNodeRow(row)).join("")}
      </tbody>
    </table>
  `;
}

function renderNodeRow(row) {
  const isEditing = appState.editingNodeID === row.id;
  return `
    <tr>
      <td>${formatCell(row.id, "id")}</td>
      <td>${formatCell(row.source_name, "source_name")}</td>
      <td>${formatCell(row.raw_name, "raw_name")}</td>
      <td>${isEditing ? renderNodeEditForm(row) : formatCell(row.display_name, "display_name")}</td>
      <td>${formatCell(row.name_mode, "name_mode")}</td>
      <td>${formatCell(row.protocol, "protocol")}</td>
      <td>${formatCell(row.tags, "tags")}</td>
      <td>${formatCell(row.status, "status")}</td>
      <td class="table-actions node-actions">
        ${
          isEditing
            ? `<button class="table-button ghost-button" type="button" data-node-action="cancel" data-node-id="${row.id}">取消</button>`
            : `<button class="table-button" type="button" data-node-action="edit" data-node-id="${row.id}">编辑</button>`
        }
        <button class="table-button ghost-button" type="button" data-node-action="reset-name" data-node-id="${row.id}" ${row.name_mode === "auto" ? "disabled" : ""}>恢复自动</button>
      </td>
    </tr>
  `;
}

function renderNodeEditForm(row) {
  return `
    <form class="inline-edit-form" data-node-edit-form data-node-id="${row.id}">
      <input name="display_name" value="${escapeHTML(row.display_name || "")}" required />
      <button class="table-button" type="submit">保存</button>
    </form>
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
          <th>${labelForColumn("id")}</th>
          <th>${labelForColumn("user_id")}</th>
          <th>${labelForColumn("token_prefix")}</th>
          <th>${labelForColumn("name")}</th>
          <th>${labelForColumn("status")}</th>
          <th>${labelForColumn("expire_at")}</th>
          <th>${labelForColumn("quota_bytes")}</th>
          <th>${labelForColumn("used_total")}</th>
          <th>${labelForColumn("actions")}</th>
        </tr>
      </thead>
      <tbody>
        ${rows
          .map(
            (row) => `
              <tr>
                <td>${formatCell(row.id, "id")}</td>
                <td>${formatCell(row.user_id, "user_id")}</td>
                <td>${formatCell(row.token_prefix, "token_prefix")}</td>
                <td>${formatCell(row.name, "name")}</td>
                <td>${formatCell(row.status, "status")}</td>
                <td>${formatCell(row.expire_at, "expire_at")}</td>
                <td>${formatCell(row.quota_bytes, "quota_bytes")}</td>
                <td>${formatCell((row.used_upload_bytes || 0) + (row.used_download_bytes || 0), "used_total")}</td>
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

function renderTrafficTokens(rows) {
  if (!rows || rows.length === 0) {
    trafficTokensEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  renderTable(trafficTokensEl, rows, [
    "token_id",
    "user_id",
    "auth_user",
    "token_status",
    "today_total_bytes",
    "month_total_bytes",
    "used_upload_bytes",
    "used_download_bytes",
    "used_total_bytes",
    "quota_bytes",
    "updated_at",
  ]);
}

function renderTrafficOutbounds(rows) {
  if (!rows || rows.length === 0) {
    trafficOutboundsEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  renderTable(trafficOutboundsEl, rows, [
    "outbound_tag",
    "upstream_node_id",
    "node_name",
    "source_name",
    "upload_bytes",
    "download_bytes",
    "total_bytes",
    "updated_at",
  ]);
}

function renderTrafficDaily(rows) {
  renderTrafficBars(trafficDailyEl, rows, {
    valueKey: "total_bytes",
    label: (row) => String(row.day || "").slice(5),
    title: (row, total) => `${row.day} ${formatBytes(total)}`,
    className: "traffic-chart-bars",
  });
}

function renderTrafficHourly(rows) {
  renderTrafficBars(trafficHourlyEl, rows, {
    valueKey: "total_bytes",
    label: (row) => {
      const date = new Date(row.hour);
      if (Number.isNaN(date.getTime())) return "--";
      return `${String(date.getHours()).padStart(2, "0")}:00`;
    },
    title: (row, total) => `${row.hour} ${formatBytes(total)}`,
    className: "traffic-chart-bars traffic-chart-bars-hourly",
  });
}

function renderTrafficBars(target, rows, options) {
  if (!rows || rows.length === 0) {
    target.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  const maxTotal = Math.max(...rows.map((row) => row[options.valueKey] || 0), 1);
  target.innerHTML = `
    <div class="${options.className}">
      ${rows
        .map((row) => {
          const total = row[options.valueKey] || 0;
          const height = total > 0 ? Math.max(8, Math.round((total / maxTotal) * 118)) : 2;
          return `
            <div class="traffic-day" title="${escapeHTML(options.title(row, total))}">
              <div class="traffic-bar-track">
                <span class="traffic-bar" style="height: ${height}px"></span>
              </div>
              <span class="traffic-day-label">${escapeHTML(options.label(row))}</span>
              <strong>${escapeHTML(formatBytes(total))}</strong>
            </div>
          `;
        })
        .join("")}
    </div>
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
        <tr>${columns.map((column) => `<th>${labelForColumn(column)}</th>`).join("")}</tr>
      </thead>
      <tbody>
        ${rows
          .map(
            (row) =>
              `<tr>${columns
                .map((column) => `<td data-column="${escapeHTML(column)}">${formatCell(row[column], column)}</td>`)
                .join("")}</tr>`,
          )
          .join("")}
      </tbody>
    </table>
  `;
}

function labelForColumn(column) {
  return escapeHTML(columnLabels[column] || column);
}

function formatCell(value, column = "") {
  if (value === null || value === undefined || value === "") {
    return `<span class="cell-muted">--</span>`;
  }
  if (isStatusColumn(column)) {
    return formatStatus(value);
  }
  if (isBytesColumn(column)) {
    if (column === "quota_bytes" && Number(value || 0) === 0) {
      return `<span class="cell-muted">不限</span>`;
    }
    return escapeHTML(formatBytes(value));
  }
  if (Array.isArray(value)) {
    return value.length > 0 ? escapeHTML(value.join(", ")) : `<span class="cell-muted">--</span>`;
  }
  if (typeof value === "object") {
    return formatLongValue(JSON.stringify(value));
  }
  if (typeof value === "string" && value.length > 38) {
    return formatLongValue(value);
  }
  return escapeHTML(String(value));
}

function isStatusColumn(column) {
  return column === "status" || column.endsWith("_status");
}

function isBytesColumn(column) {
  return column === "used_total" || column.endsWith("_bytes");
}

function formatLongValue(value) {
  const trimmed = value.slice(0, 38);
  return `<code title="${escapeHTML(value)}">${escapeHTML(trimmed)}...</code>`;
}

function formatStatus(value) {
  const status = String(value || "").trim();
  const normalized = status.toLowerCase();
  let className = "status-badge";
  if (normalized === "active" || normalized === "enabled" || normalized === "ready") {
    className += " status-active";
  } else if (normalized === "over_quota" || normalized === "inactive" || normalized === "expired") {
    className += " status-warning";
  } else if (normalized === "revoked" || normalized === "disabled" || normalized === "error") {
    className += " status-danger";
  }
  return `<span class="${className}">${escapeHTML(status)}</span>`;
}

function formatBytes(value) {
  const bytes = Number(value || 0);
  if (bytes >= 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GiB`;
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MiB`;
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(1)} KiB`;
  return `${bytes} B`;
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
