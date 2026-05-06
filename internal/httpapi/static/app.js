const statusEl = document.querySelector("#status");
const loginView = document.querySelector("#login-view");
const appView = document.querySelector("#app-view");
const viewSymbolEl = document.querySelector("#view-symbol");
const viewTitleEl = document.querySelector("#view-title");
const viewDescriptionEl = document.querySelector("#view-description");
const viewContextEl = document.querySelector("#view-context");
const viewRailEl = document.querySelector("#view-rail");
const dashboardViewSections = Array.from(document.querySelectorAll("[data-dashboard-view]"));
const dashboardNavButtons = Array.from(document.querySelectorAll("[data-view-nav]"));
const dashboardNavCountEls = Array.from(document.querySelectorAll("[data-view-count]"));
const dashboardJumpButtons = Array.from(document.querySelectorAll("[data-view-jump]"));
const overviewCardCountEls = Array.from(document.querySelectorAll("[data-overview-card-count]"));
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
const dashboardViewSymbols = {
  overview: "概",
  access: "源",
  nodes: "点",
  identity: "身",
  policies: "策",
  traffic: "量",
  ops: "运",
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
nodesEl.addEventListener("input", handleNodeFilterInput);
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
  nodeFilter: "",
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
    nodeFilter: "",
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
  if (viewSymbolEl) viewSymbolEl.textContent = dashboardViewSymbols[nextView] || title.slice(0, 1);
  viewTitleEl.textContent = title;
  viewDescriptionEl.textContent = description;
  updateViewContext();
  updateViewRail();
  updateNavigationCounts();
  updateOverviewCardCounts();
}

