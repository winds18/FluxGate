const statusEl = document.querySelector("#status");
const loginView = document.querySelector("#login-view");
const appView = document.querySelector("#app-view");
const viewSymbolEl = document.querySelector("#view-symbol");
const viewTitleEl = document.querySelector("#view-title");
const viewDescriptionEl = document.querySelector("#view-description");
const workspaceInsightEl = document.querySelector("#workspace-insight");
const viewContextEl = document.querySelector("#view-context");
const viewRailEl = document.querySelector("#view-rail");
const viewPrimaryActionEl = document.querySelector("#view-primary-action");
const dashboardViewSections = Array.from(document.querySelectorAll("[data-dashboard-view]"));
const dashboardNavButtons = Array.from(document.querySelectorAll("[data-view-nav]"));
const dashboardNavCountEls = Array.from(document.querySelectorAll("[data-view-count]"));
const dashboardJumpButtons = Array.from(document.querySelectorAll("[data-view-jump]"));
const overviewCardCountEls = Array.from(document.querySelectorAll("[data-overview-card-count]"));
const mobileMoreToggle = document.querySelector("[data-mobile-more-toggle]");
const mobileMoreMenu = document.querySelector("#mobile-more-menu");
const mobileOverflowViews = new Set(["policies", "traffic", "ops"]);
const loginForm = document.querySelector("#login-form");
const loginErrorEl = document.querySelector("#login-error");
const loginUsernameEl = document.querySelector("#login-username");
const loginPasswordEl = document.querySelector("#login-password");
const logoutEl = document.querySelector("#logout");
const overviewHeroEl = document.querySelector("#overview-hero");
const metricsEl = document.querySelector("#metrics");
const overviewInsightsEl = document.querySelector("#overview-insights");
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
const deliveryReadinessEl = document.querySelector("#delivery-readiness");
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
    ["初始化", "overview-readiness"],
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
    ["交付收口", "config-check-result"],
    ["配置操作", "config-check-result"],
  ],
};
let activeDashboardView = "overview";
let activeDashboardTarget = "";
let invalidFeedbackLocked = false;
let fallbackFieldFeedbackID = 0;
let refreshPendingDepth = 0;

