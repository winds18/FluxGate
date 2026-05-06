const statusEl = document.querySelector("#status");
const loginView = document.querySelector("#login-view");
const appView = document.querySelector("#app-view");
const viewTitleEl = document.querySelector("#view-title");
const viewDescriptionEl = document.querySelector("#view-description");
const viewContextEl = document.querySelector("#view-context");
const viewRailEl = document.querySelector("#view-rail");
const dashboardViewSections = Array.from(document.querySelectorAll("[data-dashboard-view]"));
const dashboardNavButtons = Array.from(document.querySelectorAll("[data-view-nav]"));
const dashboardJumpButtons = Array.from(document.querySelectorAll("[data-view-jump]"));
const loginForm = document.querySelector("#login-form");
const loginErrorEl = document.querySelector("#login-error");
const loginUsernameEl = document.querySelector("#login-username");
const loginPasswordEl = document.querySelector("#login-password");
const logoutEl = document.querySelector("#logout");
const metricsEl = document.querySelector("#metrics");
const overviewReadinessEl = document.querySelector("#overview-readiness");
const overviewNextStepEl = document.querySelector("#overview-next-step");
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
const panelCountEls = {
  sources: document.querySelector("#sources-panel-count"),
  nodes: document.querySelector("#nodes-panel-count"),
  virtualNodes: document.querySelector("#virtual-nodes-panel-count"),
  teams: document.querySelector("#teams-panel-count"),
  users: document.querySelector("#users-panel-count"),
  tokens: document.querySelector("#tokens-panel-count"),
  policies: document.querySelector("#policies-panel-count"),
  traffic: document.querySelector("#traffic-panel-count"),
  config: document.querySelector("#config-panel-count"),
};
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
const displayTimeZone = "Asia/Shanghai";
const displayDateTimeFormatter = new Intl.DateTimeFormat("zh-CN", {
  timeZone: displayTimeZone,
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hourCycle: "h23",
});
const displayHourFormatter = new Intl.DateTimeFormat("zh-CN", {
  timeZone: displayTimeZone,
  hour: "2-digit",
  hourCycle: "h23",
});
const dashboardViewMeta = {
  overview: ["概览", "运行状态、资源规模和下一步入口"],
  access: ["接入", "维护上游来源、刷新订阅和导入节点"],
  nodes: ["节点", "按地区聚合节点，查看详情并维护虚拟网关"],
  identity: ["身份", "管理团队、成员、Token 和可复制订阅地址"],
  policies: ["策略", "控制 Token、成员和团队可见的节点范围"],
  traffic: ["流量", "查看用量、额度消耗和上游出口流量"],
  ops: ["运维", "检查、发布、回滚并重启 sing-box 配置"],
};
const dashboardViewRail = {
  overview: [
    ["运行概览", "metrics"],
    ["测试闭环", "overview-readiness"],
    ["下一步", "overview-next-step"],
  ],
  access: [
    ["上游来源", "sources"],
    ["添加来源", "source-form", true],
    ["导入节点", "node-import-form", true],
  ],
  nodes: [
    ["节点池", "nodes"],
    ["虚拟节点", "virtual-nodes"],
    ["创建虚拟节点", "virtual-node-form", true],
  ],
  identity: [
    ["Token", "tokens"],
    ["团队", "teams"],
    ["成员", "users"],
    ["创建 Token", "token-form", true],
    ["创建团队", "team-form", true],
    ["创建成员", "user-form", true],
  ],
  policies: [
    ["策略", "policies"],
    ["创建策略", "policy-form", true],
  ],
  traffic: [
    ["Token 用量", "traffic-tokens"],
    ["小时曲线", "traffic-hourly"],
    ["每日曲线", "traffic-daily"],
    ["出口摘要", "traffic-outbounds"],
  ],
  ops: [
    ["配置操作", "config-check-result"],
  ],
};
let activeDashboardView = "overview";

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
teamsEl.addEventListener("click", handleTeamAction);
usersEl.addEventListener("click", handleUserAction);
sourcesEl.addEventListener("click", handleSourceAction);
nodesEl.addEventListener("click", handleNodeAction);
nodesEl.addEventListener("submit", handleNodeEditSubmit);
virtualNodesEl.addEventListener("click", handleVirtualNodeAction);
policiesEl.addEventListener("click", handlePolicyAction);
tokensEl.addEventListener("click", handleTokenAction);
dashboardNavButtons.forEach((button) => {
  button.addEventListener("click", () => setActiveView(button.dataset.viewNav || "overview"));
});
dashboardJumpButtons.forEach((button) => {
  button.addEventListener("click", () => setActiveView(button.dataset.viewJump || "overview"));
});
document.addEventListener("click", (event) => {
  const button = event.target.closest("[data-form-drawer-toggle]");
  if (!button) return;
  const drawer = button.closest("[data-form-drawer]");
  if (!drawer) return;
  const nextCollapsed = !drawer.classList.contains("is-collapsed");
  setFormDrawerCollapsed(drawer, nextCollapsed);
});
document.addEventListener("click", (event) => {
  const button = event.target.closest("[data-overview-jump]");
  if (!button) return;
  setActiveView(button.dataset.overviewJump || "overview");
});
document.addEventListener("click", (event) => {
  const button = event.target.closest("[data-view-rail-target]");
  if (!button) return;
  const target = document.getElementById(button.dataset.viewRailTarget || "");
  if (!target) return;
  if (button.dataset.viewRailExpand === "true") {
    const drawer = target.closest("[data-form-drawer]");
    if (drawer) {
      setFormDrawerCollapsed(drawer, false);
    }
  }
  const panelTarget = target.closest(".panel") || target;
  panelTarget.scrollIntoView({ behavior: "smooth", block: "start", inline: "nearest" });
});
bootstrap();

let appState = {
  overview: {},
  teams: [],
  users: [],
  sources: [],
  nodes: [],
  virtualNodes: [],
  policies: [],
  tokens: [],
  trafficHourly: [],
  trafficDaily: [],
  trafficOutbounds: [],
  trafficTokens: [],
  editingTeamID: null,
  editingUserID: null,
  editingSourceID: null,
  editingVirtualNodeID: null,
  editingPolicyID: null,
  editingNodeID: null,
  expandedNodeRegion: null,
  expandedNodeID: null,
  nodeDetail: null,
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
  source_id: "来源 ID",
  source_name: "来源",
  url: "URL",
  raw_name: "原始名称",
  display_name: "展示名称",
  name_mode: "命名模式",
  protocol: "协议",
  server: "服务器",
  server_port: "端口",
  region: "地区",
  uri: "节点 URI",
  uri_hash: "URI Hash",
  tags: "标签",
  listen_protocol: "监听协议",
  listen_port: "端口",
  tag_selector: "标签选择器",
  strategy: "出口策略",
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
  subscriptions: "订阅地址",
  auth_user: "网关用户",
  token_status: "Token 状态",
  today_total_bytes: "今日",
  month_total_bytes: "本月",
  used_upload_bytes: "累计上传",
  used_download_bytes: "累计下载",
  used_total_bytes: "累计总量",
  quota_usage: "额度使用率",
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
  last_seen_at: "上次出现",
  last_checked_at: "上次检测",
  last_error: "错误",
  created_at: "创建时间",
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
  appState = {
    overview: {},
    teams: [],
    users: [],
    sources: [],
    nodes: [],
    virtualNodes: [],
    policies: [],
    tokens: [],
    trafficHourly: [],
    trafficDaily: [],
    trafficOutbounds: [],
    trafficTokens: [],
    editingTeamID: null,
    editingUserID: null,
    editingSourceID: null,
    editingVirtualNodeID: null,
    editingPolicyID: null,
    editingNodeID: null,
    expandedNodeRegion: null,
    expandedNodeID: null,
    nodeDetail: null,
  };
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
  setActiveView(activeDashboardView);
}