function setFormDrawerCollapsed(drawer, collapsed) {
  drawer.classList.toggle("is-collapsed", collapsed);
  const toggle = drawer.querySelector("[data-form-drawer-toggle]");
  if (toggle) {
    toggle.innerHTML = buttonLabel(collapsed ? "+" : "−", collapsed ? "展开" : "收起");
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
    updateNavigationCounts();
    updateOverviewCardCounts();
    statusEl.textContent = "已连接";
  } catch (error) {
    if (error.status === 401) {
      showLogin();
      return;
    }
    statusEl.textContent = "异常";
    metricsEl.innerHTML = emptyState("警", "加载失败", error.message || "请稍后重试");
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
    teamsEl.querySelector(`[data-team-id="${id}"] input[data-team-field="name"]`)?.focus();
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
  const row = button.closest("[data-team-id]");
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
    usersEl.querySelector(`[data-user-id="${id}"] input[data-user-field="name"]`)?.focus();
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
  const row = button.closest("[data-user-id]");
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
    sourcesEl.querySelector(`[data-source-id="${id}"] input[data-source-field="name"]`)?.focus();
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
  const row = button.closest("[data-source-id]");
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
    virtualNodesEl.querySelector(`[data-virtual-node-id="${id}"] input[data-virtual-node-field="name"]`)?.focus();
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
  const row = button.closest("[data-virtual-node-id]");
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
    policiesEl.querySelector(`[data-policy-id="${id}"] input[data-policy-field="name"]`)?.focus();
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
  const row = button.closest("[data-policy-id]");
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
  const filterButton = event.target.closest("button[data-node-filter-action]");
  if (filterButton) {
    appState.nodeFilter = "";
    appState.editingNodeID = null;
    appState.expandedNodeID = null;
    appState.nodeDetail = null;
    renderNodes(appState.nodes);
    nodesEl.querySelector("[data-node-filter]")?.focus();
    return;
  }
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
  if (action === "copy-uri") {
    button.disabled = true;
    try {
      await copyText(button.dataset.nodeUri || "");
      statusEl.textContent = "节点 URI 已复制";
      showCopyFeedback(button, "已复制", "复制 URI");
      button.disabled = false;
    } catch (error) {
      statusEl.textContent = "复制失败";
      button.disabled = false;
    }
    return;
  }
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

function handleNodeFilterInput(event) {
  const input = event.target.closest("[data-node-filter]");
  if (!input) return;
  appState.nodeFilter = input.value;
  appState.editingNodeID = null;
  appState.expandedNodeID = null;
  appState.nodeDetail = null;
  renderNodes(appState.nodes);
  const nextInput = nodesEl.querySelector("[data-node-filter]");
  if (nextInput) {
    const cursorPosition = nextInput.value.length;
    nextInput.focus();
    nextInput.setSelectionRange(cursorPosition, cursorPosition);
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
      showCopyFeedback(button, "已复制", "复制");
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

function showCopyFeedback(button, label, fallbackLabel = "复制") {
  button.dataset.copyLabel =
    button.dataset.copyLabel || button.getAttribute("data-copy-label") || button.textContent.trim() || fallbackLabel;
  button.dataset.copySymbol = button.dataset.copySymbol || button.getAttribute("data-button-symbol") || "";
  button.textContent = label;
  button.classList.add("is-copied");
  button.setAttribute("aria-label", label);
  window.setTimeout(() => {
    if (!button.isConnected) {
      return;
    }
    const restoreLabel = button.dataset.copyLabel || fallbackLabel;
    const restoreSymbol = button.dataset.copySymbol || "";
    if (restoreSymbol) {
      button.innerHTML = buttonLabel(restoreSymbol, restoreLabel);
    } else {
      button.textContent = restoreLabel;
    }
    button.classList.remove("is-copied");
    button.removeAttribute("aria-label");
  }, 1600);
}

async function checkConfig() {
  configCheckEl.disabled = true;
  statusEl.textContent = "检查配置中";
  try {
    const result = await postJSON("/api/sing-box/config/check", {});
    showConfigResult(result.valid ? "检查通过" : "检查失败", result.valid ? "success" : "danger", [
      ["配置 Hash", escapeHTML(String(result.config_hash || "").slice(0, 12)), true],
      ["入站", formatCell(result.inbound_count)],
      ["出口", formatCell(result.outbound_count)],
      ["上游", formatCell(result.upstream_outbound_count)],
      ["用户", formatCell(result.user_count)],
    ]);
    statusEl.textContent = result.valid ? "配置可用" : "配置异常";
  } catch (error) {
    showConfigResult("检查失败", "danger", [["错误", escapeHTML(error.message), true]]);
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
    const restartText = result.restart ? formatRestartResult(result.restart) : result.restart_required ? "需要" : "无需";
    showConfigResult(result.published ? "发布完成" : "发布失败", result.published ? "success" : "danger", [
      ["配置 Hash", escapeHTML(String(result.config_hash || "").slice(0, 12)), true],
      ["上版备份", result.previous_saved ? "已保存" : "无"],
      ["重启", escapeHTML(restartText)],
      ["出口", formatCell(result.outbound_count)],
      ["用户", formatCell(result.user_count)],
    ]);
    statusEl.textContent = result.published && !result.restart_required ? "配置已发布并生效" : result.published ? "配置已发布，需重启 sing-box" : "发布失败";
  } catch (error) {
    showConfigResult("发布失败", "danger", [["错误", escapeHTML(error.message), true]]);
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
    const restartText = result.restart ? formatRestartResult(result.restart) : result.restart_required ? "需要" : "无需";
    showConfigResult(result.rolled_back ? "回滚完成" : "回滚失败", result.rolled_back ? "success" : "danger", [
      ["配置 Hash", escapeHTML(String(result.config_hash || "").slice(0, 12)), true],
      ["重启", escapeHTML(restartText)],
      ["出口", formatCell(result.outbound_count)],
      ["用户", formatCell(result.user_count)],
    ]);
    statusEl.textContent = result.rolled_back && !result.restart_required ? "配置已回滚并生效" : result.rolled_back ? "配置已回滚，需重启 sing-box" : "回滚失败";
  } catch (error) {
    showConfigResult("回滚失败", "danger", [["错误", escapeHTML(error.message), true]]);
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
    showConfigResult(result.success ? "重启已执行" : result.skipped ? "重启未启用" : "重启失败", result.success ? "success" : result.skipped ? "warning" : "danger", [
      ["启用", result.enabled ? "true" : "false"],
      ["已执行", result.executed ? "true" : "false"],
      ["耗时", formatCell(result.duration_ms)],
      ["消息", escapeHTML(result.message || "--"), true],
    ]);
    statusEl.textContent = result.success ? "服务已重启" : result.skipped ? "重启未启用" : "重启失败";
  } catch (error) {
    showConfigResult("重启失败", "danger", [["错误", escapeHTML(error.message), true]]);
    statusEl.textContent = "重启失败";
  } finally {
    configRestartEl.disabled = false;
  }
}

function showConfigResult(title, tone, details) {
  configCheckResultEl.hidden = false;
  configCheckResultEl.innerHTML = `
    <article class="ops-result-card ops-result-${escapeHTML(tone)}">
      <div class="ops-result-heading">
        <span class="ops-result-symbol ops-result-symbol-${escapeHTML(tone)}" aria-hidden="true">${opsResultSymbol(tone)}</span>
        <div class="ops-result-title">
          <strong>${escapeHTML(title)}</strong>
          <span class="ops-result-chip ops-result-chip-${escapeHTML(tone)}" data-ops-result-chip>${opsResultToneLabel(tone)}</span>
        </div>
        <span class="ops-result-time">${formatDateTimeForDisplay(new Date().toISOString())}</span>
      </div>
      <div class="ops-result-grid">
        ${details
          .map(([label, value, isCode]) => {
            const content = isCode ? `<code>${value}</code>` : `<strong>${value}</strong>`;
            return `
              <div class="ops-result-field">
                <span>${escapeHTML(label)}</span>
                ${content}
              </div>
            `;
          })
          .join("")}
      </div>
    </article>
  `;
}

function opsResultSymbol(tone) {
  if (tone === "success") return "通";
  if (tone === "warning") return "待";
  if (tone === "danger") return "警";
  return "运";
}

function opsResultToneLabel(tone) {
  if (tone === "success") return "已通过";
  if (tone === "warning") return "需处理";
  if (tone === "danger") return "异常";
  return "结果";
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
    ["teams", "团", "团队", data.teams],
    ["users", "员", "用户", data.users],
    ["tokens", "令", "Token", data.tokens],
    ["sources", "源", "来源", data.sources],
    ["nodes", "点", "节点", data.nodes],
    ["virtualNodes", "网", "虚拟节点", data.virtual_nodes],
    ["policies", "策", "策略", data.policies],
  ];
  metricsEl.innerHTML = items
    .map(
      ([key, symbol, label, value]) => `
        <div class="metric" data-metric-key="${escapeHTML(key)}">
          <span class="metric-heading">
            <span class="metric-symbol">${escapeHTML(symbol)}</span>
            <span class="metric-label">${escapeHTML(label)}</span>
          </span>
          <strong class="metric-value">${formatPlainNumber(value ?? 0)}</strong>
        </div>
      `,
    )
    .join("");
}

function overviewReadinessChecks(data) {
  return [
    ["接入来源", Number(data.sources || 0) > 0, `${data.sources || 0} 个来源`, "access", "源"],
    ["节点池", Number(data.nodes || 0) > 0, `${data.nodes || 0} 个节点`, "nodes", "点"],
    ["虚拟网关", Number(data.virtual_nodes || 0) > 0, `${data.virtual_nodes || 0} 个入口`, "nodes", "网"],
    ["团队 Token", Number(data.tokens || 0) > 0, `${data.tokens || 0} 个 Token`, "identity", "身"],
    ["访问策略", Number(data.policies || 0) > 0, `${data.policies || 0} 条策略`, "policies", "策"],
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
          ([label, ready, summary, view, symbol], index) => `
            <button class="readiness-item ${ready ? "is-ready" : ""}" type="button" data-overview-jump="${view}">
              <span class="readiness-symbol-wrap" aria-hidden="true">
                <span class="readiness-index">${String(index + 1).padStart(2, "0")}</span>
                <span class="readiness-symbol">${escapeHTML(symbol)}</span>
                <span class="readiness-dot"></span>
              </span>
              <span class="readiness-body">
                <span class="readiness-title-row">
                  <strong>${escapeHTML(label)}</strong>
                  <span class="readiness-state">${ready ? "就绪" : "待补"}</span>
                </span>
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
    <div class="next-step-card ${firstMissing ? "is-warning" : "is-ready"}" data-overview-next-step-card>
      <span class="next-step-symbol">${escapeHTML(dashboardViewSymbols[next.view] || "→")}</span>
      <span class="next-step-body">
        <strong>${escapeHTML(next.title)}</strong>
        <small>${escapeHTML(next.description)}</small>
      </span>
      <button class="primary-link-button next-step-action" type="button" data-overview-jump="${next.view}">
        ${buttonLabel("→", next.action)}
      </button>
    </div>
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
  const metric = viewRailMetricForTarget(target, expand);
  const symbol = viewRailSymbolForTarget(label, target, expand);
  return `
    <button class="view-rail-button" type="button" data-view-rail-target="${escapeHTML(target)}" data-view-rail-expand="${expand ? "true" : "false"}">
      <span class="view-rail-symbol" aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="view-rail-label">${escapeHTML(label)}</span>
      <span class="view-rail-count" data-view-rail-count>${escapeHTML(metric)}</span>
    </button>
  `;
}

function viewRailSymbolForTarget(label, target, expand) {
  const symbols = {
    metrics: "概",
    "overview-readiness": "闭",
    "overview-next-step": "步",
    sources: "源",
    "source-form": "加",
    "node-import-form": "导",
    nodes: "点",
    "virtual-nodes": "网",
    "virtual-node-form": "建",
    tokens: "钥",
    teams: "团",
    users: "员",
    "token-form": "钥",
    "team-form": "团",
    "user-form": "员",
    policies: "策",
    "policy-form": "建",
    "traffic-tokens": "量",
    "traffic-hourly": "时",
    "traffic-daily": "日",
    "traffic-outbounds": "出",
    "config-check-result": "运",
  };
  if (symbols[target]) return symbols[target];
  if (expand) return "+";
  return label.slice(0, 1) || "项";
}

function viewRailMetricForTarget(target, expand) {
  if (expand) return "+";
  const state = appState || {};
  const nodes = state.nodes || [];
  const virtualNodes = state.virtualNodes || [];
  const tokens = state.tokens || [];
  const readiness = overviewReadinessChecks(state.overview || {});
  const readyCount = readiness.filter(([, ready]) => ready).length;
  const activeNodes = countBy(nodes, (row) => row.status === "active");
  const activeVirtualNodes = countBy(virtualNodes, (row) => row.status === "active");
  const activeTokens = countBy(tokens, (row) => row.status === "active");
  const configReady = activeNodes > 0 && activeVirtualNodes > 0 && activeTokens > 0;
  const values = {
    metrics: `${formatPlainNumber(state.overview?.nodes || nodes.length)} 节点`,
    "overview-readiness": `${readyCount}/${readiness.length}`,
    "overview-next-step": readyCount === readiness.length ? "可测" : "待补",
    sources: formatPlainNumber((state.sources || []).length),
    nodes: `${formatPlainNumber(groupNodesByRegion(nodes).length)} 地区`,
    "virtual-nodes": formatPlainNumber(virtualNodes.length),
    teams: formatPlainNumber((state.teams || []).length),
    users: formatPlainNumber((state.users || []).length),
    tokens: formatPlainNumber(tokens.length),
    policies: formatPlainNumber((state.policies || []).length),
    "traffic-tokens": formatPlainNumber((state.trafficTokens || []).length),
    "traffic-hourly": formatPlainNumber((state.trafficHourly || []).length),
    "traffic-daily": formatPlainNumber((state.trafficDaily || []).length),
    "traffic-outbounds": formatPlainNumber((state.trafficOutbounds || []).length),
    "config-check-result": configReady ? "就绪" : "待检",
  };
  return values[target] || "0";
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

function updateNavigationCounts() {
  const counts = navigationCounts();
  dashboardNavCountEls.forEach((element) => {
    const view = element.dataset.viewCount || "";
    const value = counts[view] || "0";
    element.textContent = value;
    element.classList.toggle("is-empty", value === "0" || value === "0/5" || value === "待检");
    element.setAttribute("title", `${dashboardViewMeta[view]?.[0] || view}：${value}`);
  });
}

function updateOverviewCardCounts() {
  const counts = overviewCardCounts();
  overviewCardCountEls.forEach((element) => {
    const key = element.dataset.overviewCardCount || "";
    const value = counts[key] || "0";
    element.textContent = value;
    element.classList.toggle("is-empty", value === "0" || value === "待检");
    element.setAttribute("title", `${overviewCardCountLabel(key)}：${value}`);
  });
}

function overviewCardCounts() {
  const sources = appState.sources || [];
  const nodes = appState.nodes || [];
  const virtualNodes = appState.virtualNodes || [];
  const tokens = appState.tokens || [];
  const activeNodes = countBy(nodes, (row) => row.status === "active");
  const activeVirtualNodes = countBy(virtualNodes, (row) => row.status === "active");
  const activeTokens = countBy(tokens, (row) => row.status === "active");
  const configReady = activeNodes > 0 && activeVirtualNodes > 0 && activeTokens > 0;
  return {
    access: `${formatPlainNumber(sources.length)} 来源`,
    nodes: `${formatPlainNumber(groupNodesByRegion(nodes).length)} 地区`,
    identity: `${formatPlainNumber(tokens.length)} Token`,
    ops: configReady ? "就绪" : "待检",
  };
}

function overviewCardCountLabel(key) {
  const labels = {
    access: "接入来源",
    nodes: "节点池",
    identity: "分发订阅",
    ops: "发布网关",
  };
  return labels[key] || key;
}

function navigationCounts() {
  const overview = appState.overview || {};
  const sources = appState.sources || [];
  const nodes = appState.nodes || [];
  const virtualNodes = appState.virtualNodes || [];
  const policies = appState.policies || [];
  const tokens = appState.tokens || [];
  const trafficHourly = appState.trafficHourly || [];
  const trafficDaily = appState.trafficDaily || [];
  const trafficOutbounds = appState.trafficOutbounds || [];
  const trafficTokens = appState.trafficTokens || [];
  const readiness = overviewReadinessChecks(overview);
  const readyCount = readiness.filter(([, ready]) => ready).length;
  const activeNodes = countBy(nodes, (row) => row.status === "active");
  const activeVirtualNodes = countBy(virtualNodes, (row) => row.status === "active");
  const activeTokens = countBy(tokens, (row) => row.status === "active");
  const configReady = activeNodes > 0 && activeVirtualNodes > 0 && activeTokens > 0;
  return {
    overview: `${readyCount}/${readiness.length}`,
    access: formatPlainNumber(sources.length),
    nodes: formatPlainNumber(nodes.length),
    identity: formatPlainNumber(tokens.length),
    policies: formatPlainNumber(policies.length),
    traffic: formatPlainNumber(trafficTokens.length || trafficOutbounds.length || trafficDaily.length || trafficHourly.length),
    ops: configReady ? "就绪" : "待检",
  };
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

function emptyState(symbol, title, hint = "") {
  return `
    <div class="empty empty-state" data-empty-state>
      <span class="empty-state-symbol" data-empty-state-symbol aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="empty-state-copy">
        <strong class="empty-state-title">${escapeHTML(title)}</strong>
        ${hint ? `<span class="empty-state-hint">${escapeHTML(hint)}</span>` : ""}
      </span>
    </div>
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
    teamsEl.innerHTML = emptyState("团", "暂无团队", "创建团队后再为成员签发订阅 Token");
    return;
  }
  teamsEl.innerHTML = `
    <div class="identity-card-list">
      ${rows.map((row) => renderTeamCard(row)).join("")}
    </div>
  `;
}

function renderTeamCard(row) {
  const isEditing = appState.editingTeamID === row.id;
  const statusSummary = identityStatusSummary(row.status);
  return `
    <article class="identity-card ${isEditing ? "identity-card-editing" : ""}" data-team-id="${row.id}">
      <div class="identity-card-heading">
        <div>
          <span>${labelForColumn("id")} ${formatCell(row.id, "id")}</span>
          <strong>${isEditing ? teamTextInput(row, "name") : formatCell(row.name, "name")}</strong>
        </div>
        <span class="identity-card-badge">${formatCell(row.status, "status")}</span>
      </div>
      <div class="identity-card-summary" aria-label="团队摘要">
        ${identitySummaryChip("团队", "info")}
        ${identitySummaryChip(statusSummary.label, statusSummary.tone)}
        ${identitySummaryChip(teamMemberSummary(row.id))}
        ${identitySummaryChip(row.description ? "有备注" : "无备注")}
      </div>
      <div class="identity-card-grid">
        ${identityCardField("description", isEditing ? teamTextInput(row, "description", "table-edit-input-wide") : formatCell(row.description, "description"), "identity-card-field-wide")}
        ${identityCardField("status", isEditing ? teamStatusSelect(row.status) : formatCell(row.status, "status"))}
      </div>
      <div class="table-actions identity-card-actions compact-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-team-action="save" data-team-id="${row.id}">${buttonLabel("✓", "保存")}</button>
               <button class="table-button ghost-button" type="button" data-team-action="cancel" data-team-id="${row.id}">${buttonLabel("×", "取消")}</button>`
            : `<button class="table-button" type="button" data-team-action="edit" data-team-id="${row.id}">${buttonLabel("✎", "编辑")}</button>`
        }
      </div>
    </article>
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
    usersEl.innerHTML = emptyState("员", "暂无成员", "添加成员后可以独立控制 Token 有效期和额度");
    return;
  }
  usersEl.innerHTML = `
    <div class="identity-card-list">
      ${rows.map((row) => renderUserCard(row)).join("")}
    </div>
  `;
}

function renderUserCard(row) {
  const isEditing = appState.editingUserID === row.id;
  const teamLabel = teamLabelForID(row.team_id);
  const statusSummary = identityStatusSummary(row.status);
  return `
    <article class="identity-card ${isEditing ? "identity-card-editing" : ""}" data-user-id="${row.id}">
      <div class="identity-card-heading">
        <div>
          <span>${labelForColumn("id")} ${formatCell(row.id, "id")}</span>
          <strong>${isEditing ? userTextInput(row, "name") : formatCell(row.name, "name")}</strong>
        </div>
        <span class="identity-card-badge">${escapeHTML(teamLabel)}</span>
      </div>
      <div class="identity-card-summary" aria-label="成员摘要">
        ${identitySummaryChip(escapeHTML(teamLabel), "info")}
        ${identitySummaryChip(statusSummary.label, statusSummary.tone)}
        ${identitySummaryChip(row.email ? "有邮箱" : "无邮箱")}
        ${identitySummaryChip(row.remark ? "有备注" : "无备注")}
      </div>
      <div class="identity-card-grid">
        ${identityCardField("team_id", isEditing ? userTeamSelectInput(row.team_id) : formatCell(row.team_id, "team_id"))}
        ${identityCardField("email", isEditing ? userTextInput(row, "email", "table-edit-input-wide") : formatCell(row.email, "email"), "identity-card-field-wide")}
        ${identityCardField("remark", isEditing ? userTextInput(row, "remark", "table-edit-input-wide") : formatCell(row.remark, "remark"), "identity-card-field-wide")}
        ${identityCardField("status", isEditing ? userStatusSelect(row.status) : formatCell(row.status, "status"))}
      </div>
      <div class="table-actions identity-card-actions compact-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-user-action="save" data-user-id="${row.id}">${buttonLabel("✓", "保存")}</button>
               <button class="table-button ghost-button" type="button" data-user-action="cancel" data-user-id="${row.id}">${buttonLabel("×", "取消")}</button>`
            : `<button class="table-button" type="button" data-user-action="edit" data-user-id="${row.id}">${buttonLabel("✎", "编辑")}</button>`
        }
      </div>
    </article>
  `;
}

function identitySummaryChip(content, tone = "") {
  return `<span class="identity-card-summary-chip ${tone ? `identity-card-summary-chip-${tone}` : ""}" data-identity-summary-chip>${content}</span>`;
}

function identityStatusSummary(value) {
  const status = String(value || "active");
  return {
    label: escapeHTML(status),
    tone: status === "active" ? "success" : "muted",
  };
}

function teamMemberSummary(teamID) {
  const memberCount = countBy(appState.users || [], (user) => Number(user.team_id) === Number(teamID));
  return `${formatPlainNumber(memberCount)} 成员`;
}

function identityCardField(label, value, className = "") {
  return `
    <div class="identity-card-field ${className}">
      <span>${labelForColumn(label)}</span>
      <strong>${value}</strong>
    </div>
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

function teamLabelForID(value) {
  if (!value) return "未绑定团队";
  const team = (appState.teams || []).find((item) => String(item.id) === String(value));
  return team ? team.name : `团队 #${value}`;
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
    sourcesEl.innerHTML = emptyState("源", "暂无来源", "添加订阅或手动来源后会在这里统一管理");
    return;
  }
  sourcesEl.innerHTML = `
    <div class="source-card-list">
      ${rows.map((row) => renderSourceCard(row)).join("")}
    </div>
  `;
}