refreshEl.addEventListener("click", load);
deliveryReadinessEl.addEventListener("click", checkDeliveryReadiness);
configCheckEl.addEventListener("click", checkConfig);
configPublishEl.addEventListener("click", publishConfig);
configRollbackEl.addEventListener("click", rollbackConfig);
configRestartEl.addEventListener("click", restartSingBox);
logoutEl.addEventListener("click", logout);
viewPrimaryActionEl?.addEventListener("click", () => {
  const action = workspacePrimaryActionForView(activeDashboardView);
  if (!action) return;
  navigateToDashboardTarget(action.view || activeDashboardView, action.target || "", Boolean(action.expand));
});
loginForm.addEventListener("submit", login);
teamForm.addEventListener("submit", submitTeam);
userForm.addEventListener("submit", submitUser);
sourceForm.addEventListener("submit", submitSource);
nodeImportForm.addEventListener("submit", submitNodeImport);
virtualNodeForm.addEventListener("submit", submitVirtualNode);
policyForm.addEventListener("submit", submitPolicy);
tokenForm.addEventListener("submit", submitToken);
document.addEventListener("invalid", handleInvalidField, true);
document.addEventListener("input", clearInvalidFieldFeedback, true);
document.addEventListener("change", clearInvalidFieldFeedback, true);
document.addEventListener("input", handleFormDrawerDraft, true);
document.addEventListener("change", handleFormDrawerDraft, true);
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
  button.addEventListener("click", () => {
    navigateToDashboardTarget(
      button.dataset.viewJump || "overview",
      button.dataset.viewJumpTarget || "",
      button.dataset.viewJumpExpand === "true",
    );
  });
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
  const button = event.target.closest("[data-form-drawer-cancel]");
  if (!button) return;
  const drawer = button.closest("[data-form-drawer]");
  if (!drawer) return;
  cancelFormDrawer(drawer);
});
mobileMoreToggle?.addEventListener("click", (event) => {
  event.stopPropagation();
  setMobileMoreMenuOpen(mobileMoreMenu?.hidden !== false);
});
mobileMoreMenu?.addEventListener("click", (event) => {
  if (event.target.closest("[data-view-nav]")) {
    setMobileMoreMenuOpen(false);
  }
});
document.addEventListener("click", (event) => {
  if (mobileMoreMenu?.hidden !== false) return;
  if (event.target.closest("#mobile-more-menu") || event.target.closest("[data-mobile-more-toggle]")) return;
  setMobileMoreMenuOpen(false);
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") {
    setMobileMoreMenuOpen(false);
    collapseActiveFormDrawer();
  }
});
document.addEventListener("click", (event) => {
  const button = event.target.closest("[data-overview-jump]");
  if (!button) return;
  navigateToDashboardTarget(
    button.dataset.overviewJump || "overview",
    button.dataset.overviewJumpTarget || "",
    button.dataset.overviewJumpExpand === "true",
  );
});
document.addEventListener("click", (event) => {
  const button = event.target.closest("[data-workspace-insight-action]");
  if (!button) return;
  navigateToDashboardTarget(
    button.dataset.workspaceInsightView || activeDashboardView,
    button.dataset.workspaceInsightTarget || "",
    button.dataset.workspaceInsightExpand === "true",
  );
});
document.addEventListener("click", (event) => {
  const button = event.target.closest("[data-view-rail-target]");
  if (!button) return;
  navigateToDashboardTarget(activeDashboardView, button.dataset.viewRailTarget || "", button.dataset.viewRailExpand === "true");
});
document.addEventListener("click", (event) => {
  const button = event.target.closest("[data-empty-state-action]");
  if (!button) return;
  navigateToDashboardTarget(
    button.dataset.emptyStateView || activeDashboardView,
    button.dataset.emptyStateTarget || "",
    button.dataset.emptyStateExpand === "true",
  );
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
  deliveryReadiness: {},
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

function setStatus(message, tone = "info") {
  statusEl.textContent = message;
  statusEl.dataset.statusTone = tone;
  statusEl.setAttribute("aria-live", tone === "danger" || tone === "warning" ? "assertive" : "polite");
}

function handleInvalidField(event) {
  const field = event.target;
  if (!isFormControl(field)) return;
  event.preventDefault();
  if (invalidFeedbackLocked) return;
  invalidFeedbackLocked = true;
  window.setTimeout(() => {
    invalidFeedbackLocked = false;
  }, 0);

  const form = field.closest("form");
  clearInvalidFieldFeedback({ target: field, currentTarget: form || document, clearAll: true });
  field.classList.add("is-field-invalid");
  field.closest("label")?.classList.add("is-field-invalid");
  showInvalidFieldFeedback(field);

  const drawer = field.closest("[data-form-drawer]");
  if (drawer) {
    setFormDrawerCollapsed(drawer, false);
  }
  const target = drawer || form?.closest(".panel") || form || field.closest(".panel") || field;
  if (target?.id) {
    activeDashboardTarget = target.id;
    updateViewRail();
  }
  setStatus(validationStatusForField(field), "warning");
  highlightDashboardTarget(target);
  target.scrollIntoView({ behavior: "smooth", block: "center", inline: "nearest" });
  field.focus({ preventScroll: true });
}

function clearInvalidFieldFeedback(event) {
  const scope = event.clearAll ? event.currentTarget || document : null;
  if (scope?.querySelectorAll) {
    scope.querySelectorAll("[data-field-feedback]").forEach((element) => element.remove());
    scope.querySelectorAll(".is-field-invalid").forEach((element) => element.classList.remove("is-field-invalid"));
    scope.querySelectorAll('[aria-invalid="true"]').forEach((element) => clearFieldInvalidAttributes(element));
    return;
  }
  const field = event.target;
  if (!isFormControl(field)) return;
  field.classList.remove("is-field-invalid");
  field.closest("label")?.classList.remove("is-field-invalid");
  clearFieldInvalidAttributes(field);
}

function isFormControl(element) {
  return element instanceof HTMLInputElement || element instanceof HTMLSelectElement || element instanceof HTMLTextAreaElement;
}

function validationStatusForField(field) {
  const label = labelForField(field);
  const validity = field.validity;
  if (validity?.valueMissing) {
    return `请先补齐：${label}`;
  }
  if (validity?.typeMismatch) {
    return `请检查格式：${label}`;
  }
  if (validity?.rangeUnderflow || validity?.rangeOverflow || validity?.stepMismatch) {
    return `请检查数值：${label}`;
  }
  return `请检查：${label}`;
}

function showInvalidFieldFeedback(field) {
  const feedbackID = fieldFeedbackID(field);
  let feedback = document.getElementById(feedbackID);
  if (!feedback) {
    feedback = document.createElement("p");
    feedback.id = feedbackID;
    feedback.className = "field-feedback";
    feedback.dataset.fieldFeedback = "";
    field.insertAdjacentElement("afterend", feedback);
  }
  feedback.textContent = fieldFeedbackMessage(field);
  field.setAttribute("aria-invalid", "true");
  addDescribedByID(field, feedbackID);
}

function clearFieldInvalidAttributes(field) {
  if (!isFormControl(field)) return;
  const feedbackID = field.dataset.fieldFeedbackID || fieldFeedbackID(field);
  document.getElementById(feedbackID)?.remove();
  removeDescribedByID(field, feedbackID);
  field.removeAttribute("aria-invalid");
  delete field.dataset.fieldFeedbackID;
}

function fieldFeedbackID(field) {
  if (field.dataset.fieldFeedbackID) return field.dataset.fieldFeedbackID;
  const formID = field.closest("form")?.id || "form";
  const fieldKey = field.name || field.id || field.getAttribute("aria-label") || `field-${++fallbackFieldFeedbackID}`;
  const id = `${formID}-${fieldKey}-feedback`.replace(/[^A-Za-z0-9_-]+/g, "-");
  field.dataset.fieldFeedbackID = id;
  return id;
}

function fieldFeedbackMessage(field) {
  const label = labelForField(field);
  const validity = field.validity;
  if (validity?.valueMissing) return `请填写${label}`;
  if (validity?.typeMismatch) return `请检查${label}的格式`;
  if (validity?.rangeUnderflow || validity?.rangeOverflow || validity?.stepMismatch) return `请检查${label}的数值`;
  return `请检查${label}`;
}

function addDescribedByID(field, id) {
  const ids = new Set(String(field.getAttribute("aria-describedby") || "").split(/\s+/).filter(Boolean));
  ids.add(id);
  field.setAttribute("aria-describedby", Array.from(ids).join(" "));
}

function removeDescribedByID(field, id) {
  const ids = String(field.getAttribute("aria-describedby") || "")
    .split(/\s+/)
    .filter((value) => value && value !== id);
  if (ids.length) {
    field.setAttribute("aria-describedby", ids.join(" "));
  } else {
    field.removeAttribute("aria-describedby");
  }
}

function labelForField(field) {
  const label = field.closest("label");
  if (label) {
    const text = Array.from(label.childNodes)
      .filter((node) => node.nodeType === Node.TEXT_NODE)
      .map((node) => node.textContent || "")
      .join("")
      .replace(/\s+/g, " ")
      .replace(/[：:]\s*$/, "")
      .trim();
    if (text) return text;
  }
  return field.getAttribute("aria-label") || field.getAttribute("placeholder") || field.name || "这个字段";
}

function confirmDanger({ title, message, confirmLabel = "确认", cancelLabel = "取消" }) {
  const previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  return new Promise((resolve) => {
    const dialog = document.createElement("div");
    dialog.className = "confirm-dialog-backdrop";
    dialog.dataset.confirmDialog = "";
    dialog.setAttribute("role", "presentation");
    dialog.innerHTML = `
      <section class="confirm-dialog" role="dialog" aria-modal="true" aria-labelledby="confirm-dialog-title" aria-describedby="confirm-dialog-message">
        <div class="confirm-dialog-symbol" aria-hidden="true">!</div>
        <div class="confirm-dialog-content">
          <h3 id="confirm-dialog-title" data-confirm-title>${escapeHTML(title || "确认操作")}</h3>
          <p id="confirm-dialog-message">${escapeHTML(message || "这个操作会立即生效，请确认后继续。")}</p>
          <div class="confirm-dialog-actions">
            <button class="table-button ghost-button" type="button" data-confirm-cancel>${escapeHTML(cancelLabel)}</button>
            <button class="table-button danger-button" type="button" data-confirm-accept>${escapeHTML(confirmLabel)}</button>
          </div>
        </div>
      </section>
    `;

    const finish = (confirmed) => {
      document.removeEventListener("keydown", handleKeydown);
      dialog.remove();
      previousFocus?.focus?.();
      resolve(confirmed);
    };
    const handleKeydown = (event) => {
      if (event.key === "Escape") {
        event.preventDefault();
        finish(false);
      }
    };

    dialog.addEventListener("click", (event) => {
      if (event.target === dialog) finish(false);
    });
    dialog.querySelector("[data-confirm-cancel]")?.addEventListener("click", () => finish(false));
    dialog.querySelector("[data-confirm-accept]")?.addEventListener("click", () => finish(true));
    document.addEventListener("keydown", handleKeydown);
    document.body.append(dialog);
    requestAnimationFrame(() => dialog.querySelector("[data-confirm-cancel]")?.focus());
  });
}

async function bootstrap() {
  setStatus("连接中", "loading");
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
  setStatus("登录中", "loading");
  try {
    const username = loginUsernameEl.value.trim();
    const password = loginPasswordEl.value;
    await postJSON("/api/auth/login", { username, password });
    loginPasswordEl.value = "";
    showApp();
    await load();
  } catch (error) {
    setStatus("未登录", "warning");
    loginErrorEl.textContent = "账号或密码不正确";
  }
}

async function logout() {
  setLogoutPending(true);
  setStatus("退出中", "loading");
  try {
    await fetch("/api/auth/logout", { method: "POST" });
  } finally {
    setLogoutPending(false);
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
}

function showLogin() {
  document.querySelectorAll("[data-form-drawer]").forEach((drawer) => applyFormDrawerCollapsed(drawer, true));
  syncFormDrawerShellState();
  appView.hidden = true;
  loginView.hidden = false;
  logoutEl.hidden = true;
  setStatus("未登录", "warning");
  loginUsernameEl.focus();
}

function showApp() {
  loginView.hidden = true;
  appView.hidden = false;
  logoutEl.hidden = false;
  setStatus("已登录", "success");
  setActiveView(activeDashboardView);
}

function setActiveView(view) {
  const nextView = dashboardViewMeta[view] ? view : "overview";
  activeDashboardView = nextView;
  dashboardViewSections.forEach((section) => {
    section.hidden = section.dataset.dashboardView !== nextView;
  });
  const activeDrawer = document.querySelector("[data-form-drawer]:not(.is-collapsed)");
  if (activeDrawer && !activeDrawer.closest(`[data-dashboard-view="${nextView}"]`)) {
    setFormDrawerCollapsed(activeDrawer, true);
  }
  dashboardNavButtons.forEach((button) => {
    const isActive = button.dataset.viewNav === nextView;
    button.classList.toggle("is-active", isActive);
    button.setAttribute("aria-current", isActive ? "page" : "false");
  });
  const [title, description] = dashboardViewMeta[nextView];
  if (viewSymbolEl) viewSymbolEl.textContent = dashboardViewSymbols[nextView] || title.slice(0, 1);
  viewTitleEl.textContent = title;
  viewDescriptionEl.textContent = description;
  updateWorkspaceInsight();
  updateViewContext();
  updateViewRail();
  updateWorkspacePrimaryAction();
  updateNavigationCounts();
  updateOverviewCardCounts();
  mobileMoreToggle?.classList.toggle("is-active", mobileOverflowViews.has(nextView));
  if (!mobileOverflowViews.has(nextView)) {
    setMobileMoreMenuOpen(false);
  }
}

function setMobileMoreMenuOpen(open) {
  if (!mobileMoreMenu || !mobileMoreToggle) return;
  mobileMoreMenu.hidden = !open;
  mobileMoreToggle.setAttribute("aria-expanded", open ? "true" : "false");
}

function setFormDrawerCollapsed(drawer, collapsed) {
  if (!collapsed) {
    document.querySelectorAll("[data-form-drawer]").forEach((otherDrawer) => {
      if (otherDrawer !== drawer) {
        applyFormDrawerCollapsed(otherDrawer, true);
      }
    });
  }
  applyFormDrawerCollapsed(drawer, collapsed);
  syncFormDrawerShellState();
}

function applyFormDrawerCollapsed(drawer, collapsed) {
  drawer.classList.toggle("is-collapsed", collapsed);
  drawer.classList.toggle("is-side-sheet", !collapsed);
  if (!collapsed) {
    drawer.setAttribute("role", "dialog");
    drawer.setAttribute("aria-modal", "true");
  } else {
    drawer.removeAttribute("role");
    drawer.removeAttribute("aria-modal");
  }
  const toggle = drawer.querySelector("[data-form-drawer-toggle]");
  if (toggle) {
    toggle.innerHTML = buttonLabel(collapsed ? "+" : "−", collapsed ? "展开" : "收起");
    toggle.setAttribute("aria-expanded", collapsed ? "false" : "true");
  }
}

function syncFormDrawerShellState() {
  const activeDrawer = document.querySelector("[data-form-drawer]:not(.is-collapsed)");
  document.body.classList.toggle("has-form-drawer-open", Boolean(activeDrawer));
  if (activeDrawer?.id) {
    document.body.dataset.activeFormDrawer = activeDrawer.id;
  } else {
    delete document.body.dataset.activeFormDrawer;
  }
}

function collapseActiveFormDrawer() {
  const activeDrawer = document.querySelector("[data-form-drawer]:not(.is-collapsed)");
  if (!activeDrawer) return;
  setFormDrawerCollapsed(activeDrawer, true);
  setStatus(`已收起：${dashboardTargetLabel(activeDrawer.id)}`, "info");
  activeDrawer.querySelector("[data-form-drawer-toggle]")?.focus({ preventScroll: true });
}

function handleFormDrawerDraft(event) {
  const field = event.target;
  if (!isFormControl(field)) return;
  const drawer = field.closest("[data-form-drawer]");
  if (!drawer) return;
  setFormDrawerDirty(drawer, true);
}

function setFormDrawerDirty(drawer, dirty) {
  drawer.classList.toggle("is-dirty", dirty);
  const draft = drawer.querySelector("[data-form-draft]");
  if (draft) {
    draft.hidden = !dirty;
  }
}

function setFormDrawerSubmitting(drawer, submitting) {
  drawer.classList.toggle("is-submitting", submitting);
  if (submitting) {
    drawer.setAttribute("aria-busy", "true");
  } else {
    drawer.removeAttribute("aria-busy");
  }

  const submitButton = drawer.querySelector('button[type="submit"]');
  setSubmitButtonPending(submitButton, submitting);
  drawer.querySelectorAll("input, select, textarea, button").forEach((control) => {
    if (submitting) {
      if (control.disabled) {
        control.dataset.formWasDisabled = "true";
      } else {
        delete control.dataset.formWasDisabled;
      }
      control.disabled = true;
      return;
    }
    control.disabled = control.dataset.formWasDisabled === "true";
    delete control.dataset.formWasDisabled;
  });
}

function setSubmitButtonPending(button, submitting) {
  if (!button) return;
  const symbol = button.querySelector(".button-symbol");
  const label = button.querySelector(".button-label");
  if (submitting) {
    if (!button.dataset.formOriginalSymbol) {
      button.dataset.formOriginalSymbol = symbol?.textContent || "";
    }
    if (!button.dataset.formOriginalLabel) {
      button.dataset.formOriginalLabel = label?.textContent || "提交";
    }
    button.classList.add("is-pending");
    if (symbol) symbol.textContent = "…";
    if (label) label.textContent = `${button.dataset.formOriginalLabel}中`;
    return;
  }
  button.classList.remove("is-pending");
  if (symbol && button.dataset.formOriginalSymbol) {
    symbol.textContent = button.dataset.formOriginalSymbol;
  }
  if (label && button.dataset.formOriginalLabel) {
    label.textContent = button.dataset.formOriginalLabel;
  }
  delete button.dataset.formOriginalSymbol;
  delete button.dataset.formOriginalLabel;
}

function setRefreshPending(pending) {
  if (!refreshEl) return;
  refreshPendingDepth = Math.max(0, refreshPendingDepth + (pending ? 1 : -1));
  const active = refreshPendingDepth > 0;

  appView?.classList.toggle("is-refreshing", active);
  if (active) {
    appView?.setAttribute("aria-busy", "true");
  } else {
    appView?.removeAttribute("aria-busy");
  }

  if (active) {
    if (!refreshEl.dataset.refreshPendingStored) {
      refreshEl.dataset.refreshPendingStored = "true";
      refreshEl.dataset.refreshWasDisabled = refreshEl.disabled ? "true" : "false";
    }
    setSubmitButtonPending(refreshEl, true);
    refreshEl.disabled = true;
    return;
  }

  setSubmitButtonPending(refreshEl, false);
  refreshEl.disabled = refreshEl.dataset.refreshWasDisabled === "true";
  delete refreshEl.dataset.refreshPendingStored;
  delete refreshEl.dataset.refreshWasDisabled;
}

function setLogoutPending(pending) {
  if (!logoutEl) return;
  appView?.classList.toggle("is-logging-out", pending);
  if (pending) {
    appView?.setAttribute("aria-busy", "true");
  } else if (refreshPendingDepth === 0) {
    appView?.removeAttribute("aria-busy");
  }

  if (pending) {
    if (!logoutEl.dataset.logoutPendingStored) {
      logoutEl.dataset.logoutPendingStored = "true";
      logoutEl.dataset.logoutWasDisabled = logoutEl.disabled ? "true" : "false";
    }
    setSubmitButtonPending(logoutEl, true);
    logoutEl.disabled = true;
    return;
  }

  setSubmitButtonPending(logoutEl, false);
  logoutEl.disabled = logoutEl.dataset.logoutWasDisabled === "true";
  delete logoutEl.dataset.logoutPendingStored;
  delete logoutEl.dataset.logoutWasDisabled;
}

function setInlineActionPending(container, button, pending) {
  if (!container) return;
  container.classList.toggle("is-action-pending", pending);
  if (pending) {
    container.setAttribute("aria-busy", "true");
  } else {
    container.removeAttribute("aria-busy");
  }

  setSubmitButtonPending(button, pending);
  container.querySelectorAll("input, select, textarea, button").forEach((control) => {
    if (pending) {
      if (control.disabled) {
        control.dataset.actionWasDisabled = "true";
      } else {
        delete control.dataset.actionWasDisabled;
      }
      control.disabled = true;
      return;
    }
    control.disabled = control.dataset.actionWasDisabled === "true";
    delete control.dataset.actionWasDisabled;
  });
}

async function runInlineAction(container, button, statusMessage, action, failureMessage = "保存失败") {
  if (container?.dataset.actionPending === "true") {
    return { ok: false, skipped: true };
  }
  if (container) {
    container.dataset.actionPending = "true";
  }
  setInlineActionPending(container, button, true);
  setStatus(statusMessage, "loading");
  try {
    const value = await action();
    return { ok: true, value };
  } catch (error) {
    setStatus(failureMessage, "danger");
    return { ok: false, error };
  } finally {
    if (container) {
      delete container.dataset.actionPending;
    }
    setInlineActionPending(container, button, false);
  }
}

function setOpsActionPending(card, button, pending) {
  const grid = document.querySelector("#ops-actions");
  card?.classList.toggle("is-action-pending", pending);
  if (pending) {
    card?.setAttribute("aria-busy", "true");
    grid?.setAttribute("aria-busy", "true");
  } else {
    card?.removeAttribute("aria-busy");
    grid?.removeAttribute("aria-busy");
  }

  setSubmitButtonPending(button, pending);
  grid?.querySelectorAll("button").forEach((control) => {
    if (pending) {
      if (control.disabled) {
        control.dataset.opsWasDisabled = "true";
      } else {
        delete control.dataset.opsWasDisabled;
      }
      control.disabled = true;
      return;
    }
    control.disabled = control.dataset.opsWasDisabled === "true";
    delete control.dataset.opsWasDisabled;
  });
}

async function runOpsAction(button, statusMessage, action) {
  const grid = document.querySelector("#ops-actions");
  const card = button?.closest(".ops-action-card");
  if (grid?.dataset.actionPending === "true") {
    return { ok: false, skipped: true };
  }
  if (grid) {
    grid.dataset.actionPending = "true";
  }
  setOpsActionPending(card, button, true);
  setStatus(statusMessage, "loading");
  try {
    const value = await action();
    return { ok: true, value };
  } catch (error) {
    setStatus("操作失败，请查看结果", "danger");
    return { ok: false, error };
  } finally {
    if (grid) {
      delete grid.dataset.actionPending;
    }
    setOpsActionPending(card, button, false);
  }
}

async function runFormDrawerSubmit(drawer, action) {
  if (drawer.dataset.formSubmitting === "true") {
    return { ok: false, skipped: true };
  }
  drawer.dataset.formSubmitting = "true";
  setFormDrawerSubmitting(drawer, true);
  setStatus(`正在提交：${dashboardTargetLabel(drawer.id)}`, "loading");
  try {
    const value = await action();
    return { ok: true, value };
  } catch (error) {
    setStatus(`提交失败：${dashboardTargetLabel(drawer.id)}`, "danger");
    return { ok: false, error };
  } finally {
    delete drawer.dataset.formSubmitting;
    setFormDrawerSubmitting(drawer, false);
  }
}

function cancelFormDrawer(drawer) {
  drawer.reset();
  clearInvalidFieldFeedback({ currentTarget: drawer, clearAll: true });
  setFormDrawerDirty(drawer, false);
  drawer.querySelectorAll(".result-box").forEach((element) => {
    element.hidden = true;
    element.replaceChildren();
  });
  setFormDrawerCollapsed(drawer, true);
  if (activeDashboardTarget === drawer.id) {
    activeDashboardTarget = "";
    updateViewRail();
  }
  setStatus(`已取消：${dashboardTargetLabel(drawer.id)}`, "info");
  drawer.querySelector("[data-form-drawer-toggle]")?.focus({ preventScroll: true });
}

function completeFormDrawerSuccess(drawer) {
  drawer.reset();
  clearInvalidFieldFeedback({ currentTarget: drawer, clearAll: true });
  setFormDrawerDirty(drawer, false);
  setFormDrawerCollapsed(drawer, true);
}

async function load() {
  setRefreshPending(true);
  setStatus("刷新中", "loading");
  try {
    const [
      overview,
      teams,
      users,
      sources,
      nodes,
      virtualNodes,
      policies,
      tokens,
      deliveryReadiness,
      trafficHourly,
      trafficDaily,
      trafficOutbounds,
      trafficTokens,
    ] = await Promise.all([
      getJSON("/api/overview"),
      getJSON("/api/teams"),
      getJSON("/api/users"),
      getJSON("/api/sources"),
      getJSON("/api/nodes"),
      getJSON("/api/virtual-nodes"),
      getJSON("/api/policies"),
      getJSON("/api/tokens"),
      getJSON("/api/delivery/readiness"),
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
      deliveryReadiness,
      trafficHourly,
      trafficDaily,
      trafficOutbounds,
      trafficTokens,
    };
    renderOverviewHero(overview);
    renderMetrics(overview);
    renderOverviewReadiness(overview);
    renderOverviewInsights(overview);
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
    updateWorkspaceInsight();
    updateViewContext();
    updateViewRail();
    updateWorkspacePrimaryAction();
    updateNavigationCounts();
    updateOverviewCardCounts();
    setStatus("已连接", "success");
  } catch (error) {
    if (error.status === 401) {
      showLogin();
      return;
    }
    setStatus("异常", "danger");
    if (overviewHeroEl) {
      overviewHeroEl.innerHTML = emptyState("警", "加载失败", error.message || "请稍后重试");
    }
    metricsEl.innerHTML = emptyState("警", "加载失败", error.message || "请稍后重试");
    if (overviewInsightsEl) {
      overviewInsightsEl.innerHTML = emptyState("警", "加载失败", error.message || "请稍后重试");
    }
  } finally {
    setRefreshPending(false);
  }
}

async function submitTeam(event) {
  event.preventDefault();
  const form = new FormData(teamForm);
  const result = await runFormDrawerSubmit(teamForm, () => postAndReload("/api/teams", {
    name: textField(form, "name"),
    description: textField(form, "description"),
  }));
  if (!result.ok) return;
  completeFormDrawerSuccess(teamForm);
}

async function submitUser(event) {
  event.preventDefault();
  const form = new FormData(userForm);
  const teamID = numberField(form, "team_id");
  const result = await runFormDrawerSubmit(userForm, () => postAndReload("/api/users", {
    team_id: teamID > 0 ? teamID : null,
    name: textField(form, "name"),
    email: textField(form, "email"),
  }));
  if (!result.ok) return;
  completeFormDrawerSuccess(userForm);
}

async function submitSource(event) {
  event.preventDefault();
  const form = new FormData(sourceForm);
  const result = await runFormDrawerSubmit(sourceForm, () => postAndReload("/api/sources", {
    name: textField(form, "name"),
    type: textField(form, "type") || "manual",
    url: textField(form, "url"),
    default_tags: textField(form, "default_tags"),
    refresh_interval_minutes: numberField(form, "refresh_interval_minutes"),
  }));
  if (!result.ok) return;
  completeFormDrawerSuccess(sourceForm);
}

async function submitNodeImport(event) {
  event.preventDefault();
  const form = new FormData(nodeImportForm);
  const result = await runFormDrawerSubmit(nodeImportForm, () => postAndReload("/api/nodes/import", {
    source_id: numberField(form, "source_id"),
    content: textField(form, "content"),
  }));
  if (!result.ok) return;
  completeFormDrawerSuccess(nodeImportForm);
}

async function submitVirtualNode(event) {
  event.preventDefault();
  const form = new FormData(virtualNodeForm);
  const result = await runFormDrawerSubmit(virtualNodeForm, () => postAndReload("/api/virtual-nodes", {
    name: textField(form, "name"),
    listen_protocol: "vless",
    listen_port: numberField(form, "listen_port"),
    tag_selector: textField(form, "tag_selector"),
  }));
  if (!result.ok) return;
  completeFormDrawerSuccess(virtualNodeForm);
}

async function submitPolicy(event) {
  event.preventDefault();
  const form = new FormData(policyForm);
  const scopeID = numberField(form, "scope_id");
  const result = await runFormDrawerSubmit(policyForm, () => postAndReload("/api/policies", {
    name: textField(form, "name"),
    scope_type: textField(form, "scope_type") || "team",
    scope_id: scopeID > 0 ? scopeID : null,
    allowed_virtual_nodes: textField(form, "allowed_virtual_nodes"),
    include_tags: textField(form, "include_tags"),
    exclude_tags: textField(form, "exclude_tags"),
    max_nodes: numberField(form, "max_nodes"),
  }));
  if (!result.ok) return;
  completeFormDrawerSuccess(policyForm);
}

async function submitToken(event) {
  event.preventDefault();
  const form = new FormData(tokenForm);
  const quotaMiB = numberField(form, "quota_mib");
  const result = await runFormDrawerSubmit(tokenForm, () => postAndReload("/api/tokens", {
    user_id: numberField(form, "user_id"),
    name: textField(form, "name"),
    expire_days: numberField(form, "expire_days"),
    quota_bytes: quotaMiB * 1024 * 1024,
  }));
  if (!result.ok) return;
  showTokenSubscriptionResult(result.value, "订阅地址");
  completeFormDrawerSuccess(tokenForm);
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
  const row = button.closest(".identity-card[data-team-id]");
  if (!row) return;
  const id = row.dataset.teamId;
  const payload = {
    name: teamFieldValue(row, "name"),
    description: teamFieldValue(row, "description"),
    status: teamFieldValue(row, "status") || "active",
  };
  await runInlineAction(row, button, "保存团队中", async () => {
    await patchJSON(`/api/teams/${id}`, payload);
    appState.editingTeamID = null;
    await load();
  });
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
  const row = button.closest(".identity-card[data-user-id]");
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
  await runInlineAction(row, button, "保存成员中", async () => {
    await patchJSON(`/api/users/${id}`, payload);
    appState.editingUserID = null;
    await load();
  });
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
  const row = button.closest(".source-card[data-source-id]") || button;
  await runInlineAction(row, button, "刷新来源中", async () => {
    await postJSON(`/api/sources/${id}/refresh`, {});
    await load();
  }, "刷新失败");
}

async function saveSource(button) {
  const row = button.closest(".source-card[data-source-id]");
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
  await runInlineAction(row, button, "保存来源中", async () => {
    await patchJSON(`/api/sources/${id}`, payload);
    appState.editingSourceID = null;
    await load();
  });
}

async function regenerateSourceNames(button) {
  const id = button.dataset.sourceId;
  const row = button.closest(".source-card[data-source-id]") || button;
  await runInlineAction(row, button, "同步节点命名中", async () => {
    await postJSON(`/api/sources/${id}/regenerate-node-names`, {});
    await load();
  }, "同步失败");
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
  const row = button.closest(".virtual-node-card[data-virtual-node-id]");
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
  await runInlineAction(row, button, "保存虚拟节点中", async () => {
    await patchJSON(`/api/virtual-nodes/${id}`, payload);
    appState.editingVirtualNodeID = null;
    await load();
  });
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
  const row = button.closest(".policy-card[data-policy-id]");
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
  await runInlineAction(row, button, "保存策略中", async () => {
    await patchJSON(`/api/policies/${id}`, payload);
    appState.editingPolicyID = null;
    await load();
  });
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
  const action = button.dataset.nodeAction;
  if (action === "close-detail") {
    appState.expandedNodeID = null;
    appState.nodeDetail = null;
    renderNodes(appState.nodes);
    setStatus("节点详情已收起", "info");
    return;
  }
  const id = Number.parseInt(button.dataset.nodeId || "0", 10);
  if (!id) return;
  if (action === "copy-uri") {
    button.disabled = true;
    try {
      await copyText(button.dataset.nodeUri || "");
      setStatus("节点 URI 已复制", "success");
      showCopyFeedback(button, "已复制", "复制 URI");
      button.disabled = false;
    } catch (error) {
      setStatus("复制失败", "danger");
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
      setStatus("节点详情已收起", "info");
      return;
    }
    button.disabled = true;
    setStatus("加载节点详情中", "loading");
    try {
      const detail = await getJSON(`/api/nodes/${id}`);
      appState.expandedNodeID = id;
      appState.nodeDetail = detail;
      appState.editingNodeID = null;
      renderNodes(appState.nodes);
      setStatus("节点详情已打开", "success");
      nodesEl.querySelector(".node-detail-close")?.focus();
    } catch (error) {
      appState.expandedNodeID = null;
      appState.nodeDetail = null;
      setStatus("详情加载失败", "danger");
      button.disabled = false;
    }
    return;
  }
  if (action !== "reset-name") return;
  const card = button.closest(".node-card") || button;
  await runInlineAction(card, button, "恢复节点命名中", async () => {
    await postJSON(`/api/nodes/${id}/reset-display-name`, {});
    appState.editingNodeID = null;
    await load();
  }, "恢复失败");
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
  const formData = new FormData(form);
  const displayName = textField(formData, "display_name");
  const region = textField(formData, "region");
  const tags = textField(formData, "tags");
  const nameMode = textField(formData, "name_mode") || "manual";
  const button = form.querySelector('button[type="submit"]');
  const card = form.closest(".node-card") || form;
  await runInlineAction(card, button, "保存节点中", async () => {
    await patchJSON(`/api/nodes/${id}`, { display_name: displayName, region, tags, name_mode: nameMode });
    appState.editingNodeID = null;
    await load();
  });
}

async function handleTokenAction(event) {
  const button = event.target.closest("button[data-token-action]");
  if (!button) return;
  const id = button.dataset.tokenId;
  const action = button.dataset.tokenAction;
  if (action === "probe-subscription") {
    await probeSubscription(button);
    return;
  }
  if (
    action === "revoke" &&
    !(await confirmDanger({
      title: "停用 Token",
      message: "伙伴将无法继续使用对应订阅和网关访问。停用后仍可在状态控制里恢复。",
      confirmLabel: "停用",
    }))
  ) {
    setStatus("已取消停用", "info");
    return;
  }
  if (
    action === "rotate-subscription" &&
    !(await confirmDanger({
      title: "重置订阅地址",
      message: "旧订阅地址会立即失效，需要把新地址重新分发给伙伴。",
      confirmLabel: "重置",
    }))
  ) {
    setStatus("已取消重置", "info");
    return;
  }
  if (action === "copy-subscription") {
    button.disabled = true;
    setStatus("复制订阅地址中", "loading");
    try {
      await copyText(button.dataset.tokenUrl || "");
      setStatus("订阅地址已复制", "success");
      showCopyFeedback(button, "已复制", "复制");
    } catch (error) {
      setStatus("复制失败", "danger");
    } finally {
      button.disabled = false;
    }
    return;
  }
  const card = button.closest(".token-card") || button;
  const pendingStatus =
    {
      extend: "续期 Token 中",
      quota: "追加额度中",
      revoke: "停用 Token 中",
      restore: "恢复 Token 中",
      "rotate-subscription": "重置订阅地址中",
    }[action] || "更新 Token 中";
  const failureStatus =
    {
      extend: "续期失败",
      quota: "追加额度失败",
      revoke: "停用失败",
      restore: "恢复失败",
      "rotate-subscription": "重置失败",
    }[action] || "更新失败";
  await runInlineAction(card, button, pendingStatus, async () => {
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
  }, failureStatus);
}

async function probeSubscription(button) {
  const url = button.dataset.tokenUrl || "";
  if (!url) {
    setStatus("缺少订阅地址", "warning");
    return;
  }
  button.disabled = true;
  button.innerHTML = buttonLabel("…", "探测中");
  setStatus("探测订阅中", "loading");
  try {
    const result = await postJSON("/api/subscription/probe", { url });
    updateSubscriptionProbeMeta(button, result);
    if (result.available) {
      button.innerHTML = buttonLabel("✓", "可访问");
      button.classList.add("is-copied");
      setStatus(`订阅可访问：${formatBytes(result.body_bytes || 0)}`, "success");
    } else {
      button.innerHTML = buttonLabel("!", "不可用");
      button.classList.remove("is-copied");
      setStatus(`订阅不可用：${result.message || "请打开链接确认"}`, "danger");
    }
  } catch (error) {
    button.innerHTML = buttonLabel("!", "失败");
    button.classList.remove("is-copied");
    setStatus("订阅探测失败", "danger");
  } finally {
    button.disabled = false;
  }
}

function updateSubscriptionProbeMeta(button, result) {
  const item = button.closest(".token-subscription-item");
  const meta = item?.querySelector(".token-subscription-meta");
  if (!item || !meta) return;
  let probe = item.querySelector("[data-token-subscription-probed-at]");
  if (!probe) {
    probe = document.createElement("span");
    probe.className = "token-subscription-probe-result";
    probe.setAttribute("data-token-subscription-probed-at", "");
    meta.appendChild(probe);
  }
  const checkedAt = formatDateTimeForDisplay(result.checked_at) || "刚刚";
  const size = result.body_bytes ? ` · ${formatBytes(result.body_bytes)}` : "";
  probe.dataset.status = result.available ? "success" : "danger";
  probe.textContent = `${result.available ? "可访问" : "不可用"} · ${checkedAt}${size}`;
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

async function checkDeliveryReadiness() {
  await runOpsAction(deliveryReadinessEl, "检查交付收口中", async () => {
    try {
      const result = await getJSON("/api/delivery/readiness");
      const config = result.config || {};
      const nextAction = Array.isArray(result.next_actions) && result.next_actions.length > 0 ? result.next_actions[0] : "--";
      showConfigResult(result.ready ? "交付可真实测试" : "交付待补齐", result.ready ? "success" : "warning", [
        ["闭环", `${formatCell(result.ready_count)}/${formatCell(result.total_checks)}`],
        ["配置 Hash", escapeHTML(String(config.config_hash || "").slice(0, 12) || "--"), true],
        ["入站", formatCell(config.inbound_count)],
        ["上游", formatCell(config.upstream_outbound_count)],
        ["用户", formatCell(config.user_count)],
        ["下一步", escapeHTML(nextAction), true],
      ]);
      setStatus(result.ready ? "交付闭环可测" : "交付闭环待补", result.ready ? "success" : "warning");
    } catch (error) {
      showConfigResult("收口检查失败", "danger", [["错误", escapeHTML(error.message), true]]);
      setStatus("收口检查失败", "danger");
    }
  });
}

async function checkConfig() {
  await runOpsAction(configCheckEl, "检查配置中", async () => {
    try {
      const result = await postJSON("/api/sing-box/config/check", {});
      showConfigResult(result.valid ? "检查通过" : "检查失败", result.valid ? "success" : "danger", [
        ["配置 Hash", escapeHTML(String(result.config_hash || "").slice(0, 12)), true],
        ["入站", formatCell(result.inbound_count)],
        ["出口", formatCell(result.outbound_count)],
        ["上游", formatCell(result.upstream_outbound_count)],
        ["用户", formatCell(result.user_count)],
      ]);
      setStatus(result.valid ? "配置可用" : "配置异常", result.valid ? "success" : "danger");
    } catch (error) {
      showConfigResult("检查失败", "danger", [["错误", escapeHTML(error.message), true]]);
      setStatus("配置异常", "danger");
    }
  });
}

async function publishConfig() {
  await runOpsAction(configPublishEl, "发布配置中", async () => {
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
      setStatus(
        result.published && !result.restart_required ? "配置已发布并生效" : result.published ? "配置已发布，需重启 sing-box" : "发布失败",
        result.published ? "success" : "danger",
      );
    } catch (error) {
      showConfigResult("发布失败", "danger", [["错误", escapeHTML(error.message), true]]);
      setStatus("发布失败", "danger");
    }
  });
}

async function rollbackConfig() {
  await runOpsAction(configRollbackEl, "回滚配置中", async () => {
    try {
      const result = await postJSON("/api/sing-box/config/rollback", {});
      const restartText = result.restart ? formatRestartResult(result.restart) : result.restart_required ? "需要" : "无需";
      showConfigResult(result.rolled_back ? "回滚完成" : "回滚失败", result.rolled_back ? "success" : "danger", [
        ["配置 Hash", escapeHTML(String(result.config_hash || "").slice(0, 12)), true],
        ["重启", escapeHTML(restartText)],
        ["出口", formatCell(result.outbound_count)],
        ["用户", formatCell(result.user_count)],
      ]);
      setStatus(
        result.rolled_back && !result.restart_required ? "配置已回滚并生效" : result.rolled_back ? "配置已回滚，需重启 sing-box" : "回滚失败",
        result.rolled_back ? "success" : "danger",
      );
    } catch (error) {
      showConfigResult("回滚失败", "danger", [["错误", escapeHTML(error.message), true]]);
      setStatus("回滚失败", "danger");
    }
  });
}

async function restartSingBox() {
  await runOpsAction(configRestartEl, "重启服务中", async () => {
    try {
      const result = await postJSON("/api/sing-box/restart", {});
      showConfigResult(result.success ? "重启已执行" : result.skipped ? "重启未启用" : "重启失败", result.success ? "success" : result.skipped ? "warning" : "danger", [
        ["启用", result.enabled ? "true" : "false"],
        ["已执行", result.executed ? "true" : "false"],
        ["耗时", formatCell(result.duration_ms)],
        ["消息", escapeHTML(result.message || "--"), true],
      ]);
      setStatus(
        result.success ? "服务已重启" : result.skipped ? "重启未启用" : "重启失败",
        result.success ? "success" : result.skipped ? "warning" : "danger",
      );
    } catch (error) {
      showConfigResult("重启失败", "danger", [["错误", escapeHTML(error.message), true]]);
      setStatus("重启失败", "danger");
    }
  });
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
  try {
    const result = await postJSON(path, payload);
    await load();
    return result;
  } catch (error) {
    setStatus("保存失败", "danger");
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

function renderOverviewHero(data) {
  if (!overviewHeroEl) return;
  const checks = overviewReadinessChecks(data);
  const readyCount = checks.filter(([, ready]) => ready).length;
  const firstMissingIndex = checks.findIndex(([, ready]) => !ready);
  const firstMissing = checks[firstMissingIndex] || null;
  const remainingCount = checks.length - readyCount;
  const tokens = appState.tokens || [];
  const nodes = appState.nodes || [];
  const virtualNodes = appState.virtualNodes || [];
  const activeTokens = countBy(tokens, (row) => row.status === "active");
  const activeNodes = countBy(nodes, (row) => row.status === "active");
  const activeVirtualNodes = countBy(virtualNodes, (row) => row.status === "active");
  const ready = remainingCount === 0;
  const action = firstMissing
    ? {
        label: `去${dashboardViewMeta[firstMissing[3]]?.[0] || "处理"}`,
        view: firstMissing[3],
        target: firstMissing[6] || "",
        expand: !!firstMissing[7],
      }
    : {
        label: "发布配置",
        view: "ops",
        target: "config-publish",
        expand: false,
      };
  const title = ready ? "现在可以真实测试" : `还差 ${remainingCount} 步可真实测试`;
  const copy = ready
    ? "基础数据已经形成闭环，下一步发布网关配置后复制订阅到客户端验证。"
    : `先补齐「${firstMissing?.[0] || "待办"}」，后台会保留你的上下文并直接定位到对应操作。`;
  const stats = [
    overviewHeroStat("订", "订阅", activeTokens > 0 ? `${formatPlainNumber(activeTokens)} 个可用` : "待签发", activeTokens > 0),
    overviewHeroStat("点", "节点池", activeNodes > 0 ? `${formatPlainNumber(activeNodes)} 个可用` : "待同步", activeNodes > 0),
    overviewHeroStat("网", "网关", activeVirtualNodes > 0 ? `${formatPlainNumber(activeVirtualNodes)} 个入口` : "待创建", activeVirtualNodes > 0),
  ];

  overviewHeroEl.innerHTML = `
    <div class="overview-hero-main">
      <span class="overview-hero-kicker">${ready ? "交付闭环" : `初始化 ${readyCount}/${checks.length}`}</span>
      <span class="overview-hero-title">
        <span class="overview-hero-status ${ready ? "is-ready" : "is-warning"}" aria-hidden="true"></span>
        <strong>${escapeHTML(title)}</strong>
      </span>
      <span class="overview-hero-copy">${escapeHTML(copy)}</span>
    </div>
    <div class="overview-hero-side">
      <div class="overview-hero-stats" aria-label="核心状态">
        ${stats.join("")}
      </div>
      <button class="primary-link-button overview-hero-action" type="button" data-overview-hero-action data-overview-jump="${escapeHTML(action.view)}" data-overview-jump-target="${escapeHTML(action.target)}" data-overview-jump-expand="${action.expand ? "true" : "false"}">
        ${buttonLabel(ready ? "发" : "→", action.label)}
      </button>
    </div>
  `;
}

function overviewHeroStat(symbol, label, value, ready) {
  return `
    <span class="overview-hero-stat ${ready ? "is-ready" : "is-warning"}">
      <span class="overview-hero-stat-symbol" aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="overview-hero-stat-copy">
        <small>${escapeHTML(label)}</small>
        <strong>${escapeHTML(value)}</strong>
      </span>
    </span>
  `;
}

function renderMetrics(data) {
  const checks = overviewReadinessChecks(data);
  const readyCount = checks.filter(([, ready]) => ready).length;
  const sources = appState.sources || [];
  const nodes = appState.nodes || [];
  const virtualNodes = appState.virtualNodes || [];
  const policies = appState.policies || [];
  const tokens = appState.tokens || [];
  const activeNodes = countBy(nodes, (row) => row.status === "active");
  const healthySources = countBy(sources, (row) => !row.last_error);
  const activeTokens = countBy(tokens, (row) => row.status === "active");
  const activeVirtualNodes = countBy(virtualNodes, (row) => row.status === "active");
  const activePolicies = countBy(policies, (row) => row.status === "active");
  const items = [
    {
      key: "readiness",
      symbol: "测",
      label: "真实闭环",
      value: `${formatPlainNumber(readyCount)}/${formatPlainNumber(checks.length)}`,
      detail: readyCount === checks.length ? "可以真实测试" : "仍有待补齐项",
      tone: readyCount === checks.length ? "success" : "warning",
    },
    {
      key: "nodes",
      symbol: "点",
      label: "可用节点",
      value: `${formatPlainNumber(activeNodes)}/${formatPlainNumber(nodes.length || data.nodes || 0)}`,
      detail: `${formatPlainNumber(groupNodesByRegion(nodes).length)} 个地区`,
      tone: activeNodes > 0 ? "success" : "warning",
    },
    {
      key: "sources",
      symbol: "源",
      label: "上游来源",
      value: `${formatPlainNumber(healthySources)}/${formatPlainNumber(sources.length || data.sources || 0)}`,
      detail: `${formatPlainNumber(countBy(sources, (row) => row.last_error))} 个异常`,
      tone: sources.length > 0 && healthySources === sources.length ? "success" : "warning",
    },
    {
      key: "tokens",
      symbol: "订",
      label: "可用订阅",
      value: `${formatPlainNumber(activeTokens)}/${formatPlainNumber(tokens.length || data.tokens || 0)}`,
      detail: "通用 / Mihomo / sing-box",
      tone: activeTokens > 0 ? "success" : "warning",
    },
    {
      key: "virtualNodes",
      symbol: "网",
      label: "网关入口",
      value: `${formatPlainNumber(activeVirtualNodes)}/${formatPlainNumber(virtualNodes.length || data.virtual_nodes || 0)}`,
      detail: "sing-box 入站",
      tone: activeVirtualNodes > 0 ? "success" : "warning",
    },
    {
      key: "policies",
      symbol: "策",
      label: "访问策略",
      value: `${formatPlainNumber(activePolicies)}/${formatPlainNumber(policies.length || data.policies || 0)}`,
      detail: "分发边界",
      tone: activePolicies > 0 ? "success" : "warning",
    },
  ];
  metricsEl.innerHTML = items.map(overviewMetricCard).join("");
}

function overviewMetricCard(item) {
  return `
    <div class="metric metric-${escapeHTML(item.tone || "neutral")}" data-metric-key="${escapeHTML(item.key)}" data-metric-tone="${escapeHTML(item.tone || "neutral")}">
      <span class="metric-heading">
        <span class="metric-symbol">${escapeHTML(item.symbol)}</span>
        <span class="metric-label">${escapeHTML(item.label)}</span>
      </span>
      <strong class="metric-value">${escapeHTML(String(item.value ?? 0))}</strong>
      <span class="metric-detail">${escapeHTML(item.detail || "")}</span>
    </div>
  `;
}

function renderOverviewInsights(data) {
  if (!overviewInsightsEl) return;
  overviewInsightsEl.innerHTML = [
    renderOverviewSourceInsight(data),
    renderOverviewRegionInsight(data),
    renderOverviewDeliveryInsight(data),
  ].join("");
}

function renderOverviewSourceInsight() {
  const sources = appState.sources || [];
  const healthyCount = countBy(sources, (row) => !row.last_error);
  const erroredCount = countBy(sources, (row) => row.last_error);
  const syncingCount = countBy(sources, (row) => row.type === "subscription");
  const recentSources = [...sources]
    .sort((left, right) => overviewTimeSortValue(right.last_sync_at || right.updated_at) - overviewTimeSortValue(left.last_sync_at || left.updated_at))
    .slice(0, 3);
  const rows = recentSources.length
    ? recentSources.map((row) => overviewSyncRow(row)).join("")
    : overviewEmptyInsightRow("暂无上游来源", "先添加订阅源或导入节点");
  return `
    <article class="overview-insight-panel" data-overview-insight="source-health">
      ${overviewInsightHeading("源", "上游健康", `${formatPlainNumber(healthyCount)}/${formatPlainNumber(sources.length)} 正常`)}
      <div class="overview-health-pills" aria-label="上游健康摘要">
        ${overviewHealthPill("正常", healthyCount, healthyCount > 0 ? "success" : "muted")}
        ${overviewHealthPill("订阅源", syncingCount, syncingCount > 0 ? "info" : "muted")}
        ${overviewHealthPill("异常", erroredCount, erroredCount > 0 ? "warning" : "success")}
      </div>
      <div class="overview-sync-list" aria-label="最近同步">
        ${rows}
      </div>
    </article>
  `;
}

function renderOverviewRegionInsight() {
  const nodes = appState.nodes || [];
  const groups = groupNodesByRegion(nodes)
    .map((group) => ({
      ...group,
      activeCount: countBy(group.items, (row) => row.status === "active"),
    }))
    .sort((left, right) => right.items.length - left.items.length || left.region.localeCompare(right.region, "zh-CN"));
  const visibleGroups = groups.slice(0, 5);
  const rows = visibleGroups.length
    ? visibleGroups.map((group) => overviewDistributionRow(group, nodes.length)).join("")
    : overviewEmptyInsightRow("暂无地区数据", "同步节点后会按地区聚合");
  return `
    <article class="overview-insight-panel" data-overview-insight="region-distribution">
      ${overviewInsightHeading("区", "地区分布", `${formatPlainNumber(groups.length)} 个地区`)}
      <div class="overview-distribution-list" aria-label="地区节点分布">
        ${rows}
      </div>
    </article>
  `;
}

function renderOverviewDeliveryInsight() {
  const delivery = appState.deliveryReadiness || {};
  const tokens = appState.tokens || [];
  const virtualNodes = appState.virtualNodes || [];
  const policies = appState.policies || [];
  const activeTokens = countBy(tokens, (row) => row.status === "active");
  const activeVirtualNodes = countBy(virtualNodes, (row) => row.status === "active");
  const activePolicies = countBy(policies, (row) => row.status === "active");
  const publishSummary =
    Number.isFinite(Number(delivery.ready_count)) && Number.isFinite(Number(delivery.total_checks))
      ? `${formatPlainNumber(delivery.ready_count)}/${formatPlainNumber(delivery.total_checks)} 项`
      : "待检查";
  return `
    <article class="overview-insight-panel" data-overview-insight="delivery">
      ${overviewInsightHeading("发", "分发与发布", delivery.ready ? "可发布" : publishSummary)}
      <div class="overview-delivery-list" aria-label="分发发布摘要">
        ${overviewDeliveryRow("订", "有效订阅", `${formatPlainNumber(activeTokens)}/${formatPlainNumber(tokens.length)}`, activeTokens > 0 ? "success" : "warning")}
        ${overviewDeliveryRow("网", "网关入口", `${formatPlainNumber(activeVirtualNodes)}/${formatPlainNumber(virtualNodes.length)}`, activeVirtualNodes > 0 ? "success" : "warning")}
        ${overviewDeliveryRow("策", "策略生效", `${formatPlainNumber(activePolicies)}/${formatPlainNumber(policies.length)}`, activePolicies > 0 ? "success" : "warning")}
        ${overviewDeliveryRow("检", "发布检查", delivery.ready ? "就绪" : publishSummary, delivery.ready ? "success" : "warning")}
      </div>
    </article>
  `;
}

function overviewInsightHeading(symbol, title, meta) {
  return `
    <div class="overview-insight-heading">
      <span class="overview-insight-symbol" aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="overview-insight-title">
        <strong>${escapeHTML(title)}</strong>
        <small>${escapeHTML(meta || "")}</small>
      </span>
    </div>
  `;
}

function overviewHealthPill(label, value, tone = "muted") {
  return `
    <span class="overview-health-pill overview-health-pill-${escapeHTML(tone)}">
      <span>${escapeHTML(label)}</span>
      <strong>${formatPlainNumber(value)}</strong>
    </span>
  `;
}

function overviewSyncRow(row) {
  const time = formatDateTimeForDisplay(row.last_sync_at) || "未同步";
  const tone = row.last_error ? "warning" : row.last_sync_at ? "success" : "muted";
  const status = row.last_error ? "异常" : row.last_sync_at ? "已同步" : "待同步";
  return `
    <div class="overview-sync-row overview-sync-row-${tone}">
      <span class="overview-sync-name">${escapeHTML(row.name || `来源 ${row.id || ""}`)}</span>
      <span class="overview-sync-meta">${escapeHTML(status)} · ${escapeHTML(time)}</span>
    </div>
  `;
}

function overviewDistributionRow(group, totalCount) {
  const count = group.items.length;
  const percent = totalCount > 0 ? Math.max(4, Math.round((count / totalCount) * 100)) : 0;
  return `
    <div class="overview-distribution-row">
      <span class="overview-distribution-label">${escapeHTML(group.region)}</span>
      <span class="overview-distribution-bar" aria-hidden="true"><span style="width: ${percent}%"></span></span>
      <span class="overview-distribution-value">${formatPlainNumber(group.activeCount)}/${formatPlainNumber(count)}</span>
    </div>
  `;
}

function overviewDeliveryRow(symbol, label, value, tone = "muted") {
  return `
    <div class="overview-delivery-row overview-delivery-row-${escapeHTML(tone)}">
      <span class="overview-delivery-symbol" aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="overview-delivery-label">${escapeHTML(label)}</span>
      <strong>${escapeHTML(value)}</strong>
    </div>
  `;
}

function overviewEmptyInsightRow(title, detail) {
  return `
    <div class="overview-sync-row overview-sync-row-muted">
      <span class="overview-sync-name">${escapeHTML(title)}</span>
      <span class="overview-sync-meta">${escapeHTML(detail)}</span>
    </div>
  `;
}

function overviewTimeSortValue(value) {
  const date = parseDisplayTime(value);
  return date ? date.getTime() : 0;
}

function overviewReadinessChecks(data) {
  const delivery = appState.deliveryReadiness || {};
  const deliveryChecks = Array.isArray(delivery.checks) ? delivery.checks : [];
  const deliveryByKey = Object.fromEntries(deliveryChecks.map((check) => [check.key, check]));
  const sourceCount = Number(data.sources || 0);
  const nodeCount = Number(data.nodes || 0);
  const virtualNodeCount = Number(data.virtual_nodes || 0);
  const tokenCount = Number(data.tokens || 0);
  const policyCount = Number(data.policies || 0);
  const activeTokenCount = countBy(appState.tokens || [], (row) => row.status === "active");
  const usableTokenCount = Number(deliveryByKey.tokens?.summary?.match(/^\d+/)?.[0] || activeTokenCount || tokenCount || 0);
  const subscriptionReady = usableTokenCount > 0;
  const publishReady = delivery.ready === true;
  const publishSummary =
    Number.isFinite(Number(delivery.ready_count)) && Number.isFinite(Number(delivery.total_checks))
      ? `${formatPlainNumber(delivery.ready_count)}/${formatPlainNumber(delivery.total_checks)} 项`
      : "待检查";
  return [
    ["接入来源", sourceCount > 0, `${sourceCount} 个来源`, "access", "源", "先添加或导入机场订阅，让节点池有可用上游。", "source-form", true],
    ["节点池", nodeCount > 0, `${nodeCount} 个节点`, "nodes", "点", "检查地区聚合、节点详情和命名，确保伙伴能看懂节点来源。", "nodes", false],
    ["虚拟网关", virtualNodeCount > 0, `${virtualNodeCount} 个入口`, "nodes", "网", "创建 sing-box 入站入口，把节点池组合成可分发网关。", "virtual-node-form", true],
    ["团队 Token", tokenCount > 0, `${tokenCount} 个 Token`, "identity", "身", "为团队伙伴创建 Token，绑定团队伙伴的有效期和额度。", "token-form", true],
    ["订阅分发", subscriptionReady, `${usableTokenCount} 个可用订阅`, "identity", "订", "复制默认、Mihomo 或 sing-box 订阅地址，直接导入客户端测试。", "tokens", false],
    ["访问策略", policyCount > 0, `${policyCount} 条策略`, "policies", "策", "给团队或 Token 绑定可见范围和节点上限，避免误分发。", "policy-form", true],
    ["发布检查", publishReady, publishSummary, "ops", "发", "在运维模块检查交付收口并发布 sing-box 配置。", "ops-actions", false],
  ];
}

function renderOverviewReadiness(data) {
  const checks = overviewReadinessChecks(data);
  const readyCount = checks.filter(([, ready]) => ready).length;
  const firstMissingIndex = checks.findIndex(([, ready]) => !ready);
  const firstMissing = checks.find(([, ready]) => !ready);
  const progressPercent = Math.round((readyCount / checks.length) * 100);
  const next = firstMissing
    ? {
        title: `补齐${firstMissing[0]}`,
        description: "补齐后再回到运维模块检查并发布 sing-box 配置，订阅地址才更接近真实客户端体验。",
        view: firstMissing[3],
        target: firstMissing[6] || "",
        expand: !!firstMissing[7],
        action: `去${dashboardViewMeta[firstMissing[3]]?.[0] || "处理"}`,
      }
    : {
        title: "闭环已具备真实测试条件",
        description: "基础数据已齐备，可以发布网关配置，再复制 Clash/Mihomo 或 sing-box 订阅地址做客户端导入验证。",
        view: "ops",
        target: "ops-actions",
        expand: false,
        action: "去运维发布",
      };
  const guideCurrent = firstMissing
    ? {
        step: `第 ${firstMissingIndex + 1} 步`,
        title: next.title,
        description: firstMissing[5],
        view: next.view,
        target: next.target,
        expand: next.expand,
        action: next.action,
      }
    : {
        step: "闭环完成",
        title: "发布配置并导入客户端",
        description: "基础数据已齐备，现在可以发布网关配置，然后复制订阅地址到 Mihomo 或 sing-box 真实验证。",
        view: "ops",
        target: "ops-actions",
        expand: false,
        action: "去运维发布",
      };
  const guideActions = [
    {
      symbol: "订",
      title: "去复制订阅",
      hint: "默认 / Mihomo / sing-box",
      view: "identity",
      target: "tokens",
      expand: false,
    },
    {
      symbol: "收",
      title: "检查收口",
      hint: "确认真实可测",
      view: "ops",
      target: "delivery-readiness",
      expand: false,
    },
    {
      symbol: "发",
      title: "发布配置",
      hint: "写入网关配置",
      view: "ops",
      target: "config-publish",
      expand: false,
    },
  ];

  overviewReadinessEl.innerHTML = `
    <div class="overview-panel-heading">
      <span>初始化向导</span>
      <strong>${readyCount}/${checks.length}</strong>
    </div>
    <div class="guide-summary" data-guide-summary>
      <div class="guide-progress">
        <span class="guide-progress-label">可测进度</span>
        <strong data-guide-progress>${readyCount}/${checks.length}</strong>
        <span class="guide-progress-bar" aria-hidden="true">
          <span style="width: ${progressPercent}%"></span>
        </span>
      </div>
      <div class="guide-current" data-guide-current>
        <span class="guide-current-symbol" aria-hidden="true">${escapeHTML(dashboardViewSymbols[guideCurrent.view] || "→")}</span>
        <span class="guide-current-copy">
          <small>${escapeHTML(guideCurrent.step)}</small>
          <strong>${escapeHTML(guideCurrent.title)}</strong>
          <span>${escapeHTML(guideCurrent.description)}</span>
        </span>
        <button class="primary-link-button guide-current-action" type="button" data-overview-jump="${guideCurrent.view}" data-overview-jump-target="${escapeHTML(guideCurrent.target || "")}" data-overview-jump-expand="${guideCurrent.expand ? "true" : "false"}">
          ${buttonLabel("→", guideCurrent.action)}
        </button>
      </div>
      <div class="guide-action-strip" data-guide-action-strip aria-label="真实测试动作">
        ${guideActions
          .map(
            (action) => `
              <button class="guide-action-button" type="button" data-guide-action data-overview-jump="${escapeHTML(action.view)}" data-overview-jump-target="${escapeHTML(action.target)}" data-overview-jump-expand="${action.expand ? "true" : "false"}">
                <span class="guide-action-symbol" aria-hidden="true">${escapeHTML(action.symbol)}</span>
                <span class="guide-action-copy">
                  <strong>${escapeHTML(action.title)}</strong>
                  <small>${escapeHTML(action.hint)}</small>
                </span>
              </button>
            `,
          )
          .join("")}
      </div>
    </div>
    <div class="readiness-list" data-overview-readiness>
      ${checks
        .map(
          ([label, ready, summary, view, symbol, hint, target, expand], index) => `
            <button class="readiness-item ${ready ? "is-ready" : ""}" type="button" data-overview-jump="${view}" data-overview-jump-target="${escapeHTML(target || "")}" data-overview-jump-expand="${expand ? "true" : "false"}">
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
                <span class="readiness-hint">${escapeHTML(ready ? "已具备，点击可复核配置。" : hint)}</span>
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
      <button class="primary-link-button next-step-action" type="button" data-overview-jump="${next.view}" data-overview-jump-target="${escapeHTML(next.target || "")}" data-overview-jump-expand="${next.expand ? "true" : "false"}">
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

function updateWorkspaceInsight() {
  if (!workspaceInsightEl) return;
  const insight = workspaceInsightForView(activeDashboardView);
  workspaceInsightEl.hidden = !insight;
  if (!insight) {
    workspaceInsightEl.replaceChildren();
    return;
  }
  workspaceInsightEl.dataset.tone = insight.tone || "info";
  workspaceInsightEl.innerHTML = `
    <span class="workspace-insight-symbol" aria-hidden="true">${escapeHTML(insight.symbol || "现")}</span>
    <span class="workspace-insight-copy">
      <span class="workspace-insight-kicker">${escapeHTML(insight.kicker || "当前重点")}</span>
      <strong class="workspace-insight-title">${escapeHTML(insight.title || "保持可测试")}</strong>
      <span class="workspace-insight-detail">${escapeHTML(insight.detail || "")}</span>
    </span>
    ${renderWorkspaceInsightAction(insight.action)}
  `;
}

function renderWorkspaceInsightAction(action) {
  if (!action) return "";
  return `
    <button
      class="workspace-insight-action"
      type="button"
      data-workspace-insight-action
      data-workspace-insight-view="${escapeHTML(action.view || activeDashboardView)}"
      data-workspace-insight-target="${escapeHTML(action.target || "")}"
      data-workspace-insight-expand="${action.expand ? "true" : "false"}"
    >
      ${buttonLabel(action.symbol || "→", action.label || "处理")}
    </button>
  `;
}

function workspaceInsightForView(view) {
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
  const activeVirtualNodes = countBy(virtualNodes, (row) => row.status === "active");
  const activeTokens = countBy(tokens, (row) => row.status === "active");
  const activePolicies = countBy(policies, (row) => row.status === "active");

  if (view === "overview") {
    const checks = overviewReadinessChecks(overview);
    const readyCount = checks.filter(([, ready]) => ready).length;
    const firstMissing = checks.find(([, ready]) => !ready);
    const action = firstMissing
      ? { label: `补齐${firstMissing[0]}`, symbol: firstMissing[4] || "步", view: firstMissing[3], target: firstMissing[6] || "", expand: !!firstMissing[7] }
      : { label: "发布配置", symbol: "发", view: "ops", target: "config-publish" };
    return {
      tone: firstMissing ? "warning" : "success",
      symbol: firstMissing ? "待" : "测",
      kicker: "真实测试闭环",
      title: firstMissing ? `还差 ${formatPlainNumber(checks.length - readyCount)} 步` : "现在可以真实测试",
      detail: firstMissing
        ? `闭环 ${formatPlainNumber(readyCount)}/${formatPlainNumber(checks.length)}，优先处理：${firstMissing[0]}。`
        : "来源、节点、网关、Token、策略和发布检查都已就绪。",
      action,
    };
  }

  if (view === "access") {
    const subscriptionSources = countBy(sources, (row) => row.type === "subscription");
    const erroredSources = countBy(sources, (row) => row.last_error);
    if (sources.length === 0) {
      return {
        tone: "warning",
        symbol: "源",
        kicker: "接入状态",
        title: "还没有上游来源",
        detail: "先添加订阅来源或导入节点，后续分发才有真实节点池。",
        action: { label: "添加来源", symbol: "源", view: "access", target: "source-form", expand: true },
      };
    }
    return {
      tone: erroredSources > 0 ? "warning" : "success",
      symbol: erroredSources > 0 ? "警" : "同",
      kicker: "接入状态",
      title: erroredSources > 0 ? `${formatPlainNumber(erroredSources)} 个来源异常` : "来源可同步",
      detail: `${formatPlainNumber(sources.length)} 个来源，${formatPlainNumber(subscriptionSources)} 个订阅源，刷新后会进入节点池。`,
      action: { label: erroredSources > 0 ? "查看来源" : "添加来源", symbol: erroredSources > 0 ? "查" : "源", view: "access", target: "sources", expand: false },
    };
  }

  if (view === "nodes") {
    const regions = groupNodesByRegion(nodes).length;
    if (nodes.length === 0) {
      return {
        tone: "warning",
        symbol: "点",
        kicker: "节点池",
        title: "节点池为空",
        detail: "先从接入模块同步订阅或导入节点，再按地区检查详情。",
        action: { label: "导入节点", symbol: "导", view: "access", target: "node-import-form", expand: true },
      };
    }
    if (virtualNodes.length === 0) {
      return {
        tone: "warning",
        symbol: "网",
        kicker: "节点池",
        title: "缺少虚拟网关",
        detail: `${formatPlainNumber(regions)} 个地区、${formatPlainNumber(activeNodes)} 个可用节点，创建网关后才能分发。`,
        action: { label: "创建网关", symbol: "网", view: "nodes", target: "virtual-node-form", expand: true },
      };
    }
    return {
      tone: "success",
      symbol: "区",
      kicker: "节点池",
      title: `${formatPlainNumber(regions)} 个地区已聚合`,
      detail: `${formatPlainNumber(nodes.length)} 个节点，${formatPlainNumber(activeNodes)} 个可用，可进入地区查看单节点详情。`,
      action: { label: "查看地区", symbol: "区", view: "nodes", target: "nodes", expand: false },
    };
  }

  if (view === "identity") {
    if (teams.length === 0 || users.length === 0) {
      return {
        tone: "warning",
        symbol: "身",
        kicker: "身份分发",
        title: "团队或成员未齐",
        detail: `${formatPlainNumber(teams.length)} 个团队、${formatPlainNumber(users.length)} 个成员，先补齐归属再签发 Token。`,
        action: { label: teams.length === 0 ? "建团队" : "加成员", symbol: teams.length === 0 ? "团" : "员", view: "identity", target: teams.length === 0 ? "team-form" : "user-form", expand: true },
      };
    }
    if (tokens.length === 0) {
      return {
        tone: "warning",
        symbol: "钥",
        kicker: "身份分发",
        title: "还没有可复制订阅",
        detail: "为成员签发 Token 后，后台会给出通用、Mihomo 和 sing-box 地址。",
        action: { label: "签发 Token", symbol: "钥", view: "identity", target: "token-form", expand: true },
      };
    }
    return {
      tone: activeTokens > 0 ? "success" : "warning",
      symbol: "订",
      kicker: "身份分发",
      title: activeTokens > 0 ? "订阅可分发" : "Token 待启用",
      detail: `${formatPlainNumber(tokens.length)} 个 Token，${formatPlainNumber(activeTokens)} 个有效，可在卡片里反复复制地址。`,
      action: { label: "查看 Token", symbol: "钥", view: "identity", target: "tokens", expand: false },
    };
  }

  if (view === "policies") {
    const constrainedPolicies = countBy(
      policies,
      (row) => row.allowed_virtual_nodes || row.include_tags || row.exclude_tags || Number(row.max_nodes || 0) > 0,
    );
    if (policies.length === 0) {
      return {
        tone: "warning",
        symbol: "策",
        kicker: "访问策略",
        title: "还没有访问边界",
        detail: "创建策略后可以按团队、成员或 Token 控制可见节点范围。",
        action: { label: "创建策略", symbol: "策", view: "policies", target: "policy-form", expand: true },
      };
    }
    return {
      tone: activePolicies > 0 ? "success" : "warning",
      symbol: "限",
      kicker: "访问策略",
      title: `${formatPlainNumber(activePolicies)} 条策略生效`,
      detail: `${formatPlainNumber(constrainedPolicies)} 条带限制，分发前可快速确认可见范围。`,
      action: { label: "查看策略", symbol: "策", view: "policies", target: "policies", expand: false },
    };
  }

  if (view === "traffic") {
    const sampleCount = trafficHourly.length + trafficDaily.length + trafficOutbounds.length + trafficTokens.length;
    if (sampleCount === 0) {
      return {
        tone: "warning",
        symbol: "量",
        kicker: "流量统计",
        title: "暂无流量样本",
        detail: "发布配置并让客户端连入网关后，Token、出口和小时统计会开始出现。",
        action: { label: "发布配置", symbol: "发", view: "ops", target: "config-publish", expand: false },
      };
    }
    return {
      tone: "success",
      symbol: "量",
      kicker: "流量统计",
      title: "已有统计样本",
      detail: `${formatPlainNumber(trafficTokens.length)} 个 Token 用量，${formatPlainNumber(trafficOutbounds.length)} 个出口摘要。`,
      action: { label: "看 Token", symbol: "钥", view: "traffic", target: "traffic-tokens", expand: false },
    };
  }

  if (view === "ops") {
    const ready = activeNodes > 0 && activeVirtualNodes > 0 && activeTokens > 0;
    return {
      tone: ready ? "success" : "warning",
      symbol: ready ? "发" : "检",
      kicker: "发布状态",
      title: ready ? "具备发布条件" : "发布前待补齐",
      detail: `${formatPlainNumber(activeNodes)} 可用节点，${formatPlainNumber(activeVirtualNodes)} 个网关，${formatPlainNumber(activeTokens)} 个有效 Token。`,
      action: { label: ready ? "发布配置" : "检查收口", symbol: ready ? "发" : "收", view: "ops", target: ready ? "config-publish" : "ops-actions", expand: false },
    };
  }

  return null;
}

function updateViewRail() {
  if (!viewRailEl) return;
  const items = dashboardViewRail[activeDashboardView] || [];
  viewRailEl.hidden = items.length === 0;
  viewRailEl.innerHTML = items.map(renderViewRailButton).join("");
}

function updateWorkspacePrimaryAction() {
  if (!viewPrimaryActionEl) return;
  const action = workspacePrimaryActionForView(activeDashboardView);
  viewPrimaryActionEl.hidden = !action;
  if (!action) return;
  viewPrimaryActionEl.innerHTML = buttonLabel(action.symbol || "→", action.label || "下一步");
  viewPrimaryActionEl.dataset.primaryActionView = action.view || activeDashboardView;
  viewPrimaryActionEl.dataset.primaryActionTarget = action.target || "";
  viewPrimaryActionEl.dataset.primaryActionExpand = action.expand ? "true" : "false";
  viewPrimaryActionEl.setAttribute(
    "aria-label",
    `${dashboardViewMeta[activeDashboardView]?.[0] || "当前模块"}：${action.label || "下一步"}`,
  );
}

function workspacePrimaryActionForView(view) {
  if (view === "overview") return overviewPrimaryAction();
  const actions = {
    access: { label: "添加来源", symbol: "源", view: "access", target: "source-form", expand: true },
    nodes: { label: "创建网关", symbol: "网", view: "nodes", target: "virtual-node-form", expand: true },
    identity: { label: "签发 Token", symbol: "钥", view: "identity", target: "token-form", expand: true },
    policies: { label: "创建策略", symbol: "策", view: "policies", target: "policy-form", expand: true },
    traffic: { label: "查看运维", symbol: "运", view: "ops", target: "ops-actions" },
    ops: { label: "检查收口", symbol: "收", view: "ops", target: "ops-actions" },
  };
  return actions[view] || null;
}

function overviewPrimaryAction() {
  const firstMissing = overviewReadinessChecks(appState.overview || {}).find(([, ready]) => !ready);
  if (!firstMissing) {
    return { label: "去运维发布", symbol: "运", view: "ops", target: "ops-actions" };
  }
  const actions = {
    "接入来源": { label: "添加来源", symbol: "源", view: "access", target: "source-form", expand: true },
    "节点池": { label: "导入节点", symbol: "导", view: "access", target: "node-import-form", expand: true },
    "虚拟网关": { label: "创建网关", symbol: "网", view: "nodes", target: "virtual-node-form", expand: true },
    "团队 Token": { label: "签发 Token", symbol: "钥", view: "identity", target: "token-form", expand: true },
    "访问策略": { label: "创建策略", symbol: "策", view: "policies", target: "policy-form", expand: true },
  };
  return actions[firstMissing[0]] || { label: "继续初始化", symbol: "步", view: firstMissing[3], target: "overview-next-step" };
}

function renderViewRailButton([label, target, expand]) {
  const metric = viewRailMetricForTarget(target, expand);
  const symbol = viewRailSymbolForTarget(label, target, expand);
  const isCurrent = target === activeDashboardTarget;
  return `
    <button class="view-rail-button${isCurrent ? " is-current" : ""}" type="button" data-view-rail-target="${escapeHTML(target)}" data-view-rail-expand="${expand ? "true" : "false"}" aria-current="${isCurrent ? "true" : "false"}">
      <span class="view-rail-symbol" aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="view-rail-label">${escapeHTML(label)}</span>
      <span class="view-rail-count" data-view-rail-count>${escapeHTML(metric)}</span>
    </button>
  `;
}

function viewRailSymbolForTarget(label, target, expand) {
  const symbols = {
    metrics: "概",
    "overview-readiness": "向",
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
    element.classList.toggle("is-empty", value === "0" || /^0\/\d+$/.test(value) || value === "待检");
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

function navigateToDashboardTarget(view, targetID, expand = false) {
  if (view && dashboardViewMeta[view]) {
    setActiveView(view);
  }
  requestAnimationFrame(() => openDashboardTarget(targetID, expand));
}

function openDashboardTarget(targetID, expand = false) {
  const target = document.getElementById(targetID || "");
  if (!target) return;
  activeDashboardTarget = targetID || "";
  updateViewRail();
  if (expand) {
    const drawer = target.closest("[data-form-drawer]");
    if (drawer) {
      setFormDrawerCollapsed(drawer, false);
    }
  }
  const panelTarget = target.closest(".panel") || target;
  highlightDashboardTarget(panelTarget);
  setStatus(`已定位：${dashboardTargetLabel(targetID)}`, "info");
  panelTarget.scrollIntoView({ behavior: "smooth", block: "start", inline: "nearest" });
  focusDashboardTarget(target);
}

function dashboardTargetLabel(targetID) {
  const target = document.getElementById(targetID || "");
  const directTargetLabels = {
    "delivery-readiness": "检查收口",
    "config-publish": "发布配置",
  };
  if (directTargetLabels[targetID]) return directTargetLabels[targetID];
  const railItem = Object.values(dashboardViewRail)
    .flat()
    .find(([, target]) => target === targetID);
  if (railItem?.[0]) return railItem[0];
  const action = workspacePrimaryActionForView(activeDashboardView);
  if (action?.target === targetID && action.label) return action.label;
  const heading = target?.querySelector?.("h2")?.textContent?.trim();
  if (heading) return heading;
  return dashboardViewMeta[activeDashboardView]?.[0] || "当前模块";
}

function highlightDashboardTarget(target) {
  target.classList.remove("is-target-highlighted");
  window.requestAnimationFrame(() => {
    target.classList.add("is-target-highlighted");
    window.setTimeout(() => target.classList.remove("is-target-highlighted"), 1800);
  });
}

function focusDashboardTarget(target) {
  const drawer = target.matches("[data-form-drawer]") ? target : target.closest("[data-form-drawer]");
  const searchRoot = drawer?.querySelector(".form-body") || target;
  const focusTarget = searchRoot.querySelector(
    'input:not([type="hidden"]):not([disabled]), select:not([disabled]), textarea:not([disabled]), button:not([disabled]), a[href]',
  );
  if (focusTarget && typeof focusTarget.focus === "function") {
    focusTarget.focus({ preventScroll: true });
  }
}

function emptyState(symbol, title, hint = "", action = null) {
  return `
    <div class="empty empty-state" data-empty-state>
      <span class="empty-state-symbol" data-empty-state-symbol aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="empty-state-copy">
        <strong class="empty-state-title">${escapeHTML(title)}</strong>
        ${hint ? `<span class="empty-state-hint">${escapeHTML(hint)}</span>` : ""}
      </span>
      ${renderEmptyStateAction(action)}
    </div>
  `;
}

function renderEmptyStateAction(action) {
  if (!action || !action.target) return "";
  const view = action.view || activeDashboardView;
  const label = action.label || "去处理";
  const symbol = action.symbol || "→";
  return `
    <button class="primary-link-button empty-state-action" type="button"
      data-empty-state-action
      data-empty-state-view="${escapeHTML(view)}"
      data-empty-state-target="${escapeHTML(action.target)}"
      data-empty-state-expand="${action.expand ? "true" : "false"}">
      ${buttonLabel(symbol, label)}
    </button>
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
    teamsEl.innerHTML = emptyState("团", "暂无团队", "创建团队后再为成员签发订阅 Token", {
      label: "创建团队",
      symbol: "团",
      view: "identity",
      target: "team-form",
      expand: true,
    });
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
    usersEl.innerHTML = emptyState("员", "暂无成员", "添加成员后可以独立控制 Token 有效期和额度", {
      label: "添加成员",
      symbol: "员",
      view: "identity",
      target: "user-form",
      expand: true,
    });
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
    sourcesEl.innerHTML = emptyState("源", "暂无来源", "添加订阅或手动来源后会在这里统一管理", {
      label: "添加来源",
      symbol: "源",
      view: "access",
      target: "source-form",
      expand: true,
    });
    return;
  }
  sourcesEl.innerHTML = `
    ${renderSourceWorkbench(rows)}
    <div class="source-card-list">
      ${rows.map((row) => renderSourceCard(row)).join("")}
    </div>
  `;
}

function renderSourceWorkbench(rows) {
  const sourceRows = rows || [];
  const healthyCount = countBy(sourceRows, (row) => !row.last_error);
  const subscriptionCount = countBy(sourceRows, (row) => row.type === "subscription");
  const refreshableCount = countBy(sourceRows, (row) => Number(row.refresh_interval_minutes || 0) > 0);
  const errorCount = countBy(sourceRows, (row) => row.last_error);
  const typeGroups = [...sourceRows.reduce((map, row) => {
    const type = String(row.type || "unknown");
    if (!map.has(type)) map.set(type, []);
    map.get(type).push(row);
    return map;
  }, new Map()).entries()]
    .map(([type, items]) => ({ type, items }))
    .sort((left, right) => right.items.length - left.items.length || left.type.localeCompare(right.type, "zh-CN"));
  return `
    <section class="source-workbench" data-source-workbench aria-label="接入来源工作台">
      <div class="source-workbench-heading">
        <span class="source-workbench-symbol" aria-hidden="true">接</span>
        <span class="source-workbench-copy">
          <small class="source-workbench-kicker">接入来源工作台</small>
          <strong class="source-workbench-title">先看同步健康，再处理单个来源</strong>
        </span>
      </div>
      <div class="source-workbench-stats" aria-label="来源健康摘要">
        ${sourceWorkbenchStat("源", "来源总数", `${formatPlainNumber(sourceRows.length)} 个`, `${formatPlainNumber(subscriptionCount)} 个订阅源`)}
        ${sourceWorkbenchStat("健", "健康来源", `${formatPlainNumber(healthyCount)}/${formatPlainNumber(sourceRows.length)}`, "无错误才进入稳定导入")}
        ${sourceWorkbenchStat("刷", "自动刷新", `${formatPlainNumber(refreshableCount)} 个`, "刷新后进入节点池")}
        ${sourceWorkbenchStat("异", "异常来源", `${formatPlainNumber(errorCount)} 个`, errorCount > 0 ? "优先查看错误字段" : "暂无同步异常")}
      </div>
      <div class="source-workbench-type-grid" aria-label="来源类型分布">
        ${typeGroups.map((group) => sourceWorkbenchTypeCard(group)).join("")}
      </div>
    </section>
  `;
}

function sourceWorkbenchStat(symbol, label, value, detail) {
  return `
    <span class="source-workbench-stat" data-source-workbench-stat>
      <span class="source-workbench-stat-symbol" data-source-workbench-symbol aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="source-workbench-stat-copy">
        <span class="source-workbench-stat-label">${escapeHTML(label)}</span>
        <strong class="source-workbench-stat-value">${escapeHTML(String(value))}</strong>
        <small class="source-workbench-stat-detail">${escapeHTML(detail || "")}</small>
      </span>
    </span>
  `;
}

function sourceWorkbenchTypeCard(group) {
  const syncedCount = countBy(group.items, (row) => row.last_sync_at && !row.last_error);
  const refreshableCount = countBy(group.items, (row) => Number(row.refresh_interval_minutes || 0) > 0);
  return `
    <article class="source-workbench-type-card" data-source-workbench-type-card>
      <span class="source-workbench-type-symbol" aria-hidden="true">${escapeHTML(sourceTypeSymbol(group.type))}</span>
      <span class="source-workbench-type-copy">
        <strong class="source-workbench-type-name">${escapeHTML(sourceTypeLabel(group.type))}</strong>
        <small class="source-workbench-type-meta">${formatPlainNumber(group.items.length)} 个来源 · ${formatPlainNumber(syncedCount)} 已同步</small>
      </span>
      <span class="source-workbench-type-hint">${formatPlainNumber(refreshableCount)} 自动刷新</span>
    </article>
  `;
}

function sourceTypeLabel(type) {
  const normalized = String(type || "unknown");
  const labels = {
    subscription: "订阅来源",
    manual: "手动来源",
    url: "URL 来源",
    unknown: "未知来源",
  };
  return labels[normalized] || normalized;
}

function sourceTypeSymbol(type) {
  const normalized = String(type || "unknown");
  const symbols = {
    subscription: "订",
    manual: "手",
    url: "链",
    unknown: "?",
  };
  return symbols[normalized] || normalized.slice(0, 1).toUpperCase() || "源";
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
    virtualNodesEl.innerHTML = emptyState("网", "暂无虚拟网关", "创建网关入口后才能生成可用的 sing-box 入站", {
      label: "创建网关",
      symbol: "网",
      view: "nodes",
      target: "virtual-node-form",
      expand: true,
    });
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
    policiesEl.innerHTML = emptyState("策", "暂无策略", "配置策略后可以限制团队可见节点和网关范围", {
      label: "创建策略",
      symbol: "策",
      view: "policies",
      target: "policy-form",
      expand: true,
    });
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
    nodesEl.innerHTML = emptyState("点", "暂无节点", "导入上游订阅后会按地区自动聚合节点", {
      label: "导入节点",
      symbol: "导",
      view: "access",
      target: "node-import-form",
      expand: true,
    });
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
    ${renderNodeWorkbench(groups, total, matched)}
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

function renderNodeWorkbench(groups, total, matched) {
  const safeGroups = groups || [];
  const nodes = safeGroups.flatMap((group) => group.items || []);
  const activeCount = countBy(nodes, (node) => node.status === "active");
  const protocolCount = new Set(nodes.map((node) => node.protocol).filter(Boolean)).size;
  const sourceCount = new Set(nodes.map((node) => node.source_name || node.source_id).filter(Boolean)).size;
  const topRegions = [...safeGroups]
    .sort((left, right) => right.items.length - left.items.length || left.region.localeCompare(right.region, "zh-CN"))
    .slice(0, 4);
  return `
    <section class="node-workbench" data-node-workbench aria-label="节点池工作台">
      <div class="node-workbench-heading">
        <span class="node-workbench-symbol" aria-hidden="true">池</span>
        <span class="node-workbench-copy">
          <small class="node-workbench-kicker">节点池工作台</small>
          <strong class="node-workbench-title">先看地区承载，再进入节点卡片和详情</strong>
        </span>
      </div>
      <div class="node-workbench-stats" aria-label="节点池摘要">
        ${nodeWorkbenchStat("点", "节点匹配", `${formatPlainNumber(matched)}/${formatPlainNumber(total)}`, "地区、协议、来源都可搜索")}
        ${nodeWorkbenchStat("活", "可用节点", `${formatPlainNumber(activeCount)}/${formatPlainNumber(matched)}`, "active 节点可进入分发")}
        ${nodeWorkbenchStat("区", "地区类别", `${formatPlainNumber(safeGroups.length)} 个`, "无地区信息自动归到其他")}
        ${nodeWorkbenchStat("协", "协议/来源", `${formatPlainNumber(protocolCount)}/${formatPlainNumber(sourceCount)}`, "协议数 / 来源数")}
      </div>
      <div class="node-workbench-region-grid" aria-label="重点地区">
        ${
          topRegions.length > 0
            ? topRegions.map((group) => nodeWorkbenchRegionCard(group, matched)).join("")
            : `<div class="node-workbench-region-empty">当前筛选没有匹配地区</div>`
        }
      </div>
    </section>
  `;
}

function nodeWorkbenchStat(symbol, label, value, detail) {
  return `
    <span class="node-workbench-stat" data-node-workbench-stat>
      <span class="node-workbench-stat-symbol" data-node-workbench-symbol aria-hidden="true">${escapeHTML(symbol)}</span>
      <span class="node-workbench-stat-copy">
        <span class="node-workbench-stat-label">${escapeHTML(label)}</span>
        <strong class="node-workbench-stat-value">${escapeHTML(String(value))}</strong>
        <small class="node-workbench-stat-detail">${escapeHTML(detail || "")}</small>
      </span>
    </span>
  `;
}

function nodeWorkbenchRegionCard(group, matched) {
  const activeCount = countBy(group.items, (node) => node.status === "active");
  const activeRatio = group.items.length > 0 ? Math.round((activeCount / group.items.length) * 100) : 0;
  const shareRatio = matched > 0 ? Math.round((group.items.length / matched) * 100) : 0;
  const barWidth = Math.max(4, Math.min(100, activeRatio));
  return `
    <article class="node-workbench-region-card" data-node-workbench-region-card>
      <span class="node-workbench-region-symbol" aria-hidden="true">${escapeHTML(regionSymbolForRegion(group.region))}</span>
      <span class="node-workbench-region-copy">
        <strong class="node-workbench-region-name">${escapeHTML(group.region)}</strong>
        <small class="node-workbench-region-meta">${formatPlainNumber(group.items.length)} 个节点 · ${formatPlainNumber(activeCount)} 可用 · 占 ${formatPlainNumber(shareRatio)}%</small>
        <span class="node-workbench-region-bar" aria-hidden="true">
          <span style="width: ${barWidth}%"></span>
        </span>
      </span>
      <span class="node-workbench-region-rate">${formatPlainNumber(activeRatio)}%</span>
    </article>
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
  const selectedNode = group.items.find((row) => row.id === appState.expandedNodeID);
  const detail = selectedNode ? appState.nodeDetail || selectedNode : null;
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
      <div class="node-region-workspace ${detail ? "has-node-detail" : ""}">
        <div class="node-card-grid">
          ${group.items.map((row) => renderNodeCard(row)).join("")}
        </div>
        ${detail ? renderNodeDetailPanel(detail) : ""}
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
            ? `<button class="table-button ghost-button" type="button" data-node-action="cancel" data-node-id="${row.id}" aria-label="取消节点编辑" title="取消节点编辑">${buttonLabel("×", "取消")}</button>`
            : `<button class="table-button" type="button" data-node-action="edit" data-node-id="${row.id}" aria-label="编辑节点" title="编辑节点">${buttonLabel("✎", "编辑")}</button>`
        }
        <button class="table-button ghost-button" type="button" data-node-action="reset-name" data-node-id="${row.id}" aria-label="恢复自动命名" title="恢复自动命名" ${row.name_mode === "auto" ? "disabled" : ""}>${buttonLabel("↺", "恢复自动")}</button>
      </div>
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
    <aside class="node-detail-drawer" role="dialog" aria-label="节点详情" aria-modal="false">
      <div class="node-detail-panel">
        <div class="node-detail-toolbar">
          <span class="node-detail-eyebrow">节点详情</span>
          <button class="table-button ghost-button node-detail-close" type="button" data-node-action="close-detail">${buttonLabel("×", "关闭")}</button>
        </div>
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
    </aside>
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
  const tags = Array.isArray(row.tags) ? row.tags.join(", ") : String(row.tags || "");
  const nameMode = String(row.name_mode || "auto").toLowerCase() === "manual" ? "manual" : "auto";
  return `
    <form class="node-edit-form" data-node-edit-form data-node-id="${row.id}">
      <label class="node-edit-field node-edit-field-wide">
        <span>展示名</span>
        <input name="display_name" value="${escapeHTML(row.display_name || "")}" required />
      </label>
      <label class="node-edit-field">
        <span>地区</span>
        <input name="region" value="${escapeHTML(row.region || "")}" placeholder="其他" />
      </label>
      <label class="node-edit-field">
        <span>命名模式</span>
        <select name="name_mode">
          <option value="auto" ${nameMode === "auto" ? "selected" : ""}>自动</option>
          <option value="manual" ${nameMode === "manual" ? "selected" : ""}>手动</option>
        </select>
      </label>
      <label class="node-edit-field node-edit-field-wide">
        <span>标签</span>
        <input name="tags" value="${escapeHTML(tags)}" placeholder="QA, VIP" />
      </label>
      <div class="node-edit-actions">
        <button class="table-button" type="submit">${buttonLabel("✓", "保存")}</button>
      </div>
    </form>
  `;
}

function renderTokens(rows) {
  appState.tokens = rows || [];
  if (!rows || rows.length === 0) {
    tokensEl.innerHTML = emptyState("钥", "暂无 Token", "签发 Token 后会生成通用、Mihomo 和 sing-box 订阅地址", {
      label: "签发 Token",
      symbol: "钥",
      view: "identity",
      target: "token-form",
      expand: true,
    });
    return;
  }
  tokensEl.innerHTML = `
    ${renderTokenWorkbench(rows)}
    <div class="token-card-list">
      ${rows.map((row) => renderTokenCard(row)).join("")}
    </div>
  `;
}

function renderTokenWorkbench(rows) {
  const tokenRows = rows || [];
  const activeRows = tokenRows.filter((row) => String(row.status || "active").toLowerCase() === "active");
  const reusableRows = tokenRows.filter((row) => row.subscription_available && tokenSubscriptionItems(row).length > 0);
  const formatCards = [
    {
      key: "default",
      symbol: "通",
      title: "通用 URI",
      detail: "Clash、Shadowrocket 等通用导入",
      urls: tokenRows.map((row) => tokenSubscriptionItems(row).find(([label]) => label === "默认订阅")?.[1]).filter(Boolean),
    },
    {
      key: "mihomo",
      symbol: "米",
      title: "Mihomo YAML",
      detail: "可直接导入 Mihomo / Clash Meta",
      urls: tokenRows.map((row) => tokenSubscriptionItems(row).find(([label]) => label === "Mihomo")?.[1]).filter(Boolean),
    },
    {
      key: "sing-box",
      symbol: "箱",
      title: "sing-box JSON",
      detail: "可直接导入 sing-box 客户端",
      urls: tokenRows.map((row) => tokenSubscriptionItems(row).find(([label]) => label === "sing-box")?.[1]).filter(Boolean),
    },
  ];
  const totalURLs = formatCards.reduce((total, card) => total + card.urls.length, 0);
  const allURLs = formatCards.flatMap((card) => card.urls);
  return `
    <section class="token-workbench" data-token-workbench aria-label="订阅分发工作台">
      <div class="token-workbench-heading">
        <span class="token-workbench-symbol" aria-hidden="true">订</span>
        <span class="token-workbench-title">
          <strong>订阅分发工作台</strong>
          <small>先确认格式、地址来源和可用 Token，再在下方卡片复制或探测。</small>
        </span>
      </div>
      <div class="token-workbench-stats" aria-label="订阅分发摘要">
        ${tokenWorkbenchStat("可用 Token", `${formatPlainNumber(reusableRows.length)}/${formatPlainNumber(tokenRows.length)}`, "可反复复制订阅")}
        ${tokenWorkbenchStat("启用状态", `${formatPlainNumber(activeRows.length)}/${formatPlainNumber(tokenRows.length)}`, "active 才可稳定分发")}
        ${tokenWorkbenchStat("订阅地址", `${formatPlainNumber(totalURLs)} 条`, tokenWorkbenchOriginSummary(allURLs))}
      </div>
      <div class="token-workbench-format-grid" aria-label="客户端格式">
        ${formatCards.map((card) => tokenWorkbenchFormatCard(card)).join("")}
      </div>
    </section>
  `;
}

function tokenWorkbenchStat(label, value, detail) {
  return `
    <span class="token-workbench-stat" data-token-workbench-stat>
      <span>${escapeHTML(label)}</span>
      <strong>${escapeHTML(String(value))}</strong>
      <small>${escapeHTML(detail || "")}</small>
    </span>
  `;
}

function tokenWorkbenchFormatCard(card) {
  const count = card.urls.length;
  const tone = count > 0 ? "success" : "warning";
  return `
    <article class="token-workbench-format-card token-workbench-format-card-${tone}" data-token-format-card="${escapeHTML(card.key)}">
      <span class="token-workbench-format-symbol" data-token-format-symbol aria-hidden="true">${escapeHTML(card.symbol)}</span>
      <span class="token-workbench-format-copy">
        <strong class="token-workbench-format-title">${escapeHTML(card.title)}</strong>
        <small class="token-workbench-format-meta">${escapeHTML(card.detail)}</small>
      </span>
      <span class="token-workbench-format-hint">${escapeHTML(count > 0 ? `${formatPlainNumber(count)} 条可用` : "待重置订阅")}</span>
    </article>
  `;
}

function tokenWorkbenchOriginSummary(urls) {
  if (!urls.length) return "暂无可用地址";
  const counts = urls.reduce((map, url) => {
    const origin = subscriptionOriginForURL(url);
    map.set(origin, (map.get(origin) || 0) + 1);
    return map;
  }, new Map());
  return [...counts.entries()]
    .sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0], "zh-CN"))
    .map(([origin, count]) => `${origin} ${formatPlainNumber(count)}`)
    .join(" · ");
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
  const items = tokenSubscriptionItems(row);
  if (!row.subscription_available || items.length === 0) {
    const message = row.subscription_error || "旧 Token 无法反复显示，可重置订阅";
    return `<div class="token-subscription-empty">${escapeHTML(message)}</div>`;
  }
  return renderSubscriptionCardList(items, row.id);
}

function tokenSubscriptionItems(row) {
  const subscriptions = row.subscriptions || {};
  return [
    ["默认订阅", subscriptions.default || row.subscription || ""],
    ["Mihomo", subscriptions.clash || ""],
    ["sing-box", subscriptions.sing_box || ""],
  ].filter((item) => item[1]);
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
                  <button class="table-button ghost-button" type="button" data-token-action="copy-subscription" data-token-id="${escapeHTML(String(tokenID || ""))}" data-token-url="${escapeHTML(url)}" data-button-symbol="⧉" data-copy-label="复制">${buttonLabel("⧉", "复制")}</button>
                  <button class="table-button ghost-button" type="button" data-token-action="probe-subscription" data-token-id="${escapeHTML(String(tokenID || ""))}" data-token-url="${escapeHTML(url)}">${buttonLabel("测", "探测")}</button>
                  <a class="table-button ghost-button link-button" href="${escapeHTML(url)}" target="_blank" rel="noopener noreferrer" data-token-subscription-link>${buttonLabel("↗", "打开")}</a>
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
  if (label === "Mihomo" || label === "Clash/Mihomo") return "Mihomo";
  if (label === "sing-box") return "sing-box";
  return "通用";
}

function subscriptionProfileForLabel(label) {
  if (label === "Mihomo" || label === "Clash/Mihomo") return "YAML · Mihomo";
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
    ["默认订阅", defaultSubscription],
    ["Mihomo", clashSubscription],
    ["sing-box", singBoxSubscription],
  ].filter((item) => item[1]);
  const formatCount = `${formatPlainNumber(items.length)} 格式`;
  tokenResultEl.hidden = false;
  tokenResultEl.innerHTML = `
    <div class="token-result-card">
      <div class="token-result-heading" data-token-result-heading>
        <span class="token-result-symbol" data-token-result-symbol aria-hidden="true">订</span>
        <span class="token-result-copy">
          <strong>${escapeHTML(title)}</strong>
          <span>复制后可直接导入对应客户端</span>
        </span>
        <span class="token-result-meta" data-token-result-meta>${escapeHTML(formatCount)}</span>
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
        ${emptyState("钥", "暂无 Token 用量", "有 Token 订阅访问或网关流量后会显示用量", {
          label: "签发 Token",
          symbol: "钥",
          view: "identity",
          target: "token-form",
          expand: true,
        })}
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
        ${emptyState("出", "暂无出口流量", "sing-box 网关产生真实流量后会按上游出口聚合", {
          label: "查看运维",
          symbol: "运",
          view: "ops",
          target: "ops-actions",
        })}
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
    target.innerHTML = emptyState("量", "暂无流量数据", "真实网关流量进入统计链路后会生成趋势图", {
      label: "查看运维",
      symbol: "运",
      view: "ops",
      target: "ops-actions",
    });
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