function setActiveView(view) {
  const nextView = dashboardViewMeta[view] ? view : "overview";
  activeDashboardView = nextView;
  dashboardViewSections.forEach((section) => {
    section.hidden = section.dataset.dashboardView !== nextView;
  });
  dashboardNavButtons.forEach((button) => {
    const isActive = button.dataset.viewNav === nextView;
    button.classList.toggle("is-active", isActive);
    button.setAttribute("aria-current", isActive ? "page" : "false");
  });
  const [title, description] = dashboardViewMeta[nextView];
  viewTitleEl.textContent = title;
  viewDescriptionEl.textContent = description;
  updateViewContext();
  updateViewRail();
}

function setFormDrawerCollapsed(drawer, collapsed) {
  drawer.classList.toggle("is-collapsed", collapsed);
  const toggle = drawer.querySelector("[data-form-drawer-toggle]");
  if (toggle) {
    toggle.textContent = collapsed ? "展开" : "收起";
  }
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
    appState = {
      ...appState,
      overview,
      teams,
      users,
      sources,
      nodes,
      virtualNodes,
      policies,
      tokens,
      trafficHourly,
      trafficDaily,
      trafficOutbounds,
      trafficTokens,
    };
    renderMetrics(overview);
    renderOverviewReadiness(overview);
    renderSelectors();
    renderTeams(teams);
    renderUsers(users);
    renderSources(sources);
    renderNodes(nodes);
    renderVirtualNodes(virtualNodes);
    renderPolicies(policies);
    renderTokens(tokens);
    renderTrafficHourly(trafficHourly);
    renderTrafficDaily(trafficDaily);
    renderTrafficOutbounds(trafficOutbounds);
    renderTrafficTokens(trafficTokens);
    updatePanelCounts();
    updateViewContext();
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
  showTokenSubscriptionResult(result, "订阅地址");
  tokenForm.reset();
}

async function handleTeamAction(event) {
  const button = event.target.closest("button[data-team-action]");
  if (!button) return;
  const id = button.dataset.teamId;
  const action = button.dataset.teamAction;
  if (action === "edit") {
    appState.editingTeamID = Number.parseInt(id || "0", 10);
    renderTeams(appState.teams);
    teamsEl.querySelector(`tr[data-team-id="${id}"] input[data-team-field="name"]`)?.focus();
    return;
  }
  if (action === "cancel") {
    appState.editingTeamID = null;
    renderTeams(appState.teams);
    return;
  }
  if (action === "save") {
    await saveTeam(button);
  }
}

async function saveTeam(button) {
  const row = button.closest("tr[data-team-id]");
  if (!row) return;
  const id = row.dataset.teamId;
  const payload = {
    name: teamFieldValue(row, "name"),
    description: teamFieldValue(row, "description"),
    status: teamFieldValue(row, "status") || "active",
  };
  row.querySelectorAll("button").forEach((item) => {
    item.disabled = true;
  });
  statusEl.textContent = "保存团队中";
  try {
    await patchJSON(`/api/teams/${id}`, payload);
    appState.editingTeamID = null;
    await load();
  } catch (error) {
    statusEl.textContent = "保存失败";
    row.querySelectorAll("button").forEach((item) => {
      item.disabled = false;
    });
  }
}

function teamFieldValue(row, name) {
  return String(row.querySelector(`[data-team-field="${name}"]`)?.value || "").trim();
}

async function handleUserAction(event) {
  const button = event.target.closest("button[data-user-action]");
  if (!button) return;
  const id = button.dataset.userId;
  const action = button.dataset.userAction;
  if (action === "edit") {
    appState.editingUserID = Number.parseInt(id || "0", 10);
    renderUsers(appState.users);
    usersEl.querySelector(`tr[data-user-id="${id}"] input[data-user-field="name"]`)?.focus();
    return;
  }
  if (action === "cancel") {
    appState.editingUserID = null;
    renderUsers(appState.users);
    return;
  }
  if (action === "save") {
    await saveUser(button);
  }
}

async function saveUser(button) {
  const row = button.closest("tr[data-user-id]");
  if (!row) return;
  const id = row.dataset.userId;
  const teamID = Number.parseInt(row.querySelector('[data-user-field="team_id"]')?.value || "0", 10);
  const payload = {
    team_id: Number.isFinite(teamID) && teamID > 0 ? teamID : null,
    name: userFieldValue(row, "name"),
    email: userFieldValue(row, "email"),
    remark: userFieldValue(row, "remark"),
    status: userFieldValue(row, "status") || "active",
  };
  row.querySelectorAll("button").forEach((item) => {
    item.disabled = true;
  });
  statusEl.textContent = "保存成员中";
  try {
    await patchJSON(`/api/users/${id}`, payload);
    appState.editingUserID = null;
    await load();
  } catch (error) {
    statusEl.textContent = "保存失败";
    row.querySelectorAll("button").forEach((item) => {
      item.disabled = false;
    });
  }
}