function renderSourceCard(row) {
  const isEditing = appState.editingSourceID === row.id;
  const syncSummary = sourceSyncSummary(row);
  return `
    <article class="source-card ${isEditing ? "source-card-editing" : ""}" data-source-id="${row.id}">
      <div class="source-card-heading">
        <div>
          <span>${labelForColumn("id")} ${formatCell(row.id, "id")}</span>
          <strong>${isEditing ? sourceTextInput(row, "name") : formatCell(row.name, "name")}</strong>
        </div>
        <span class="source-type-badge">${isEditing ? sourceTypeSelect(row.type) : formatCell(row.type, "type")}</span>
      </div>
      <div class="source-card-summary" aria-label="来源摘要">
        ${sourceSummaryChip(formatCell(row.type, "type"), "info")}
        ${sourceSummaryChip(sourcePrefixSummary(row.display_prefix))}
        ${sourceSummaryChip(sourceRefreshSummary(row.refresh_interval_minutes))}
        ${sourceSummaryChip(syncSummary.label, syncSummary.tone)}
      </div>
      <div class="source-card-grid">
        ${sourceCardField("url", isEditing ? sourceTextInput(row, "url", "table-edit-input-wide") : formatCell(row.url, "url"))}
        ${sourceCardField("display_prefix", isEditing ? sourceTextInput(row, "display_prefix", "table-edit-input", "留空自动") : formatCell(row.display_prefix, "display_prefix"))}
        ${sourceCardField("default_tags", isEditing ? sourceTextInput(row, "default_tags", "table-edit-input", "HK, Premium") : formatCell(row.default_tags, "default_tags"))}
        ${sourceCardField("refresh_interval_minutes", isEditing ? sourceNumberInput(row, "refresh_interval_minutes") : formatCell(row.refresh_interval_minutes, "refresh_interval_minutes"))}
        ${sourceCardField("last_sync_at", formatCell(row.last_sync_at, "last_sync_at"))}
        ${sourceCardField("last_error", formatCell(row.last_error, "last_error"), row.last_error ? "source-card-field-warning" : "")}
      </div>
      <div class="table-actions source-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-source-action="save" data-source-id="${row.id}">${buttonLabel("✓", "保存")}</button>
               <button class="table-button ghost-button" type="button" data-source-action="cancel" data-source-id="${row.id}">${buttonLabel("×", "取消")}</button>`
            : `<button class="table-button" type="button" data-source-action="edit" data-source-id="${row.id}">${buttonLabel("✎", "编辑")}</button>`
        }
        <button class="table-button" type="button" data-source-action="refresh" data-source-id="${row.id}">${buttonLabel("↻", "刷新")}</button>
        <button class="table-button ghost-button" type="button" data-source-action="regenerate" data-source-id="${row.id}">${buttonLabel("↺", "同步命名")}</button>
      </div>
    </article>
  `;
}

function sourceSummaryChip(content, tone = "") {
  return `<span class="source-card-summary-chip ${tone ? `source-card-summary-chip-${tone}` : ""}" data-source-summary-chip>${content}</span>`;
}

function sourcePrefixSummary(value) {
  const prefix = String(value || "").trim();
  return prefix ? `前缀 ${escapeHTML(prefix)}` : "自动前缀";
}

function sourceRefreshSummary(value) {
  const minutes = Number(value || 0);
  if (!Number.isFinite(minutes) || minutes <= 0) {
    return "手动刷新";
  }
  return `${formatPlainNumber(minutes)} 分钟刷新`;
}

function sourceSyncSummary(row) {
  if (row?.last_error) {
    return { label: "同步异常", tone: "warning" };
  }
  if (row?.last_sync_at) {
    return { label: "已同步", tone: "success" };
  }
  return { label: "待同步", tone: "muted" };
}

function sourceCardField(label, value, className = "") {
  return `
    <div class="source-card-field ${className}">
      <span>${labelForColumn(label)}</span>
      <strong>${value}</strong>
    </div>
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
    virtualNodesEl.innerHTML = emptyState("网", "暂无虚拟网关", "创建网关入口后才能生成可用的 sing-box 入站");
    return;
  }
  virtualNodesEl.innerHTML = `
    <div class="virtual-node-card-list">
      ${rows.map((row) => renderVirtualNodeCard(row)).join("")}
    </div>
  `;
}

