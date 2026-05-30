# Admin UI v2 Overview Control Room Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the overview first screen into a compact operations control room that immediately shows whether FluxGate is truly testable and where to act next.

**Architecture:** Keep the static HTML/CSS/JS frontend and existing APIs. Replace generic overview counters with six readiness-oriented KPI cards, then add a live insight grid derived from current sources, nodes, tokens, virtual gateways, policies, and delivery readiness.

**Tech Stack:** Static HTML, vanilla JavaScript, CSS, Bash QA scripts, Playwright-based browser QA.

---

## File Map

- `internal/httpapi/static/index.html`: add the overview insight grid host.
- `internal/httpapi/static/app.js`: render six KPI cards and three insight panels from live state.
- `internal/httpapi/static/calm-ops.css`: style the overview control room, insight panels, distribution rows, and compact metric cards.
- `scripts/qa/ui-static.sh`: require the overview control room host, renderer, and CSS contracts.
- `scripts/qa/browser-login.sh`: verify the six KPI cards and three insight panels do not overflow.

## Task 1: Add Failing Overview Contracts

**Files:**
- Modify: `scripts/qa/ui-static.sh`
- Modify: `scripts/qa/browser-login.sh`

- [ ] **Step 1: Add static contracts**

Require:

```bash
require_pattern "$INDEX_FILE" 'id="overview-insights"' "overview must expose a live control-room insight grid"
require_pattern "$APP_FILE" 'renderOverviewInsights\(' "overview insights must be rendered from live state"
require_pattern "$APP_FILE" 'overviewMetricCard\(' "overview metrics must use readiness-oriented KPI cards"
require_pattern "$CALM_CSS_FILE" '\.overview-insights \{' "overview insights must have a governed grid"
require_pattern "$CALM_CSS_FILE" '\.overview-insight-panel \{' "overview insight panels must be visually governed"
require_pattern "$CALM_CSS_FILE" '\.overview-distribution-row \{' "overview region distribution rows must be stable"
require_pattern "$CALM_CSS_FILE" '\.overview-sync-row \{' "overview recent sync rows must be stable"
```

- [ ] **Step 2: Add browser QA checks**

Update `scripts/qa/browser-login.sh` so overview QA expects:

- `#metrics .metric` count is 6.
- `#overview-insights [data-overview-insight]` count is 3.
- each insight panel has one symbol and one title.
- region rows, sync rows, and health pills exist.
- no child element escapes its insight panel.

- [ ] **Step 3: Run static QA and watch it fail**

Run: `scripts/qa/ui-static.sh`

Expected: FAIL because `#overview-insights` and `renderOverviewInsights(` do not exist yet.

## Task 2: Implement Overview Control Room

**Files:**
- Modify: `internal/httpapi/static/index.html`
- Modify: `internal/httpapi/static/app.js`
- Modify: `internal/httpapi/static/calm-ops.css`

- [ ] **Step 1: Add insight host**

Add this after `overview-board` and before `quick-grid`:

```html
<section id="overview-insights" class="overview-insights" aria-label="运行洞察"></section>
```

- [ ] **Step 2: Render six KPI cards**

Change `renderMetrics(data)` to render six readiness-oriented cards:

- 真实闭环: `readyCount/checkCount`
- 可用节点: `activeNodes/totalNodes`
- 上游来源: `healthySources/totalSources`
- 可用订阅: `activeTokens/totalTokens`
- 网关入口: `activeVirtualNodes/totalVirtualNodes`
- 访问策略: `activePolicies/totalPolicies`

Keep the existing `.metric`, `.metric-symbol`, and `.metric-value` classes so existing layout checks still apply.

- [ ] **Step 3: Render three insight panels**

Add `renderOverviewInsights(data)` and call it during dashboard render. Panels:

- `data-overview-insight="source-health"`: source health pills and the three most recently synced sources.
- `data-overview-insight="region-distribution"`: top five node regions with count bars and active/total summary.
- `data-overview-insight="delivery"`: token, virtual gateway, policy, and publish readiness summary.

- [ ] **Step 4: Style insight panels**

Add v2 CSS for:

- `.overview-insights`
- `.overview-insight-panel`
- `.overview-insight-heading`
- `.overview-health-pills`
- `.overview-health-pill`
- `.overview-distribution-list`
- `.overview-distribution-row`
- `.overview-sync-list`
- `.overview-sync-row`
- `.overview-delivery-list`

All rows must use `min-width: 0`, `overflow: hidden`, and internal ellipsis/anywhere wrapping for long names.

## Task 3: Verify And Ship

**Files:**
- Test only after implementation unless failures expose issues.

- [ ] **Step 1: Run static QA**

Run: `scripts/qa/ui-static.sh`

Expected: PASS.

- [ ] **Step 2: Run full local suite**

Run: `scripts/qa/local-suite.sh`

Expected: PASS, including browser QA and screenshot capture.

- [ ] **Step 3: Cleanup and public scan**

Run:

```bash
scripts/qa/cleanup.sh
scripts/qa/public-scan.sh
```

Expected: both pass.

- [ ] **Step 4: Commit, push, and deploy**

Run:

```bash
git add docs/superpowers/plans/2026-05-30-admin-ui-v2-overview-control-room.md internal/httpapi/static/index.html internal/httpapi/static/app.js internal/httpapi/static/calm-ops.css scripts/qa/ui-static.sh scripts/qa/browser-login.sh
git commit -m "Refine admin overview control room"
git push
scripts/deploy/push-and-deploy.sh
scripts/qa/usable-probe.sh http://192.168.66.10:18080
```