function userFieldValue(row, name) {
  return String(row.querySelector(`[data-user-field="${name}"]`)?.value || "").trim();
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

async function handleVirtualNodeAction(event) {
  const button = event.target.closest("button[data-virtual-node-action]");
  if (!button) return;
  const id = button.dataset.virtualNodeId;
  const action = button.dataset.virtualNodeAction;
  if (action === "edit") {
    appState.editingVirtualNodeID = Number.parseInt(id || "0", 10);
    renderVirtualNodes(appState.virtualNodes);
    virtualNodesEl.querySelector(`tr[data-virtual-node-id="${id}"] input[data-virtual-node-field="name"]`)?.focus();
    return;
  }
  if (action === "cancel") {
    appState.editingVirtualNodeID = null;
    renderVirtualNodes(appState.virtualNodes);
    return;
  }
  if (action === "save") {
    await saveVirtualNode(button);
  }
}

async function saveVirtualNode(button) {
  const row = button.closest("tr[data-virtual-node-id]");
  if (!row) return;
  const id = row.dataset.virtualNodeId;
  const listenPort = Number.parseInt(row.querySelector('[data-virtual-node-field="listen_port"]')?.value || "0", 10);
  const payload = {
    name: virtualNodeFieldValue(row, "name"),
    listen_protocol: virtualNodeFieldValue(row, "listen_protocol") || "vless",
    listen_port: Number.isFinite(listenPort) ? listenPort : 0,
    tag_selector: virtualNodeFieldValue(row, "tag_selector") || "{}",
    strategy: virtualNodeFieldValue(row, "strategy") || "selector",
    status: virtualNodeFieldValue(row, "status") || "active",
  };
  row.querySelectorAll("button").forEach((item) => {
    item.disabled = true;
  });
  statusEl.textContent = "保存虚拟节点中";
  try {
    await patchJSON(`/api/virtual-nodes/${id}`, payload);
    appState.editingVirtualNodeID = null;
    await load();
  } catch (error) {
    statusEl.textContent = "保存失败";
    row.querySelectorAll("button").forEach((item) => {
      item.disabled = false;
    });
  }
}

function virtualNodeFieldValue(row, name) {
  return String(row.querySelector(`[data-virtual-node-field="${name}"]`)?.value || "").trim();
}

async function handlePolicyAction(event) {
  const button = event.target.closest("button[data-policy-action]");
  if (!button) return;
  const id = button.dataset.policyId;
  const action = button.dataset.policyAction;
  if (action === "edit") {
    appState.editingPolicyID = Number.parseInt(id || "0", 10);
    renderPolicies(appState.policies);
    policiesEl.querySelector(`tr[data-policy-id="${id}"] input[data-policy-field="name"]`)?.focus();
    return;
  }
  if (action === "cancel") {
    appState.editingPolicyID = null;
    renderPolicies(appState.policies);
    return;
  }
  if (action === "save") {
    await savePolicy(button);
  }
}

async function savePolicy(button) {
  const row = button.closest("tr[data-policy-id]");
  if (!row) return;
  const id = row.dataset.policyId;
  const scopeID = Number.parseInt(row.querySelector('[data-policy-field="scope_id"]')?.value || "0", 10);
  const maxNodes = Number.parseInt(row.querySelector('[data-policy-field="max_nodes"]')?.value || "0", 10);
  const payload = {
    name: policyFieldValue(row, "name"),
    scope_type: policyFieldValue(row, "scope_type") || "team",
    scope_id: Number.isFinite(scopeID) && scopeID > 0 ? scopeID : null,
    include_tags: policyFieldValue(row, "include_tags"),
    exclude_tags: policyFieldValue(row, "exclude_tags"),
    allowed_virtual_nodes: policyFieldValue(row, "allowed_virtual_nodes"),
    max_nodes: Number.isFinite(maxNodes) ? maxNodes : 0,
    status: policyFieldValue(row, "status") || "active",
  };
  row.querySelectorAll("button").forEach((item) => {
    item.disabled = true;
  });
  statusEl.textContent = "保存策略中";
  try {
    await patchJSON(`/api/policies/${id}`, payload);
    appState.editingPolicyID = null;
    await load();
  } catch (error) {
    statusEl.textContent = "保存失败";
    row.querySelectorAll("button").forEach((item) => {
      item.disabled = false;
    });
  }
}

function policyFieldValue(row, name) {
  return String(row.querySelector(`[data-policy-field="${name}"]`)?.value || "").trim();
}

async function handleNodeAction(event) {
  const regionButton = event.target.closest("button[data-node-region-action]");
  if (regionButton) {
    const action = regionButton.dataset.nodeRegionAction;
    if (action === "open") {
      appState.expandedNodeRegion = regionButton.dataset.nodeRegion || "";
      appState.editingNodeID = null;
      appState.expandedNodeID = null;
      appState.nodeDetail = null;
      renderNodes(appState.nodes);
      return;
    }
    if (action === "back") {
      appState.expandedNodeRegion = null;
      appState.editingNodeID = null;
      appState.expandedNodeID = null;
      appState.nodeDetail = null;
      renderNodes(appState.nodes);
      return;
    }
  }
  const button = event.target.closest("button[data-node-action]");
  if (!button) return;
  const id = Number.parseInt(button.dataset.nodeId || "0", 10);
  if (!id) return;
  const action = button.dataset.nodeAction;
  if (action === "edit") {
    appState.editingNodeID = id;
    appState.expandedNodeID = null;
    appState.nodeDetail = null;
    renderNodes(appState.nodes);
    nodesEl.querySelector(`form[data-node-id="${id}"] input[name="display_name"]`)?.focus();
    return;
  }
  if (action === "cancel") {
    appState.editingNodeID = null;
    renderNodes(appState.nodes);
    return;
  }
  if (action === "detail") {
    if (appState.expandedNodeID === id) {
      appState.expandedNodeID = null;
      appState.nodeDetail = null;
      renderNodes(appState.nodes);
      return;
    }
    button.disabled = true;
    statusEl.textContent = "加载节点详情中";
    try {
      const detail = await getJSON(`/api/nodes/${id}`);
      appState.expandedNodeID = id;
      appState.nodeDetail = detail;
      appState.editingNodeID = null;
      renderNodes(appState.nodes);
      statusEl.textContent = "已连接";
    } catch (error) {
      appState.expandedNodeID = null;
      appState.nodeDetail = null;
      statusEl.textContent = "详情加载失败";
      button.disabled = false;
    }
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
    if (action === "copy-subscription") {
      await copyText(button.dataset.tokenUrl || "");
      statusEl.textContent = "订阅地址已复制";
      button.disabled = false;
      return;
    }
    if (action === "extend") {
      const input = button.closest("[data-token-action-group]")?.querySelector("[data-token-extend-days]");
      const extendDays = numberInputValue(input, 30);
      if (extendDays <= 0) throw new Error("extend days must be greater than 0");
      await postJSON(`/api/tokens/${id}/extend`, { extend_days: extendDays });
    } else if (action === "quota") {
      const input = button.closest("[data-token-action-group]")?.querySelector("[data-token-quota-mib]");
      const quotaMiB = numberInputValue(input, 1024);
      if (quotaMiB <= 0) throw new Error("quota MiB must be greater than 0");
      await postJSON(`/api/tokens/${id}/quota`, { quota_bytes: quotaMiB * 1024 * 1024 });
    } else if (action === "revoke") {
      await postJSON(`/api/tokens/${id}/revoke`, {});
    } else if (action === "restore") {
      await postJSON(`/api/tokens/${id}/restore`, {});
    } else if (action === "rotate-subscription") {
      const result = await postJSON(`/api/tokens/${id}/rotate-subscription`, {});
      showTokenSubscriptionResult(result, "订阅地址已重置");
    }
    await load();
  } catch (error) {
    statusEl.textContent = "更新失败";
    button.disabled = false;
  }
}

async function copyText(text) {
  if (!text) throw new Error("empty text");
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "readonly");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand("copy");
  textarea.remove();
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

function numberInputValue(input, fallback) {
  const value = Number.parseInt(String(input?.value || fallback || "0"), 10);
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

function overviewReadinessChecks(data) {
  return [
    ["接入来源", Number(data.sources || 0) > 0, `${data.sources || 0} 个来源`, "access"],
    ["节点池", Number(data.nodes || 0) > 0, `${data.nodes || 0} 个节点`, "nodes"],
    ["虚拟网关", Number(data.virtual_nodes || 0) > 0, `${data.virtual_nodes || 0} 个入口`, "nodes"],
    ["团队 Token", Number(data.tokens || 0) > 0, `${data.tokens || 0} 个 Token`, "identity"],
    ["访问策略", Number(data.policies || 0) > 0, `${data.policies || 0} 条策略`, "policies"],
  ];
}

function renderOverviewReadiness(data) {
  const checks = overviewReadinessChecks(data);
  const readyCount = checks.filter(([, ready]) => ready).length;
  const firstMissing = checks.find(([, ready]) => !ready);
  const next = firstMissing
    ? {
        title: `补齐${firstMissing[0]}`,
        description: "补齐后再回到运维模块检查并发布 sing-box 配置，订阅地址才更接近真实客户端体验。",
        view: firstMissing[3],
        action: `去${dashboardViewMeta[firstMissing[3]]?.[0] || "处理"}`,
      }
    : {
        title: "闭环已具备真实测试条件",
        description: "基础数据已齐备，可以发布网关配置，再复制 Clash/Mihomo 或 sing-box 订阅地址做客户端导入验证。",
        view: "ops",
        action: "去运维发布",
      };

  overviewReadinessEl.innerHTML = `
    <div class="overview-panel-heading">
      <span>真实测试闭环</span>
      <strong>${readyCount}/${checks.length}</strong>
    </div>
    <div class="readiness-list" data-overview-readiness>
      ${checks
        .map(
          ([label, ready, summary, view]) => `
            <button class="readiness-item ${ready ? "is-ready" : ""}" type="button" data-overview-jump="${view}">
              <span class="readiness-dot"></span>
              <span>
                <strong>${escapeHTML(label)}</strong>
                <small>${escapeHTML(summary)}</small>
              </span>
            </button>
          `,
        )
        .join("")}
    </div>
  `;
  overviewNextStepEl.innerHTML = `
    <div class="overview-panel-heading">
      <span>下一步</span>
      <strong>${firstMissing ? "待补齐" : "可测试"}</strong>
    </div>
    <p>${escapeHTML(next.description)}</p>
    <button class="primary-link-button" type="button" data-overview-jump="${next.view}">
      ${escapeHTML(next.action)}
    </button>
  `;
}

function updateViewContext() {
  if (!viewContextEl) return;
  const items = viewContextItems(activeDashboardView);
  viewContextEl.hidden = items.length === 0;
  viewContextEl.innerHTML = items.map(renderContextChip).join("");
}

function updateViewRail() {
  if (!viewRailEl) return;
  const items = dashboardViewRail[activeDashboardView] || [];
  viewRailEl.hidden = items.length === 0;
  viewRailEl.innerHTML = items.map(renderViewRailButton).join("");
}

function renderViewRailButton([label, target, expand]) {
  return `
    <button class="view-rail-button" type="button" data-view-rail-target="${escapeHTML(target)}" data-view-rail-expand="${expand ? "true" : "false"}">
      ${escapeHTML(label)}
    </button>
  `;
}

function updatePanelCounts() {
  const nodes = appState.nodes || [];
  const virtualNodes = appState.virtualNodes || [];
  const tokens = appState.tokens || [];
  const activeNodes = countBy(nodes, (row) => row.status === "active");
  const activeVirtualNodes = countBy(virtualNodes, (row) => row.status === "active");
  const activeTokens = countBy(tokens, (row) => row.status === "active");
  const configReady = activeNodes > 0 && activeVirtualNodes > 0 && activeTokens > 0;

  setPanelCount("sources", `${formatPlainNumber(appState.sources.length)} 个`);
  setPanelCount("nodes", `${formatPlainNumber(nodes.length)} 个 / ${formatPlainNumber(groupNodesByRegion(nodes).length)} 地区`);
  setPanelCount("virtualNodes", `${formatPlainNumber(virtualNodes.length)} 个`);
  setPanelCount("teams", `${formatPlainNumber(appState.teams.length)} 个`);
  setPanelCount("users", `${formatPlainNumber(appState.users.length)} 个`);
  setPanelCount("tokens", `${formatPlainNumber(tokens.length)} 个`);
  setPanelCount("policies", `${formatPlainNumber(appState.policies.length)} 条`);
  setPanelCount("traffic", `${formatPlainNumber(appState.trafficTokens.length)} Token`);
  setPanelCount("config", configReady ? "就绪" : "待补齐", configReady ? "success" : "warning");
}

function setPanelCount(key, value, tone = "") {
  const element = panelCountEls[key];
  if (!element) return;
  element.textContent = value;
  element.classList.toggle("panel-count-success", tone === "success");
  element.classList.toggle("panel-count-warning", tone === "warning");
}

function viewContextItems(view) {
  const overview = appState.overview || {};
  const sources = appState.sources || [];
  const nodes = appState.nodes || [];
  const virtualNodes = appState.virtualNodes || [];
  const teams = appState.teams || [];
  const users = appState.users || [];
  const policies = appState.policies || [];
  const tokens = appState.tokens || [];
  const trafficHourly = appState.trafficHourly || [];
  const trafficDaily = appState.trafficDaily || [];
  const trafficOutbounds = appState.trafficOutbounds || [];
  const trafficTokens = appState.trafficTokens || [];
  const activeNodes = countBy(nodes, (row) => row.status === "active");
  const activeTokens = countBy(tokens, (row) => row.status === "active");
  const activePolicies = countBy(policies, (row) => row.status === "active");
  const readiness = overviewReadinessChecks(overview);
  const readyCount = readiness.filter(([, ready]) => ready).length;

  if (view === "access") {
    const subscriptionSources = countBy(sources, (row) => row.type === "subscription");
    const erroredSources = countBy(sources, (row) => row.last_error);
    return [
      contextItem("来源", sources.length),
      contextItem("订阅源", subscriptionSources),
      contextItem("异常", erroredSources, erroredSources > 0 ? "warning" : "success"),
    ];
  }
  if (view === "nodes") {
    const regions = groupNodesByRegion(nodes).length;
    return [
      contextItem("地区", regions),
      contextItem("节点", nodes.length),
      contextItem("可用", activeNodes, activeNodes > 0 ? "success" : "warning"),
      contextItem("虚拟网关", virtualNodes.length, virtualNodes.length > 0 ? "success" : "warning"),
    ];
  }
  if (view === "identity") {
    return [
      contextItem("团队", teams.length),
      contextItem("成员", users.length),
      contextItem("Token", tokens.length),
      contextItem("有效", activeTokens, activeTokens > 0 ? "success" : "warning"),
    ];
  }
  if (view === "policies") {
    const constrainedPolicies = countBy(
      policies,
      (row) => row.allowed_virtual_nodes || row.include_tags || row.exclude_tags || Number(row.max_nodes || 0) > 0,
    );
    return [
      contextItem("策略", policies.length),
      contextItem("生效", activePolicies, activePolicies > 0 ? "success" : "warning"),
      contextItem("限制项", constrainedPolicies),
    ];
  }
  if (view === "traffic") {
    return [
      contextItem("24h 样本", trafficHourly.length),
      contextItem("14天样本", trafficDaily.length),
      contextItem("Token 用量", trafficTokens.length),
      contextItem("出口摘要", trafficOutbounds.length),
    ];
  }
  if (view === "ops") {
    return [
      contextItem("虚拟网关", virtualNodes.length, virtualNodes.length > 0 ? "success" : "warning"),
      contextItem("活跃 Token", activeTokens, activeTokens > 0 ? "success" : "warning"),
      contextItem("可用节点", activeNodes, activeNodes > 0 ? "success" : "warning"),
    ];
  }
  return [
    contextItem("闭环", `${readyCount}/${readiness.length}`, readyCount === readiness.length ? "success" : "warning"),
    contextItem("节点", Number(overview.nodes || nodes.length || 0)),
    contextItem("Token", Number(overview.tokens || tokens.length || 0)),
    contextItem("策略", Number(overview.policies || policies.length || 0)),
  ];
}

function contextItem(label, value, tone = "") {
  return { label, value, tone };
}

function renderContextChip(item) {
  const toneClass = item.tone ? ` context-chip-${item.tone}` : "";
  return `
    <span class="context-chip${toneClass}">
      <span>${escapeHTML(item.label)}</span>
      <strong>${escapeHTML(String(item.value ?? 0))}</strong>
    </span>
  `;
}

function countBy(rows, predicate) {
  return (rows || []).filter(predicate).length;
}

function formatPlainNumber(value) {
  return new Intl.NumberFormat("zh-CN").format(Number(value || 0));
}

function renderTeams(rows) {
  appState.teams = rows || [];
  if (!rows || rows.length === 0) {
    teamsEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  teamsEl.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>${labelForColumn("id")}</th>
          <th>${labelForColumn("name")}</th>
          <th>${labelForColumn("description")}</th>
          <th>${labelForColumn("status")}</th>
          <th>${labelForColumn("actions")}</th>
        </tr>
      </thead>
      <tbody>
        ${rows.map((row) => renderTeamRow(row)).join("")}
      </tbody>
    </table>
  `;
}

function renderTeamRow(row) {
  const isEditing = appState.editingTeamID === row.id;
  return `
    <tr data-team-id="${row.id}" class="${isEditing ? "team-edit-row" : ""}">
      <td>${formatCell(row.id, "id")}</td>
      <td>${isEditing ? teamTextInput(row, "name") : formatCell(row.name, "name")}</td>
      <td>${isEditing ? teamTextInput(row, "description", "table-edit-input-wide") : formatCell(row.description, "description")}</td>
      <td>${isEditing ? teamStatusSelect(row.status) : formatCell(row.status, "status")}</td>
      <td class="table-actions compact-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-team-action="save" data-team-id="${row.id}">保存</button>
               <button class="table-button ghost-button" type="button" data-team-action="cancel" data-team-id="${row.id}">取消</button>`
            : `<button class="table-button" type="button" data-team-action="edit" data-team-id="${row.id}">编辑</button>`
        }
      </td>
    </tr>
  `;
}

function teamTextInput(row, field, className = "table-edit-input") {
  return `<input class="${className}" data-team-field="${field}" value="${escapeHTML(String(row[field] || ""))}" />`;
}

function teamStatusSelect(value) {
  const current = String(value || "active");
  return `
    <select class="table-edit-input" data-team-field="status">
      <option value="active" ${current === "active" ? "selected" : ""}>active</option>
      <option value="inactive" ${current === "inactive" ? "selected" : ""}>inactive</option>
    </select>
  `;
}

function renderUsers(rows) {
  appState.users = rows || [];
  if (!rows || rows.length === 0) {
    usersEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  usersEl.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>${labelForColumn("id")}</th>
          <th>${labelForColumn("team_id")}</th>
          <th>${labelForColumn("name")}</th>
          <th>${labelForColumn("email")}</th>
          <th>${labelForColumn("remark")}</th>
          <th>${labelForColumn("status")}</th>
          <th>${labelForColumn("actions")}</th>
        </tr>
      </thead>
      <tbody>
        ${rows.map((row) => renderUserRow(row)).join("")}
      </tbody>
    </table>
  `;
}

function renderUserRow(row) {
  const isEditing = appState.editingUserID === row.id;
  return `
    <tr data-user-id="${row.id}" class="${isEditing ? "user-edit-row" : ""}">
      <td>${formatCell(row.id, "id")}</td>
      <td>${isEditing ? userTeamSelectInput(row.team_id) : formatCell(row.team_id, "team_id")}</td>
      <td>${isEditing ? userTextInput(row, "name") : formatCell(row.name, "name")}</td>
      <td>${isEditing ? userTextInput(row, "email", "table-edit-input-wide") : formatCell(row.email, "email")}</td>
      <td>${isEditing ? userTextInput(row, "remark", "table-edit-input-wide") : formatCell(row.remark, "remark")}</td>
      <td>${isEditing ? userStatusSelect(row.status) : formatCell(row.status, "status")}</td>
      <td class="table-actions compact-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-user-action="save" data-user-id="${row.id}">保存</button>
               <button class="table-button ghost-button" type="button" data-user-action="cancel" data-user-id="${row.id}">取消</button>`
            : `<button class="table-button" type="button" data-user-action="edit" data-user-id="${row.id}">编辑</button>`
        }
      </td>
    </tr>
  `;
}

function userTextInput(row, field, className = "table-edit-input") {
  return `<input class="${className}" data-user-field="${field}" value="${escapeHTML(String(row[field] || ""))}" />`;
}

function userTeamSelectInput(value) {
  const current = String(value || "");
  const options = [`<option value="">不绑定团队</option>`].concat(
    appState.teams.map((team) => `<option value="${team.id}" ${String(team.id) === current ? "selected" : ""}>${escapeHTML(team.name)}</option>`),
  );
  return `<select class="table-edit-input" data-user-field="team_id">${options.join("")}</select>`;
}

function userStatusSelect(value) {
  const current = String(value || "active");
  return `
    <select class="table-edit-input" data-user-field="status">
      <option value="active" ${current === "active" ? "selected" : ""}>active</option>
      <option value="inactive" ${current === "inactive" ? "selected" : ""}>inactive</option>
    </select>
  `;
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

function renderVirtualNodes(rows) {
  appState.virtualNodes = rows || [];
  if (!rows || rows.length === 0) {
    virtualNodesEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  virtualNodesEl.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>${labelForColumn("id")}</th>
          <th>${labelForColumn("name")}</th>
          <th>${labelForColumn("listen_protocol")}</th>
          <th>${labelForColumn("listen_port")}</th>
          <th>${labelForColumn("tag_selector")}</th>
          <th>${labelForColumn("strategy")}</th>
          <th>${labelForColumn("status")}</th>
          <th>${labelForColumn("actions")}</th>
        </tr>
      </thead>
      <tbody>
        ${rows.map((row) => renderVirtualNodeRow(row)).join("")}
      </tbody>
    </table>
  `;
}

function renderVirtualNodeRow(row) {
  const isEditing = appState.editingVirtualNodeID === row.id;
  return `
    <tr data-virtual-node-id="${row.id}" class="${isEditing ? "virtual-node-edit-row" : ""}">
      <td>${formatCell(row.id, "id")}</td>
      <td>${isEditing ? virtualNodeTextInput(row, "name") : formatCell(row.name, "name")}</td>
      <td>${isEditing ? virtualNodeProtocolSelect(row.listen_protocol) : formatCell(row.listen_protocol, "listen_protocol")}</td>
      <td>${isEditing ? virtualNodeNumberInput(row, "listen_port") : formatCell(row.listen_port, "listen_port")}</td>
      <td>${isEditing ? virtualNodeTextInput(row, "tag_selector", "table-edit-input-wide", '{"include":["HK"]}') : formatCell(row.tag_selector, "tag_selector")}</td>
      <td>${isEditing ? virtualNodeStrategySelect(row.strategy) : formatCell(row.strategy, "strategy")}</td>
      <td>${isEditing ? virtualNodeStatusSelect(row.status) : formatCell(row.status, "status")}</td>
      <td class="table-actions virtual-node-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-virtual-node-action="save" data-virtual-node-id="${row.id}">保存</button>
               <button class="table-button ghost-button" type="button" data-virtual-node-action="cancel" data-virtual-node-id="${row.id}">取消</button>`
            : `<button class="table-button" type="button" data-virtual-node-action="edit" data-virtual-node-id="${row.id}">编辑</button>`
        }
      </td>
    </tr>
  `;
}

function virtualNodeTextInput(row, field, className = "table-edit-input", placeholder = "") {
  return `<input class="${className}" data-virtual-node-field="${field}" value="${escapeHTML(String(row[field] || ""))}" placeholder="${escapeHTML(placeholder)}" />`;
}

function virtualNodeNumberInput(row, field) {
  return `<input class="table-edit-input table-edit-number" data-virtual-node-field="${field}" type="number" min="1" max="65535" step="1" value="${escapeHTML(String(row[field] || 0))}" />`;
}

function virtualNodeProtocolSelect(value) {
  const current = String(value || "vless");
  return `
    <select class="table-edit-input" data-virtual-node-field="listen_protocol">
      <option value="vless" ${current === "vless" ? "selected" : ""}>vless</option>
    </select>
  `;
}

function virtualNodeStrategySelect(value) {
  const current = String(value || "selector");
  return `
    <select class="table-edit-input" data-virtual-node-field="strategy">
      <option value="selector" ${current === "selector" ? "selected" : ""}>selector</option>
    </select>
  `;
}

function virtualNodeStatusSelect(value) {
  const current = String(value || "active");
  return `
    <select class="table-edit-input" data-virtual-node-field="status">
      <option value="active" ${current === "active" ? "selected" : ""}>active</option>
      <option value="inactive" ${current === "inactive" ? "selected" : ""}>inactive</option>
    </select>
  `;
}

function renderPolicies(rows) {
  appState.policies = rows || [];
  if (!rows || rows.length === 0) {
    policiesEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  policiesEl.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>${labelForColumn("id")}</th>
          <th>${labelForColumn("name")}</th>
          <th>${labelForColumn("scope_type")}</th>
          <th>${labelForColumn("scope_id")}</th>
          <th>${labelForColumn("include_tags")}</th>
          <th>${labelForColumn("exclude_tags")}</th>
          <th>${labelForColumn("allowed_virtual_nodes")}</th>
          <th>${labelForColumn("max_nodes")}</th>
          <th>${labelForColumn("status")}</th>
          <th>${labelForColumn("actions")}</th>
        </tr>
      </thead>
      <tbody>
        ${rows.map((row) => renderPolicyRow(row)).join("")}
      </tbody>
    </table>
  `;
}

function renderPolicyRow(row) {
  const isEditing = appState.editingPolicyID === row.id;
  return `
    <tr data-policy-id="${row.id}" class="${isEditing ? "policy-edit-row" : ""}">
      <td>${formatCell(row.id, "id")}</td>
      <td>${isEditing ? policyTextInput(row, "name") : formatCell(row.name, "name")}</td>
      <td>${isEditing ? policyScopeSelect(row.scope_type) : formatCell(row.scope_type, "scope_type")}</td>
      <td>${isEditing ? policyScopeIDInput(row) : formatCell(row.scope_id, "scope_id")}</td>
      <td>${isEditing ? policyTextInput(row, "include_tags", "table-edit-input", "HK, Premium") : formatCell(row.include_tags, "include_tags")}</td>
      <td>${isEditing ? policyTextInput(row, "exclude_tags", "table-edit-input", "Backup") : formatCell(row.exclude_tags, "exclude_tags")}</td>
      <td>${isEditing ? policyTextInput(row, "allowed_virtual_nodes", "table-edit-input-wide", "FluxGate-HK, FluxGate-SG") : formatCell(row.allowed_virtual_nodes, "allowed_virtual_nodes")}</td>
      <td>${isEditing ? policyMaxNodesInput(row) : formatCell(row.max_nodes, "max_nodes")}</td>
      <td>${isEditing ? policyStatusSelect(row.status) : formatCell(row.status, "status")}</td>
      <td class="table-actions policy-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-policy-action="save" data-policy-id="${row.id}">保存</button>
               <button class="table-button ghost-button" type="button" data-policy-action="cancel" data-policy-id="${row.id}">取消</button>`
            : `<button class="table-button" type="button" data-policy-action="edit" data-policy-id="${row.id}">编辑</button>`
        }
      </td>
    </tr>
  `;
}

function policyTextInput(row, field, className = "table-edit-input", placeholder = "") {
  return `<input class="${className}" data-policy-field="${field}" value="${escapeHTML(String(row[field] || ""))}" placeholder="${escapeHTML(placeholder)}" />`;
}

function policyScopeIDInput(row) {
  return `<input class="table-edit-input table-edit-number" data-policy-field="scope_id" type="number" min="1" step="1" value="${escapeHTML(String(row.scope_id || ""))}" />`;
}

function policyMaxNodesInput(row) {
  return `<input class="table-edit-input table-edit-number" data-policy-field="max_nodes" type="number" min="0" step="1" value="${escapeHTML(String(row.max_nodes || 0))}" />`;
}

function policyScopeSelect(value) {
  const current = String(value || "team");
  return `
    <select class="table-edit-input" data-policy-field="scope_type">
      <option value="team" ${current === "team" ? "selected" : ""}>team</option>
      <option value="user" ${current === "user" ? "selected" : ""}>user</option>
      <option value="token" ${current === "token" ? "selected" : ""}>token</option>
    </select>
  `;
}

function policyStatusSelect(value) {
  const current = String(value || "active");
  return `
    <select class="table-edit-input" data-policy-field="status">
      <option value="active" ${current === "active" ? "selected" : ""}>active</option>
      <option value="inactive" ${current === "inactive" ? "selected" : ""}>inactive</option>
    </select>
  `;
}

function renderNodes(rows) {
  appState.nodes = rows || [];
  if (!rows || rows.length === 0) {
    nodesEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  const groups = groupNodesByRegion(rows);
  const selectedGroup = groups.find((group) => group.region === appState.expandedNodeRegion);
  if (appState.expandedNodeRegion && !selectedGroup) {
    appState.expandedNodeRegion = null;
    appState.editingNodeID = null;
    appState.expandedNodeID = null;
    appState.nodeDetail = null;
  }
  nodesEl.innerHTML = appState.expandedNodeRegion && selectedGroup ? renderNodeRegion(selectedGroup, groups) : renderNodeRegions(groups);
}

function groupNodesByRegion(rows) {
  const groups = new Map();
  for (const row of rows || []) {
    const region = normalizedNodeRegion(row);
    if (!groups.has(region)) {
      groups.set(region, []);
    }
    groups.get(region).push(row);
  }
  return Array.from(groups.entries())
    .map(([region, items]) => ({ region, items }))
    .sort((left, right) => {
      if (left.region === "其他") return 1;
      if (right.region === "其他") return -1;
      return left.region.localeCompare(right.region, "zh-CN");
    });
}

function normalizedNodeRegion(row) {
  const region = String(row?.region || "").trim();
  return region || "其他";
}

function renderNodeRegions(groups) {
  const total = groups.reduce((sum, group) => sum + group.items.length, 0);
  return `
    <div class="node-browser" data-node-view="regions">
      <div class="node-browser-header">
        <div>
          <strong>地区聚合</strong>
          <span>${formatCell(total)} 个节点 · ${formatCell(groups.length)} 个地区</span>
        </div>
      </div>
      <div class="node-region-grid">
        ${groups.map((group) => renderNodeRegionCard(group)).join("")}
      </div>
    </div>
  `;
}

function renderNodeRegionCard(group) {
  const activeCount = group.items.filter((node) => node.status === "active").length;
  const protocolCount = new Set(group.items.map((node) => node.protocol).filter(Boolean)).size;
  const sourceCount = new Set(group.items.map((node) => node.source_name || node.source_id).filter(Boolean)).size;
  return `
    <button class="node-region-card" type="button" data-node-region-action="open" data-node-region="${escapeHTML(group.region)}">
      <span class="node-region-name">${escapeHTML(group.region)}</span>
      <strong>${formatCell(group.items.length)} 个节点</strong>
      <span>${formatCell(activeCount)} 可用 · ${formatCell(protocolCount)} 协议 · ${formatCell(sourceCount)} 来源</span>
    </button>
  `;
}

function renderNodeRegion(group, groups) {
  return `
    <div class="node-browser" data-node-view="cards" data-node-region="${escapeHTML(group.region)}">
      <div class="node-browser-header">
        <button class="table-button ghost-button" type="button" data-node-region-action="back">返回地区</button>
        <div>
          <strong>${escapeHTML(group.region)}</strong>
          <span>${formatCell(group.items.length)} 个节点 · 共 ${formatCell(groups.length)} 个地区</span>
        </div>
      </div>
      <div class="node-card-grid">
        ${group.items.map((row) => renderNodeCard(row)).join("")}
      </div>
    </div>
  `;
}

function renderNodeCard(row) {
  const isEditing = appState.editingNodeID === row.id;
  const isExpanded = appState.expandedNodeID === row.id;
  const detail = isExpanded ? appState.nodeDetail || row : null;
  return `
    <article class="node-card ${isExpanded ? "node-card-expanded" : ""}" data-node-id="${row.id}">
      <button class="node-card-main" type="button" data-node-action="detail" data-node-id="${row.id}">
        <span class="node-card-title">${escapeHTML(row.display_name || row.raw_name || `节点 ${row.id}`)}</span>
        <span class="node-card-subtitle">${escapeHTML(row.source_name || "未知来源")} · ${escapeHTML(row.protocol || "--")}</span>
        <span class="node-card-meta">
          <span>${formatStatus(row.status)}</span>
          <span>${labelForColumn("name_mode")}：${formatCell(row.name_mode, "name_mode")}</span>
        </span>
        <span class="node-card-tags">${formatCell(row.tags, "tags")}</span>
      </button>
      ${isEditing ? renderNodeEditForm(row) : ""}
      <div class="table-actions node-actions">
        ${
          isEditing
            ? `<button class="table-button ghost-button" type="button" data-node-action="cancel" data-node-id="${row.id}">取消</button>`
            : `<button class="table-button" type="button" data-node-action="edit" data-node-id="${row.id}">编辑</button>`
        }
        <button class="table-button ghost-button" type="button" data-node-action="reset-name" data-node-id="${row.id}" ${row.name_mode === "auto" ? "disabled" : ""}>恢复自动</button>
      </div>
      ${detail ? renderNodeDetailPanel(detail) : ""}
    </article>
  `;
}

function renderNodeDetailPanel(node) {
  return `
    <div class="node-detail-row">
      <div class="node-detail-panel">
        <div class="detail-grid">
          ${nodeDetailItem("id", node.id)}
          ${nodeDetailItem("source_id", node.source_id)}
          ${nodeDetailItem("source_name", node.source_name)}
          ${nodeDetailItem("raw_name", node.raw_name)}
          ${nodeDetailItem("display_name", node.display_name)}
          ${nodeDetailItem("name_mode", node.name_mode)}
          ${nodeDetailItem("protocol", node.protocol)}
          ${nodeDetailItem("server", node.server)}
          ${nodeDetailItem("server_port", node.server_port)}
          ${nodeDetailItem("region", node.region || "其他")}
          ${nodeDetailItem("tags", node.tags)}
          ${nodeDetailItem("status", node.status)}
          ${nodeDetailItem("last_seen_at", node.last_seen_at)}
          ${nodeDetailItem("last_checked_at", node.last_checked_at)}
          ${nodeDetailItem("last_error", node.last_error)}
          ${nodeDetailItem("created_at", node.created_at)}
          ${nodeDetailItem("updated_at", node.updated_at)}
          ${nodeDetailItem("uri_hash", node.uri_hash)}
        </div>
        <div class="detail-item detail-item-wide">
          <span>${labelForColumn("uri")}</span>
          <code class="detail-code">${escapeHTML(String(node.uri || ""))}</code>
        </div>
      </div>
    </div>
  `;
}

function nodeDetailItem(label, value) {
  return `
    <div class="detail-item">
      <span>${labelForColumn(label)}</span>
      <strong>${formatDetailValue(value, label)}</strong>
    </div>
  `;
}

function formatDetailValue(value, label = "") {
  if (value === null || value === undefined || value === "") {
    return `<span class="cell-muted">--</span>`;
  }
  if (isTimeColumn(label)) {
    return formatTimeCell(value);
  }
  if (Array.isArray(value)) {
    return value.length > 0 ? escapeHTML(value.join(", ")) : `<span class="cell-muted">--</span>`;
  }
  return escapeHTML(String(value));
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
  appState.tokens = rows || [];
  if (!rows || rows.length === 0) {
    tokensEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  tokensEl.innerHTML = `
    <div class="token-card-list">
      ${rows.map((row) => renderTokenCard(row)).join("")}
    </div>
  `;
}

function renderTokenCard(row) {
  return `
    <article class="token-card" data-token-id="${row.id}">
      <div class="token-card-heading">
        <div>
          <span>Token 前缀</span>
          <strong>${formatCell(row.token_prefix, "token_prefix")}</strong>
        </div>
        ${formatCell(row.status, "status")}
      </div>
      <div class="token-card-grid">
        ${tokenCardField("ID", row.id)}
        ${tokenCardField("成员 ID", row.user_id)}
        ${tokenCardField("名称", row.name)}
        ${tokenCardField("到期时间", row.expire_at, "expire_at")}
        ${tokenCardField("额度", row.quota_bytes, "quota_bytes")}
        ${tokenCardField("已用", (row.used_upload_bytes || 0) + (row.used_download_bytes || 0), "used_total")}
      </div>
      <div class="token-card-subscriptions">
        <div class="token-card-section-title">订阅地址</div>
        ${renderTokenSubscriptions(row)}
      </div>
      <div class="token-card-actions">
        <span class="token-action-group" data-token-action-group>
          <input data-token-extend-days="${row.id}" type="number" min="1" value="30" aria-label="续期天数" />
          <button class="table-button" data-token-action="extend" data-token-id="${row.id}">续期</button>
        </span>
        <span class="token-action-group" data-token-action-group>
          <input data-token-quota-mib="${row.id}" type="number" min="1" value="1024" aria-label="追加额度 MiB" />
          <button class="table-button" data-token-action="quota" data-token-id="${row.id}">加额</button>
        </span>
        <button class="table-button" data-token-action="restore" data-token-id="${row.id}">恢复</button>
        <button class="table-button ghost-button" data-token-action="rotate-subscription" data-token-id="${row.id}">重置订阅</button>
        <button class="table-button danger-button" data-token-action="revoke" data-token-id="${row.id}">撤销</button>
      </div>
    </article>
  `;
}

function tokenCardField(label, value, column = "") {
  return `
    <div class="token-card-field">
      <span>${escapeHTML(label)}</span>
      <strong>${formatCell(value, column)}</strong>
    </div>
  `;
}

function renderTokenSubscriptions(row) {
  const subscriptions = row.subscriptions || {};
  const items = [
    ["默认", subscriptions.default || row.subscription || ""],
    ["Clash/Mihomo", subscriptions.clash || ""],
    ["sing-box", subscriptions.sing_box || ""],
  ].filter((item) => item[1]);
  if (!row.subscription_available || items.length === 0) {
    const message = row.subscription_error || "旧 Token 无法反复显示，可重置订阅";
    return `<div class="token-subscription-empty">${escapeHTML(message)}</div>`;
  }
  return `
    <div class="token-subscription-list">
      ${items
        .map(
          ([label, url]) => `
            <div class="token-subscription-item">
              <span>${escapeHTML(label)}</span>
              <code title="${escapeHTML(url)}">${escapeHTML(url)}</code>
              <button class="table-button ghost-button" type="button" data-token-action="copy-subscription" data-token-id="${row.id}" data-token-url="${escapeHTML(url)}">复制</button>
            </div>
          `,
        )
        .join("")}
    </div>
  `;
}

function showTokenSubscriptionResult(result, title) {
  const subscriptions = result.subscriptions || {};
  const defaultSubscription = subscriptions.default || result.subscription || "";
  const clashSubscription = subscriptions.clash || `${defaultSubscription}?target=clash`;
  const singBoxSubscription = subscriptions.sing_box || `${defaultSubscription}?target=sing-box`;
  tokenResultEl.hidden = false;
  tokenResultEl.innerHTML = `
    <strong>${escapeHTML(title)}</strong>
    <div class="subscription-list">
      <span>默认</span>
      <code>${escapeHTML(defaultSubscription)}</code>
      <span>Clash/Mihomo</span>
      <code>${escapeHTML(clashSubscription)}</code>
      <span>sing-box</span>
      <code>${escapeHTML(singBoxSubscription)}</code>
    </div>
  `;
}

function renderTrafficTokens(rows) {
  if (!rows || rows.length === 0) {
    trafficTokensEl.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  trafficTokensEl.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>${labelForColumn("token_id")}</th>
          <th>${labelForColumn("user_id")}</th>
          <th>${labelForColumn("auth_user")}</th>
          <th>${labelForColumn("token_status")}</th>
          <th>${labelForColumn("today_total_bytes")}</th>
          <th>${labelForColumn("month_total_bytes")}</th>
          <th>${labelForColumn("used_upload_bytes")}</th>
          <th>${labelForColumn("used_download_bytes")}</th>
          <th>${labelForColumn("used_total_bytes")}</th>
          <th>${labelForColumn("quota_bytes")}</th>
          <th>${labelForColumn("quota_usage")}</th>
          <th>${labelForColumn("updated_at")}</th>
        </tr>
      </thead>
      <tbody>
        ${rows
          .map(
            (row) => `
              <tr>
                <td>${formatCell(row.token_id, "token_id")}</td>
                <td>${formatCell(row.user_id, "user_id")}</td>
                <td>${formatCell(row.auth_user, "auth_user")}</td>
                <td>${formatCell(row.token_status, "token_status")}</td>
                <td>${formatCell(row.today_total_bytes, "today_total_bytes")}</td>
                <td>${formatCell(row.month_total_bytes, "month_total_bytes")}</td>
                <td>${formatCell(row.used_upload_bytes, "used_upload_bytes")}</td>
                <td>${formatCell(row.used_download_bytes, "used_download_bytes")}</td>
                <td>${formatCell(row.used_total_bytes, "used_total_bytes")}</td>
                <td>${formatCell(row.quota_bytes, "quota_bytes")}</td>
                <td>${formatQuotaUsage(row.used_total_bytes, row.quota_bytes, row.token_status)}</td>
                <td>${formatCell(row.updated_at, "updated_at")}</td>
              </tr>
            `,
          )
          .join("")}
      </tbody>
    </table>
  `;
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
      return formatHourForDisplay(row.hour);
    },
    title: (row, total) => `${formatDateTimeForDisplay(row.hour)} ${formatBytes(total)}`,
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
  if (isTimeColumn(column)) {
    return formatTimeCell(value);
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

function isTimeColumn(column) {
  return column === "hour" || column === "expire_at" || column.endsWith("_at");
}

function formatTimeCell(value) {
  const formatted = formatDateTimeForDisplay(value);
  if (!formatted) {
    return escapeHTML(String(value));
  }
  return `<time datetime="${escapeHTML(String(value))}" title="${escapeHTML(String(value))}">${escapeHTML(formatted)}</time>`;
}

function formatDateTimeForDisplay(value) {
  const date = parseDisplayTime(value);
  if (!date) {
    return "";
  }
  const parts = Object.fromEntries(displayDateTimeFormatter.formatToParts(date).map((part) => [part.type, part.value]));
  return `${parts.year}-${parts.month}-${parts.day} ${parts.hour}:${parts.minute}:${parts.second}`;
}

function formatHourForDisplay(value) {
  const date = parseDisplayTime(value);
  if (!date) {
    return "--";
  }
  const parts = Object.fromEntries(displayHourFormatter.formatToParts(date).map((part) => [part.type, part.value]));
  return `${parts.hour}:00`;
}

function parseDisplayTime(value) {
  if (value instanceof Date) {
    return Number.isNaN(value.getTime()) ? null : value;
  }
  if (typeof value !== "string") {
    return null;
  }
  const text = value.trim();
  if (!text) {
    return null;
  }
  let normalized = text;
  if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}(:\d{2})?$/.test(text)) {
    normalized = `${text.replace(" ", "T")}Z`;
  } else if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2})?(\.\d+)?$/.test(text)) {
    normalized = `${text}Z`;
  }
  const date = new Date(normalized);
  return Number.isNaN(date.getTime()) ? null : date;
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

function formatQuotaUsage(usedBytes, quotaBytes, status = "") {
  const used = Math.max(0, Number(usedBytes || 0));
  const quota = Number(quotaBytes || 0);
  if (quota <= 0) {
    return `
      <div class="quota-meter quota-meter-unlimited" data-quota-usage="unlimited">
        <span class="quota-meter-track"><span class="quota-meter-fill" style="width: 0%"></span></span>
        <strong>不限</strong>
      </div>
    `;
  }
  const ratio = used / quota;
  const percent = Math.round(ratio * 1000) / 10;
  const width = Math.max(2, Math.min(100, percent));
  let className = "quota-meter";
  if (String(status || "").toLowerCase() === "over_quota" || ratio >= 1) {
    className += " quota-meter-danger";
  } else if (ratio >= 0.8) {
    className += " quota-meter-warning";
  }
  return `
    <div class="${className}" data-quota-usage="${escapeHTML(String(percent))}">
      <span class="quota-meter-track"><span class="quota-meter-fill" style="width: ${width}%"></span></span>
      <strong>${escapeHTML(percent.toFixed(1))}%</strong>
    </div>
  `;
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