function renderVirtualNodeCard(row) {
  const isEditing = appState.editingVirtualNodeID === row.id;
  const listenSummary = `${row.listen_protocol || "vless"}:${row.listen_port || "--"}`;
  const statusSummary = virtualNodeStatusSummary(row.status);
  return `
    <article class="virtual-node-card ${isEditing ? "virtual-node-card-editing" : ""}" data-virtual-node-id="${row.id}">
      <div class="virtual-node-card-heading">
        <div>
          <span>${labelForColumn("id")} ${formatCell(row.id, "id")}</span>
          <strong>${isEditing ? virtualNodeTextInput(row, "name") : formatCell(row.name, "name")}</strong>
        </div>
        <span class="virtual-node-listen-badge">${escapeHTML(listenSummary)}</span>
      </div>
      <div class="virtual-node-card-summary" aria-label="虚拟网关摘要">
        ${virtualNodeSummaryChip(`监听 ${escapeHTML(listenSummary)}`, "info")}
        ${virtualNodeSummaryChip(virtualNodeStrategySummary(row.strategy))}
        ${virtualNodeSummaryChip(statusSummary.label, statusSummary.tone)}
        ${virtualNodeSummaryChip(virtualNodeTagSummary(row.tag_selector))}
      </div>
      <div class="virtual-node-card-grid">
        ${virtualNodeCardField("listen_protocol", isEditing ? virtualNodeProtocolSelect(row.listen_protocol) : formatCell(row.listen_protocol, "listen_protocol"))}
        ${virtualNodeCardField("listen_port", isEditing ? virtualNodeNumberInput(row, "listen_port") : formatCell(row.listen_port, "listen_port"))}
        ${virtualNodeCardField("strategy", isEditing ? virtualNodeStrategySelect(row.strategy) : formatCell(row.strategy, "strategy"))}
        ${virtualNodeCardField("status", isEditing ? virtualNodeStatusSelect(row.status) : formatCell(row.status, "status"))}
        ${virtualNodeCardField("tag_selector", isEditing ? virtualNodeTextInput(row, "tag_selector", "table-edit-input-wide", '{"include":["HK"]}') : formatCell(row.tag_selector, "tag_selector"), "virtual-node-card-field-wide")}
      </div>
      <div class="table-actions virtual-node-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-virtual-node-action="save" data-virtual-node-id="${row.id}">${buttonLabel("✓", "保存")}</button>
               <button class="table-button ghost-button" type="button" data-virtual-node-action="cancel" data-virtual-node-id="${row.id}">${buttonLabel("×", "取消")}</button>`
            : `<button class="table-button" type="button" data-virtual-node-action="edit" data-virtual-node-id="${row.id}">${buttonLabel("✎", "编辑")}</button>`
        }
      </div>
    </article>
  `;
}

function virtualNodeSummaryChip(content, tone = "") {
  return `<span class="virtual-node-card-summary-chip ${tone ? `virtual-node-card-summary-chip-${tone}` : ""}" data-virtual-node-summary-chip>${content}</span>`;
}

function virtualNodeStrategySummary(value) {
  return `策略 ${escapeHTML(String(value || "selector"))}`;
}

function virtualNodeStatusSummary(value) {
  const status = String(value || "active");
  return {
    label: escapeHTML(status),
    tone: status === "active" ? "success" : "muted",
  };
}

function virtualNodeTagSummary(value) {
  return String(value || "").trim() ? "标签筛选" : "全部节点";
}

function virtualNodeCardField(label, value, className = "") {
  return `
    <div class="virtual-node-card-field ${className}">
      <span>${labelForColumn(label)}</span>
      <strong>${value}</strong>
    </div>
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
    policiesEl.innerHTML = emptyState("策", "暂无策略", "配置策略后可以限制团队可见节点和网关范围");
    return;
  }
  policiesEl.innerHTML = `
    <div class="policy-card-list">
      ${rows.map((row) => renderPolicyCard(row)).join("")}
    </div>
  `;
}

function renderPolicyCard(row) {
  const isEditing = appState.editingPolicyID === row.id;
  const scopeSummary = `${row.scope_type || "team"} #${row.scope_id || "--"}`;
  const statusSummary = policyStatusSummary(row.status);
  return `
    <article class="policy-card ${isEditing ? "policy-card-editing" : ""}" data-policy-id="${row.id}">
      <div class="policy-card-heading">
        <div>
          <span>${labelForColumn("id")} ${formatCell(row.id, "id")}</span>
          <strong>${isEditing ? policyTextInput(row, "name") : formatCell(row.name, "name")}</strong>
        </div>
        <span class="policy-scope-badge">${escapeHTML(scopeSummary)}</span>
      </div>
      <div class="policy-card-summary" aria-label="策略摘要">
        ${policySummaryChip(`作用域 ${escapeHTML(scopeSummary)}`, "info")}
        ${policySummaryChip(policyMaxNodesSummary(row.max_nodes))}
        ${policySummaryChip(statusSummary.label, statusSummary.tone)}
        ${policySummaryChip(policyVirtualNodesSummary(row.allowed_virtual_nodes))}
      </div>
      <div class="policy-card-grid">
        ${policyCardField("scope_type", isEditing ? policyScopeSelect(row.scope_type) : formatCell(row.scope_type, "scope_type"))}
        ${policyCardField("scope_id", isEditing ? policyScopeIDInput(row) : formatCell(row.scope_id, "scope_id"))}
        ${policyCardField("max_nodes", isEditing ? policyMaxNodesInput(row) : formatCell(row.max_nodes, "max_nodes"))}
        ${policyCardField("status", isEditing ? policyStatusSelect(row.status) : formatCell(row.status, "status"))}
        ${policyCardField("include_tags", isEditing ? policyTextInput(row, "include_tags", "table-edit-input", "HK, Premium") : formatCell(row.include_tags, "include_tags"))}
        ${policyCardField("exclude_tags", isEditing ? policyTextInput(row, "exclude_tags", "table-edit-input", "Backup") : formatCell(row.exclude_tags, "exclude_tags"))}
        ${policyCardField("allowed_virtual_nodes", isEditing ? policyTextInput(row, "allowed_virtual_nodes", "table-edit-input-wide", "FluxGate-HK, FluxGate-SG") : formatCell(row.allowed_virtual_nodes, "allowed_virtual_nodes"), "policy-card-field-wide")}
      </div>
      <div class="table-actions policy-actions">
        ${
          isEditing
            ? `<button class="table-button" type="button" data-policy-action="save" data-policy-id="${row.id}">${buttonLabel("✓", "保存")}</button>
               <button class="table-button ghost-button" type="button" data-policy-action="cancel" data-policy-id="${row.id}">${buttonLabel("×", "取消")}</button>`
            : `<button class="table-button" type="button" data-policy-action="edit" data-policy-id="${row.id}">${buttonLabel("✎", "编辑")}</button>`
        }
      </div>
    </article>
  `;
}

function policySummaryChip(content, tone = "") {
  return `<span class="policy-card-summary-chip ${tone ? `policy-card-summary-chip-${tone}` : ""}" data-policy-summary-chip>${content}</span>`;
}

function policyMaxNodesSummary(value) {
  const maxNodes = Number(value || 0);
  if (!Number.isFinite(maxNodes) || maxNodes <= 0) {
    return "不限节点";
  }
  return `最多 ${formatPlainNumber(maxNodes)} 节点`;
}

function policyStatusSummary(value) {
  const status = String(value || "active");
  return {
    label: escapeHTML(status),
    tone: status === "active" ? "success" : "muted",
  };
}

function policyVirtualNodesSummary(value) {
  return String(value || "").trim() ? "指定网关" : "全部网关";
}

function policyCardField(label, value, className = "") {
  return `
    <div class="policy-card-field ${className}">
      <span>${labelForColumn(label)}</span>
      <strong>${value}</strong>
    </div>
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
    nodesEl.innerHTML = emptyState("点", "暂无节点", "导入上游订阅后会按地区自动聚合节点");
    return;
  }
  const filteredRows = filterNodesByQuery(rows, appState.nodeFilter);
  const groups = groupNodesByRegion(filteredRows);
  const selectedGroup = groups.find((group) => group.region === appState.expandedNodeRegion);
  if (appState.expandedNodeRegion && !selectedGroup) {
    appState.expandedNodeRegion = null;
    appState.editingNodeID = null;
    appState.expandedNodeID = null;
    appState.nodeDetail = null;
  }
  const totalGroups = groupNodesByRegion(rows);
  nodesEl.innerHTML =
    appState.expandedNodeRegion && selectedGroup
      ? renderNodeRegion(selectedGroup, groups, rows.length, filteredRows.length, totalGroups.length)
      : renderNodeRegions(groups, rows.length, filteredRows.length);
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

function filterNodesByQuery(rows, query) {
  const normalizedQuery = normalizeSearchText(query);
  if (!normalizedQuery) return rows || [];
  return (rows || []).filter((row) => nodeSearchText(row).includes(normalizedQuery));
}

function nodeSearchText(row) {
  const fields = [
    row.id,
    normalizedNodeRegion(row),
    row.display_name,
    row.raw_name,
    row.source_name,
    row.protocol,
    row.server,
    row.server_port,
    row.status,
    row.name_mode,
    Array.isArray(row.tags) ? row.tags.join(" ") : row.tags,
  ];
  return normalizeSearchText(fields.filter((field) => field !== null && field !== undefined).join(" "));
}

function normalizeSearchText(value) {
  return String(value || "")
    .trim()
    .toLocaleLowerCase("zh-CN");
}

function renderNodeRegions(groups, total, matched) {
  return `
    <div class="node-browser" data-node-view="regions">
      <div class="node-browser-header">
        <div>
          <strong>地区聚合</strong>
          <span>${formatCell(matched)} / ${formatCell(total)} 个节点 · ${formatCell(groups.length)} 个地区</span>
        </div>
      </div>
      ${renderNodeFilterBar(total, matched, groups.length)}
      ${
        groups.length > 0
          ? `<div class="node-region-grid">${groups.map((group) => renderNodeRegionCard(group)).join("")}</div>`
          : emptyState("搜", "没有匹配的节点", "换一个地区、协议、来源、标签或服务器关键词试试")
      }
    </div>
  `;
}

function renderNodeRegionCard(group) {
  const activeCount = group.items.filter((node) => node.status === "active").length;
  const protocolCount = new Set(group.items.map((node) => node.protocol).filter(Boolean)).size;
  const sourceCount = new Set(group.items.map((node) => node.source_name || node.source_id).filter(Boolean)).size;
  return `
    <button class="node-region-card" type="button" data-node-region-action="open" data-node-region="${escapeHTML(group.region)}">
      <span class="node-region-heading">
        <span class="node-region-symbol" aria-hidden="true">${escapeHTML(regionSymbolForRegion(group.region))}</span>
        <span class="node-region-copy">
          <span class="node-region-name">${escapeHTML(group.region)}</span>
          <strong>${formatCell(group.items.length)} 个节点</strong>
        </span>
        <span class="node-region-badge">${formatCell(activeCount)} 可用</span>
      </span>
      <span class="node-region-stats">
        <span data-node-region-stat>${formatCell(protocolCount)} 协议</span>
        <span data-node-region-stat>${formatCell(sourceCount)} 来源</span>
      </span>
    </button>
  `;
}

function regionSymbolForRegion(region) {
  const text = String(region || "其他").trim() || "其他";
  const flag = text.match(/[\u{1f1e6}-\u{1f1ff}]{2}/u);
  if (flag) return flag[0];
  const regionName = text.includes("|") ? text.split("|").pop().trim() : text;
  return regionName.slice(0, 1) || "区";
}

function renderNodeRegion(group, groups, total, matched, totalRegionCount) {
  return `
    <div class="node-browser" data-node-view="cards" data-node-region="${escapeHTML(group.region)}">
      <div class="node-browser-header">
        <button class="table-button ghost-button" type="button" data-node-region-action="back">${buttonLabel("←", "返回地区")}</button>
        <div>
          <strong>${escapeHTML(group.region)}</strong>
          <span>${formatCell(group.items.length)} 个节点 · 筛选 ${formatCell(matched)} / ${formatCell(total)} · 共 ${formatCell(totalRegionCount)} 个地区</span>
        </div>
      </div>
      ${renderNodeRegionSummary(group)}
      ${renderNodeFilterBar(total, matched, groups.length)}
      <div class="node-card-grid">
        ${group.items.map((row) => renderNodeCard(row)).join("")}
      </div>
    </div>
  `;
}

function renderNodeRegionSummary(group) {
  const activeCount = group.items.filter((node) => node.status === "active").length;
  const protocolCount = new Set(group.items.map((node) => node.protocol).filter(Boolean)).size;
  const sourceCount = new Set(group.items.map((node) => node.source_name || node.source_id).filter(Boolean)).size;
  return `
    <div class="node-region-summary" data-node-region-summary>
      <span class="node-region-summary-symbol" aria-hidden="true">${escapeHTML(regionSymbolForRegion(group.region))}</span>
      <div class="node-region-summary-copy">
        <strong>${escapeHTML(group.region)}</strong>
        <span>本地区节点概况</span>
      </div>
      <div class="node-region-summary-chips">
        <span data-node-region-summary-chip>${formatCell(group.items.length)} 节点</span>
        <span data-node-region-summary-chip>${formatCell(activeCount)} 可用</span>
        <span data-node-region-summary-chip>${formatCell(protocolCount)} 协议</span>
        <span data-node-region-summary-chip>${formatCell(sourceCount)} 来源</span>
      </div>
    </div>
  `;
}

function renderNodeFilterBar(total, matched, regionCount) {
  const query = String(appState.nodeFilter || "");
  const hasQuery = normalizeSearchText(query) !== "";
  return `
    <div class="node-filter-bar">
      <label>
        <span>搜索节点</span>
        <input data-node-filter value="${escapeHTML(query)}" placeholder="地区、节点名、协议、来源、标签或服务器" autocomplete="off" />
      </label>
      <div class="node-filter-summary">
        <strong>${formatCell(matched)}</strong>
        <span>/ ${formatCell(total)} 节点 · ${formatCell(regionCount)} 地区</span>
      </div>
      <button class="table-button ghost-button" type="button" data-node-filter-action="clear" ${hasQuery ? "" : "disabled"}>${buttonLabel("×", "清空")}</button>
    </div>
  `;
}

function renderNodeCard(row) {
  const isEditing = appState.editingNodeID === row.id;
  const isExpanded = appState.expandedNodeID === row.id;
  const detail = isExpanded ? appState.nodeDetail || row : null;
  const endpoint = nodeEndpointLabel(row);
  return `
    <article class="node-card ${isExpanded ? "node-card-expanded" : ""}" data-node-id="${row.id}">
      <button class="node-card-main" type="button" data-node-action="detail" data-node-id="${row.id}">
        <span class="node-card-heading">
          <span class="node-card-protocol-symbol" aria-hidden="true">${escapeHTML(nodeProtocolSymbol(row.protocol))}</span>
          <span class="node-card-copy">
            <span class="node-card-title">${escapeHTML(row.display_name || row.raw_name || `节点 ${row.id}`)}</span>
            <span class="node-card-subtitle">${escapeHTML(row.source_name || "未知来源")}</span>
          </span>
        </span>
        <span class="node-card-chip-row">
          <span class="node-card-chip" data-node-card-chip>${formatStatus(row.status)}</span>
          <span class="node-card-chip" data-node-card-chip>${formatCell(row.protocol, "protocol")}</span>
          <span class="node-card-chip" data-node-card-chip>${labelForColumn("name_mode")}：${formatCell(row.name_mode, "name_mode")}</span>
        </span>
        <span class="node-card-endpoint" data-node-card-endpoint>
          <span aria-hidden="true">端</span>
          <code>${escapeHTML(endpoint)}</code>
        </span>
        <span class="node-card-tags">${formatCell(row.tags, "tags")}</span>
      </button>
      ${isEditing ? renderNodeEditForm(row) : ""}
      <div class="table-actions node-actions">
        ${
          isEditing
            ? `<button class="table-button ghost-button" type="button" data-node-action="cancel" data-node-id="${row.id}">${buttonLabel("×", "取消")}</button>`
            : `<button class="table-button" type="button" data-node-action="edit" data-node-id="${row.id}">${buttonLabel("✎", "编辑")}</button>`
        }
        <button class="table-button ghost-button" type="button" data-node-action="reset-name" data-node-id="${row.id}" ${row.name_mode === "auto" ? "disabled" : ""}>${buttonLabel("↺", "恢复自动")}</button>
      </div>
      ${detail ? renderNodeDetailPanel(detail) : ""}
    </article>
  `;
}

function nodeEndpointLabel(row) {
  const server = String(row?.server || "").trim();
  const port = String(row?.server_port || "").trim();
  if (server && port) return `${server}:${port}`;
  if (server) return server;
  if (port) return `:${port}`;
  return "未配置服务端";
}

function nodeProtocolSymbol(protocol) {
  const normalized = String(protocol || "")
    .trim()
    .replace(/[^a-z0-9+]/gi, "")
    .toUpperCase();
  const symbols = {
    VLESS: "VL",
    VMESS: "VM",
    TROJAN: "TR",
    SHADOWSOCKS: "SS",
    SS: "SS",
    HYSTERIA2: "H2",
    HY2: "H2",
    HYSTERIA: "HY",
    TUIC: "TU",
    JUICITY: "JU",
    ANYTLS: "AT",
    SHADOWTLS: "ST",
    NAIVE: "NV",
    HTTP: "HT",
    HTTPS: "HS",
    SOCKS: "SO",
    SOCKS5: "S5",
    SSH: "SH",
    WIREGUARD: "WG",
    TOR: "TO",
    DIRECT: "直",
    FREEDOM: "直",
    BLOCK: "拦",
    BLACKHOLE: "拦",
    REJECT: "拦",
    DNS: "DN",
  };
  return symbols[normalized] || normalized.slice(0, 2) || "--";
}

function renderNodeDetailPanel(node) {
  return `
    <div class="node-detail-row">
      <div class="node-detail-panel">
        <div class="node-detail-summary">
          <div class="node-detail-summary-title">
            <span class="node-card-protocol-symbol" aria-hidden="true">${escapeHTML(nodeProtocolSymbol(node.protocol))}</span>
            <div>
              <strong>${escapeHTML(node.display_name || node.raw_name || `节点 ${node.id}`)}</strong>
              <span>${escapeHTML(node.source_name || "未知来源")}</span>
            </div>
          </div>
          <div class="node-detail-summary-chips">
            <span data-node-detail-chip>${formatStatus(node.status)}</span>
            <span data-node-detail-chip>${formatCell(node.protocol, "protocol")}</span>
            <span data-node-detail-chip>${formatCell(node.region || "其他", "region")}</span>
            <span data-node-detail-chip>${formatCell(node.server, "server")}:${formatCell(node.server_port, "server_port")}</span>
          </div>
        </div>
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
          <div class="detail-item-heading">
            <span>${labelForColumn("uri")}</span>
            <button class="table-button ghost-button" type="button" data-node-action="copy-uri" data-node-id="${escapeHTML(String(node.id || ""))}" data-node-uri="${escapeHTML(String(node.uri || ""))}" data-button-symbol="⧉" data-copy-label="复制 URI">${buttonLabel("⧉", "复制 URI")}</button>
          </div>
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
      <button class="table-button" type="submit">${buttonLabel("✓", "保存")}</button>
    </form>
  `;
}

function renderTokens(rows) {
  appState.tokens = rows || [];
  if (!rows || rows.length === 0) {
    tokensEl.innerHTML = emptyState("钥", "暂无 Token", "签发 Token 后会生成通用、Mihomo 和 sing-box 订阅地址");
    return;
  }
  tokensEl.innerHTML = `
    <div class="token-card-list">
      ${rows.map((row) => renderTokenCard(row)).join("")}
    </div>
  `;
}

function renderTokenCard(row) {
  const usedBytes = (row.used_upload_bytes || 0) + (row.used_download_bytes || 0);
  const statusSummary = tokenStatusSummary(row.status);
  const expirySummary = tokenExpirySummary(row.expire_at);
  const quotaSummary = tokenQuotaSummary(usedBytes, row.quota_bytes, row.status);
  return `
    <article class="token-card" data-token-id="${row.id}">
      <div class="token-card-heading">
        <div>
          <span>Token 前缀</span>
          <strong>${formatCell(row.token_prefix, "token_prefix")}</strong>
        </div>
        ${formatCell(row.status, "status")}
      </div>
      <div class="token-card-summary" aria-label="Token 摘要">
        ${tokenSummaryChip(statusSummary.label, statusSummary.tone)}
        ${tokenSummaryChip(tokenUserSummary(row.user_id), "info")}
        ${tokenSummaryChip(expirySummary.label, expirySummary.tone)}
        ${tokenSummaryChip(quotaSummary.label, quotaSummary.tone)}
      </div>
      <div class="token-card-grid">
        ${tokenCardField("ID", row.id)}
        ${tokenCardField("成员 ID", row.user_id)}
        ${tokenCardField("名称", row.name)}
        ${tokenCardField("到期时间", row.expire_at, "expire_at")}
        ${tokenCardField("额度", row.quota_bytes, "quota_bytes")}
        ${tokenCardField("已用", usedBytes, "used_total")}
      </div>
      <div class="token-card-meter" data-token-quota-meter>
        ${tokenCardSectionHeading("额", "额度使用率", quotaSummary.label)}
        ${formatQuotaUsage(usedBytes, row.quota_bytes, row.status)}
      </div>
      <div class="token-card-subscriptions">
        ${tokenCardSectionHeading("订", "订阅地址", "通用 · Mihomo · sing-box")}
        ${renderTokenSubscriptions(row)}
      </div>
      <div class="token-card-control">
        ${tokenCardSectionHeading("控", "Token 操作", "续期 · 加额 · 状态")}
        <div class="token-card-actions">
          <span class="token-action-group" data-token-action-group>
            <label class="token-action-field">
              <span>续期天数</span>
              <span class="token-action-input">
                <input data-token-extend-days="${row.id}" type="number" min="1" value="30" aria-label="续期天数" />
                <em>天</em>
              </span>
            </label>
            <button class="table-button" data-token-action="extend" data-token-id="${row.id}">${buttonLabel("+", "续期")}</button>
          </span>
          <span class="token-action-group" data-token-action-group>
            <label class="token-action-field">
              <span>追加额度</span>
              <span class="token-action-input">
                <input data-token-quota-mib="${row.id}" type="number" min="1" value="1024" aria-label="追加额度 MiB" />
                <em>MiB</em>
              </span>
            </label>
            <button class="table-button" data-token-action="quota" data-token-id="${row.id}">${buttonLabel("+", "加额")}</button>
          </span>
          <span class="token-command-group" data-token-command-group>
            <span class="token-command-label">状态控制</span>
            <span class="token-command-buttons">
              <button class="table-button" data-token-action="restore" data-token-id="${row.id}">${buttonLabel("↺", "恢复")}</button>
              <button class="table-button ghost-button" data-token-action="rotate-subscription" data-token-id="${row.id}">${buttonLabel("⧉", "重置订阅")}</button>
              <button class="table-button danger-button" data-token-action="revoke" data-token-id="${row.id}">${buttonLabel("!", "撤销")}</button>
            </span>
          </span>
        </div>
      </div>
    </article>
  `;
}

function tokenCardSectionHeading(symbol, title, meta = "") {
  return `
    <div class="token-card-section-heading" data-token-section-heading>
      <span class="token-card-section-symbol" data-token-section-symbol aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="token-card-section-title">${escapeHTML(title)}</span>
      ${meta ? `<span class="token-card-section-meta">${escapeHTML(meta)}</span>` : ""}
    </div>
  `;
}

function tokenSummaryChip(content, tone = "") {
  return `<span class="token-card-summary-chip ${tone ? `token-card-summary-chip-${tone}` : ""}" data-token-summary-chip>${escapeHTML(content)}</span>`;
}

function tokenStatusSummary(value) {
  const status = String(value || "active");
  const normalized = status.toLowerCase();
  let tone = "muted";
  if (normalized === "active") {
    tone = "success";
  } else if (normalized === "expired" || normalized === "over_quota") {
    tone = "warning";
  } else if (normalized === "revoked" || normalized === "disabled") {
    tone = "danger";
  }
  return { label: status, tone };
}

function tokenUserSummary(userID) {
  if (!userID) return "未绑定成员";
  const user = (appState.users || []).find((item) => String(item.id) === String(userID));
  return user ? user.name : `成员 #${userID}`;
}

function tokenExpirySummary(value) {
  const date = parseDisplayTime(value);
  if (!date) {
    return { label: "无到期", tone: "muted" };
  }
  const diffMs = date.getTime() - Date.now();
  if (diffMs <= 0) {
    return { label: "已到期", tone: "danger" };
  }
  const days = Math.ceil(diffMs / 86400000);
  return {
    label: days <= 7 ? `${formatPlainNumber(days)} 天内到期` : `${formatPlainNumber(days)} 天到期`,
    tone: days <= 7 ? "warning" : "success",
  };
}

function tokenQuotaSummary(usedBytes, quotaBytes, status = "") {
  const quota = Number(quotaBytes || 0);
  if (quota <= 0) {
    return { label: "不限额度", tone: "info" };
  }
  const used = Math.max(0, Number(usedBytes || 0));
  const ratio = used / quota;
  const percent = Math.round(ratio * 1000) / 10;
  if (String(status || "").toLowerCase() === "over_quota" || ratio >= 1) {
    return { label: "额度用尽", tone: "danger" };
  }
  return {
    label: `${percent.toFixed(1)}% 已用`,
    tone: ratio >= 0.8 ? "warning" : "success",
  };
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
  return renderSubscriptionCardList(items, row.id);
}

function renderSubscriptionCardList(items, tokenID = "") {
  return `
    <div class="token-subscription-list">
      ${items
        .map(
          ([label, url]) => `
            <div class="token-subscription-item">
              <div class="token-subscription-heading">
                <span class="token-subscription-title">
                  <span>${escapeHTML(label)}</span>
                  <span class="token-subscription-kind" data-token-subscription-kind>${escapeHTML(subscriptionKindForLabel(label))}</span>
                </span>
                <span class="token-subscription-actions">
                  <a class="table-button ghost-button link-button" href="${escapeHTML(url)}" target="_blank" rel="noopener noreferrer" data-token-subscription-link>${buttonLabel("↗", "打开")}</a>
                  <button class="table-button ghost-button" type="button" data-token-action="copy-subscription" data-token-id="${escapeHTML(String(tokenID || ""))}" data-token-url="${escapeHTML(url)}" data-button-symbol="⧉" data-copy-label="复制">${buttonLabel("⧉", "复制")}</button>
                </span>
              </div>
              <span class="token-subscription-meta">
                <span class="token-subscription-profile" data-token-subscription-profile>${escapeHTML(subscriptionProfileForLabel(label))}</span>
                <span class="token-subscription-origin" data-token-subscription-origin>${escapeHTML(subscriptionOriginForURL(url))}</span>
              </span>
              <code title="${escapeHTML(url)}">${escapeHTML(url)}</code>
            </div>
          `,
        )
        .join("")}
    </div>
  `;
}

function buttonLabel(symbol, label) {
  return `<span class="button-symbol" aria-hidden="true">${escapeHTML(symbol)}</span><span class="button-label">${escapeHTML(label)}</span>`;
}

function subscriptionKindForLabel(label) {
  if (label === "Clash/Mihomo") return "Mihomo";
  if (label === "sing-box") return "sing-box";
  return "通用";
}

function subscriptionProfileForLabel(label) {
  if (label === "Clash/Mihomo") return "YAML · Mihomo";
  if (label === "sing-box") return "JSON · sing-box";
  return "URI · 通用";
}

function subscriptionOriginForURL(url) {
  try {
    const parsed = new URL(url, window.location.href);
    if (!parsed.host) return "相对地址";
    if (parsed.host === window.location.host) return "当前访问域名";
    return "配置域名";
  } catch {
    return "地址待确认";
  }
}

function showTokenSubscriptionResult(result, title) {
  const subscriptions = result.subscriptions || {};
  const defaultSubscription = subscriptions.default || result.subscription || "";
  const clashSubscription = subscriptions.clash || `${defaultSubscription}?target=clash`;
  const singBoxSubscription = subscriptions.sing_box || `${defaultSubscription}?target=sing-box`;
  const items = [
    ["默认", defaultSubscription],
    ["Clash/Mihomo", clashSubscription],
    ["sing-box", singBoxSubscription],
  ].filter((item) => item[1]);
  tokenResultEl.hidden = false;
  tokenResultEl.innerHTML = `
    <div class="token-result-card">
      <div class="token-result-heading">
        <strong>${escapeHTML(title)}</strong>
        <span>复制后可直接导入对应客户端</span>
      </div>
      ${items.length > 0 ? renderSubscriptionCardList(items, result.id || result.token_id || "") : `<div class="token-subscription-empty">暂无可显示的订阅地址</div>`}
    </div>
  `;
}

function renderTrafficTokens(rows) {
  if (!rows || rows.length === 0) {
    trafficTokensEl.innerHTML = `
      <div class="traffic-card-section" data-traffic-card-section>
        ${trafficSectionHeader("钥", "Token 用量", 0, "按团队 Token 聚合订阅与网关流量")}
        ${emptyState("钥", "暂无 Token 用量", "有 Token 订阅访问或网关流量后会显示用量")}
      </div>
    `;
    return;
  }
  trafficTokensEl.innerHTML = `
    <div class="traffic-card-section" data-traffic-card-section>
      ${trafficSectionHeader("钥", "Token 用量", rows.length, "按团队 Token 聚合订阅与网关流量")}
      <div class="traffic-card-list traffic-token-card-list">
        ${rows.map((row) => renderTrafficTokenCard(row)).join("")}
      </div>
    </div>
  `;
}

function renderTrafficOutbounds(rows) {
  if (!rows || rows.length === 0) {
    trafficOutboundsEl.innerHTML = `
      <div class="traffic-card-section" data-traffic-card-section>
        ${trafficSectionHeader("出", "出口摘要", 0, "按上游节点出口聚合真实转发流量")}
        ${emptyState("出", "暂无出口流量", "sing-box 网关产生真实流量后会按上游出口聚合")}
      </div>
    `;
    return;
  }
  trafficOutboundsEl.innerHTML = `
    <div class="traffic-card-section" data-traffic-card-section>
      ${trafficSectionHeader("出", "出口摘要", rows.length, "按上游节点出口聚合真实转发流量")}
      <div class="traffic-card-list traffic-outbound-card-list">
        ${rows.map((row) => renderTrafficOutboundCard(row)).join("")}
      </div>
    </div>
  `;
}

function trafficSectionHeader(symbol, title, count, hint) {
  return `
    <div class="traffic-card-section-heading">
      <span class="traffic-card-section-symbol" data-traffic-section-symbol aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="traffic-card-section-title">
        <strong>${escapeHTML(title)}</strong>
        <span>${escapeHTML(hint)}</span>
      </span>
      <span class="traffic-card-section-count" data-traffic-section-count>${formatPlainNumber(count)}</span>
    </div>
  `;
}

function renderTrafficTokenCard(row) {
  const statusSummary = tokenStatusSummary(row.token_status);
  return `
    <article class="traffic-card traffic-token-card" data-traffic-token-id="${escapeHTML(row.token_id || "")}">
      <div class="traffic-card-heading">
        <div>
          <span>Token #${formatCell(row.token_id, "token_id")}</span>
          <strong>${formatCell(row.auth_user, "auth_user")}</strong>
        </div>
        ${formatCell(row.token_status, "token_status")}
      </div>
      <div class="traffic-card-summary" aria-label="Token 用量摘要">
        ${trafficSummaryChip(statusSummary.label, statusSummary.tone)}
        ${trafficSummaryChip(`成员 #${row.user_id || "--"}`, "info")}
        ${trafficSummaryChip(`今日 ${formatBytes(row.today_total_bytes)}`)}
        ${trafficSummaryChip(`本月 ${formatBytes(row.month_total_bytes)}`)}
      </div>
      <div class="traffic-card-meter">
        ${formatQuotaUsage(row.used_total_bytes, row.quota_bytes, row.token_status)}
      </div>
      <div class="traffic-card-grid">
        ${trafficCardField("成员 ID", row.user_id, "user_id")}
        ${trafficCardField("今日", row.today_total_bytes, "today_total_bytes")}
        ${trafficCardField("本月", row.month_total_bytes, "month_total_bytes")}
        ${trafficCardField("累计上传", row.used_upload_bytes, "used_upload_bytes")}
        ${trafficCardField("累计下载", row.used_download_bytes, "used_download_bytes")}
        ${trafficCardField("累计总量", row.used_total_bytes, "used_total_bytes")}
        ${trafficCardField("额度", row.quota_bytes, "quota_bytes")}
        ${trafficCardField("更新时间", row.updated_at, "updated_at")}
      </div>
    </article>
  `;
}

function renderTrafficOutboundCard(row) {
  return `
    <article class="traffic-card traffic-outbound-card" data-traffic-outbound="${escapeHTML(row.outbound_tag || "")}">
      <div class="traffic-card-heading">
        <div>
          <span>${formatCell(row.source_name, "source_name")}</span>
          <strong>${formatCell(row.node_name, "node_name")}</strong>
        </div>
        <code title="${escapeHTML(row.outbound_tag || "")}">${formatCell(row.outbound_tag, "outbound_tag")}</code>
      </div>
      <div class="traffic-card-summary" aria-label="出口摘要">
        ${trafficSummaryChip(trafficSourceSummary(row.source_name), "info")}
        ${trafficSummaryChip(`上传 ${formatBytes(row.upload_bytes)}`)}
        ${trafficSummaryChip(`下载 ${formatBytes(row.download_bytes)}`)}
        ${trafficSummaryChip(`总量 ${formatBytes(row.total_bytes)}`, trafficTotalTone(row.total_bytes))}
      </div>
      <div class="traffic-card-grid">
        ${trafficCardField("节点 ID", row.upstream_node_id, "upstream_node_id")}
        ${trafficCardField("上传", row.upload_bytes, "upload_bytes")}
        ${trafficCardField("下载", row.download_bytes, "download_bytes")}
        ${trafficCardField("总量", row.total_bytes, "total_bytes")}
        ${trafficCardField("更新时间", row.updated_at, "updated_at")}
      </div>
    </article>
  `;
}

function trafficSummaryChip(content, tone = "") {
  return `<span class="traffic-card-summary-chip ${tone ? `traffic-card-summary-chip-${tone}` : ""}" data-traffic-summary-chip>${escapeHTML(content)}</span>`;
}

function trafficSourceSummary(value) {
  return value ? String(value) : "未知来源";
}

function trafficTotalTone(value) {
  return Number(value || 0) > 0 ? "success" : "muted";
}

function trafficCardField(label, value, column = "") {
  return `
    <div class="traffic-card-field">
      <span>${escapeHTML(label)}</span>
      <strong>${formatCell(value, column)}</strong>
    </div>
  `;
}

function renderTrafficDaily(rows) {
  renderTrafficBars(trafficDailyEl, rows, {
    heading: "最近 14 天",
    hint: "日流量趋势",
    symbol: "日",
    valueKey: "total_bytes",
    label: (row) => String(row.day || "").slice(5),
    title: (row, total) => `${row.day} ${formatBytes(total)}`,
    className: "traffic-chart-bars",
  });
}

function renderTrafficHourly(rows) {
  renderTrafficBars(trafficHourlyEl, rows, {
    heading: "最近 24 小时",
    hint: "小时流量趋势",
    symbol: "时",
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
    target.innerHTML = emptyState("量", "暂无流量数据", "真实网关流量进入统计链路后会生成趋势图");
    return;
  }
  const values = rows.map((row) => Number(row[options.valueKey] || 0));
  const totalSum = values.reduce((sum, value) => sum + value, 0);
  const peakIndex = values.reduce((maxIndex, value, index) => (value > values[maxIndex] ? index : maxIndex), 0);
  const peakRow = rows[peakIndex] || rows[0];
  const peakTotal = values[peakIndex] || 0;
  const activeCount = values.filter((value) => value > 0).length;
  const maxTotal = Math.max(...values, 1);
  target.innerHTML = `
    <div class="traffic-chart-panel" data-traffic-chart-panel>
      <div class="traffic-chart-heading">
        <span class="traffic-chart-symbol" data-traffic-chart-symbol aria-hidden="true">${escapeHTML(options.symbol || "量")}</span>
        <div class="traffic-chart-title">
          <strong>${escapeHTML(options.heading || "流量趋势")}</strong>
          <span>${escapeHTML(options.hint || "真实网关流量趋势")}</span>
        </div>
        <div class="traffic-chart-summary" aria-label="流量趋势摘要">
          ${trafficChartSummaryChip("总量", formatBytes(totalSum))}
          ${trafficChartSummaryChip("峰值", `${options.label(peakRow)} · ${formatBytes(peakTotal)}`)}
          ${trafficChartSummaryChip("活跃", `${activeCount}/${rows.length}`)}
        </div>
      </div>
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
    </div>
  `;
}

function trafficChartSummaryChip(label, value) {
  return `
    <span class="traffic-chart-summary-chip" data-traffic-chart-summary-chip>
      <span>${escapeHTML(label)}</span>
      <strong>${escapeHTML(value)}</strong>
    </span>
  `;
}

function renderTable(target, rows, columns) {
  if (!rows || rows.length === 0) {
    target.innerHTML = emptyState("表", "暂无数据", "数据同步或创建后会出现在这里");
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
  return String(value ?? "").replace(/[&<>"']/g, (char) => {
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
