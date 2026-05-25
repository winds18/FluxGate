# Admin UI Calm Ops Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构 FluxGate 管理后台为「冷静指挥台」体验，让后台现代、统一、无遮挡、可真实测试。

**Architecture:** 保持现有 Go 后端和静态前端，不引入前端框架；先用独立 `calm-ops.css` 建立覆盖层，再小步拆出 UI 状态与反馈逻辑。节点详情改为抽屉式交互，Token 卡片强化订阅复制与探测入口，QA 脚本覆盖桌面和移动视口的溢出与可用性。

**Tech Stack:** Go `net/http` 后端、SQLite store、原生 HTML/CSS/JavaScript、Playwright QA 脚本、现有 `scripts/dev/*` 与 `scripts/qa/*`。

---

## File Structure

- Modify: `internal/httpapi/static/index.html`
  - 加载新的 `calm-ops.css`，保留现有 DOM 结构以降低风险。
- Create: `internal/httpapi/static/calm-ops.css`
  - 承载新版视觉语言、统一控件、响应式布局、抽屉、节点/Token 卡片、长文本防溢出。
- Modify: `internal/httpapi/static/app.js`
  - 增加人性化反馈工具函数、危险操作确认、节点详情抽屉渲染、Token 订阅操作文案优化。
- Modify: `scripts/qa/browser-login.sh`
  - 扩展现有 Playwright 检查：多视口、横向溢出、节点卡片标题、Token 订阅地址、移动 Dock 遮挡。
- Modify: `docs/11-current-implementation.md`
  - 同步 UI 重构进度和验收口径。

## Task 1: Add Calm Ops Visual Layer

**Files:**
- Modify: `internal/httpapi/static/index.html`
- Create: `internal/httpapi/static/calm-ops.css`

- [ ] **Step 1: Load the new stylesheet after the legacy stylesheet**

```html
<link rel="stylesheet" href="/assets/styles.css" />
<link rel="stylesheet" href="/assets/calm-ops.css" />
```

- [ ] **Step 2: Add the Calm Ops design token layer**

```css
:root {
  --fg-bg: #eef3f7;
  --fg-surface: rgba(255, 255, 255, 0.88);
  --fg-surface-solid: #ffffff;
  --fg-border: #d8e2ec;
  --fg-text: #111827;
  --fg-muted: #667085;
  --fg-primary: #1769e0;
  --fg-success: #0f7a55;
  --fg-warning: #9a5b00;
  --fg-danger: #b42318;
  --fg-shadow-soft: 0 10px 28px rgba(15, 23, 42, 0.08);
}
```

- [ ] **Step 3: Run static build check**

Run: `scripts/dev/build.sh`
Expected: exits 0 and reports successful build.

- [ ] **Step 4: Commit**

```bash
git add internal/httpapi/static/index.html internal/httpapi/static/calm-ops.css
git commit -m "Add calm ops admin visual layer"
```

## Task 2: Humanize Core Interactions

**Files:**
- Modify: `internal/httpapi/static/app.js`

- [ ] **Step 1: Add small UI feedback helpers near `showCopyFeedback`**

```js
function setStatus(message, tone = "neutral") {
  statusEl.textContent = message;
  statusEl.dataset.tone = tone;
}

function confirmDanger(message) {
  return window.confirm(message);
}
```

- [ ] **Step 2: Replace direct status mutations for login, load, copy, token actions**

Use `setStatus("刷新中", "neutral")`, `setStatus("已连接", "success")`, `setStatus("更新失败", "danger")`.

- [ ] **Step 3: Protect dangerous Token actions**

```js
if (action === "revoke" && !confirmDanger("确认撤销这个 Token？撤销后伙伴将无法继续使用该订阅。")) {
  button.disabled = false;
  setStatus("已取消撤销", "neutral");
  return;
}
```

- [ ] **Step 4: Run JS syntax check through build**

Run: `scripts/dev/build.sh`
Expected: exits 0.

- [ ] **Step 5: Commit**

```bash
git add internal/httpapi/static/app.js
git commit -m "Humanize admin interaction feedback"
```

## Task 3: Convert Node Details Into Drawer Flow

**Files:**
- Modify: `internal/httpapi/static/app.js`
- Modify: `internal/httpapi/static/calm-ops.css`

- [ ] **Step 1: Render node detail outside the card grid**

In `renderNodeRegion`, compute the selected node and append:

```js
const selectedNode = group.items.find((node) => node.id === appState.expandedNodeID);
${selectedNode ? renderNodeDetailPanel(appState.nodeDetail || selectedNode) : ""}
```

- [ ] **Step 2: Change `renderNodeDetailPanel` wrapper to a drawer**

```html
<aside class="node-detail-drawer" role="dialog" aria-label="节点详情">
  <div class="node-detail-drawer-header">
    <strong>节点详情</strong>
    <button class="table-button ghost-button" type="button" data-node-action="close-detail" data-node-id="...">关闭</button>
  </div>
  ...
</aside>
```

- [ ] **Step 3: Handle `close-detail` before loading details**

```js
if (action === "close-detail") {
  appState.expandedNodeID = null;
  appState.nodeDetail = null;
  renderNodes(appState.nodes);
  return;
}
```

- [ ] **Step 4: Style drawer for desktop and mobile**

Desktop: fixed right panel, width 480px, max-width `calc(100vw - 32px)`.
Mobile: full-screen panel above bottom Dock, with enough bottom padding.

- [ ] **Step 5: Run build and browser login QA**

Run: `scripts/dev/build.sh`
Run: `scripts/qa/browser-login.sh http://127.0.0.1:8081` after local server is available.
Expected: no JS error and node module remains navigable.

- [ ] **Step 6: Commit**

```bash
git add internal/httpapi/static/app.js internal/httpapi/static/calm-ops.css
git commit -m "Move node details into a drawer"
```

## Task 4: Rework Token Subscription Cards

**Files:**
- Modify: `internal/httpapi/static/app.js`
- Modify: `internal/httpapi/static/calm-ops.css`

- [ ] **Step 1: Rename subscription labels in `renderTokenSubscriptions`**

Use `默认订阅`, `Mihomo`, `sing-box` as visible labels while keeping existing URLs.

- [ ] **Step 2: Render copy-first actions**

Each subscription row must show Copy first, Open second:

```html
<button class="table-button" type="button" data-token-action="copy-subscription" ...>复制</button>
<a class="table-button ghost-button link-button" href="..." target="_blank" rel="noopener noreferrer">打开</a>
```

- [ ] **Step 3: Add stable long-URL styling**

Subscription URL code blocks use `word-break: break-all`, max-height, and no layout growth beyond card width.

- [ ] **Step 4: Run build**

Run: `scripts/dev/build.sh`
Expected: exits 0.

- [ ] **Step 5: Commit**

```bash
git add internal/httpapi/static/app.js internal/httpapi/static/calm-ops.css
git commit -m "Refine token subscription cards"
```

## Task 5: Expand Browser Visual QA

**Files:**
- Modify: `scripts/qa/browser-login.sh`

- [ ] **Step 1: Add desktop and mobile viewport loops**

Add checks for `1440x900`, `1280x720`, `390x844`.

- [ ] **Step 2: Add generic overflow detector**

```js
const overflow = await page.evaluate(() => Math.max(0, document.documentElement.scrollWidth - window.innerWidth));
expect(overflow).toBeLessThanOrEqual(1);
```

- [ ] **Step 3: Add Token and node-specific checks**

Check that `.token-subscription-item code`, `.node-card-title`, `.node-detail-drawer` have widths not exceeding their parent containers.

- [ ] **Step 4: Run browser QA against local server**

Run: `scripts/qa/local-suite.sh`
Expected: local suite passes and cleans temporary artifacts.

- [ ] **Step 5: Commit**

```bash
git add scripts/qa/browser-login.sh
git commit -m "Expand admin UI visual QA"
```

## Task 6: Final Verification, Docs, and Deployment

**Files:**
- Modify: `docs/11-current-implementation.md`

- [ ] **Step 1: Update current implementation notes**

Record that the admin UI now uses Calm Ops visual language, node detail drawer, copy-first Token subscriptions, and multi-viewport QA.

- [ ] **Step 2: Run full validation**

Run: `scripts/qa/local-suite.sh`
Run: `scripts/qa/cleanup.sh`
Run: `scripts/qa/public-scan.sh`
Expected: all exit 0, with no retained temporary files unless `KEEP_ARTIFACTS=true`.

- [ ] **Step 3: Commit docs**

```bash
git add docs/11-current-implementation.md
git commit -m "Document calm ops admin UI implementation"
```

- [ ] **Step 4: Push and deploy**

Run: `scripts/deploy/push-and-deploy.sh`
Expected: push succeeds, remote build succeeds, remote health/ready/login/overview verification passes.

## Task 7: Humanize Node Safe Editing

**Files:**
- Modify: `internal/store/nodes.go`
- Modify: `internal/httpapi/server.go`
- Modify: `internal/httpapi/static/app.js`
- Modify: `internal/httpapi/static/calm-ops.css`
- Modify: `internal/store/store_test.go`
- Modify: `scripts/qa/api-flow.sh`
- Modify: `scripts/qa/browser-login.sh`
- Modify: `docs/11-current-implementation.md`

- [x] **Step 1: Add Store behavior coverage**

Add a failing test for editable node fields: display name, region override, tags, name mode, and auto-name reset preserving operator-managed fields.

- [x] **Step 2: Add `UpdateNode` safe-field API in the store**

Keep `UpdateNodeDisplayName` and `ResetNodeDisplayName` compatible while routing them through the new safe-field update path.

- [x] **Step 3: Expose safe fields through `PATCH /api/nodes/{id}`**

Decode `store.UpdateNodeInput` and keep raw URI/protocol/server fields read-only through the detail drawer.

- [x] **Step 4: Replace single-field node edit with a labeled edit panel**

Show display name, region, name mode and tags with unified control sizing and responsive no-overflow layout.

- [x] **Step 5: Expand API and browser QA**

API flow verifies node edit/reset behavior and browser QA verifies field presence plus edit panel overflow.

## Task 8: Replace Native Token Confirms With In-App Dialogs

**Files:**
- Modify: `internal/httpapi/static/app.js`
- Modify: `internal/httpapi/static/calm-ops.css`
- Modify: `scripts/qa/browser-login.sh`
- Modify: `docs/11-current-implementation.md`

- [x] **Step 1: Add browser QA before implementation**

Browser QA clicks the Token reset-subscription action, expects a visible in-app confirmation dialog, verifies the `重置订阅地址` title, checks that cancel/confirm actions are both present, then cancels and confirms the dialog closes with an `已取消` status.

- [x] **Step 2: Implement the custom dangerous-action dialog**

`confirmDanger` now renders an accessible in-app dialog with title, impact copy, cancel/confirm buttons, backdrop cancel, `Escape` cancel and focus restoration.

- [x] **Step 3: Apply the dialog to dangerous Token actions**

Token revoke and subscription reset both use the same dialog pattern and return clear cancellation feedback before any API mutation is sent.

- [x] **Step 4: Add responsive dialog styling**

The dialog uses the Calm Ops radius, shadow, danger symbol, compact action layout, and mobile-safe single-column actions without relying on native browser confirm UI.

- [x] **Step 5: Verify through build and full local QA**

Run `scripts/dev/build.sh` and `scripts/qa/local-suite.sh`; expected result is zero exit status, with browser QA confirming dialog visibility, title, action count and cancellation.

## Task 9: Humanize Overview Initialization Guide

**Files:**
- Modify: `internal/httpapi/server.go`
- Modify: `internal/httpapi/static/app.js`
- Modify: `internal/httpapi/static/calm-ops.css`
- Modify: `scripts/qa/browser-login.sh`
- Modify: `docs/11-current-implementation.md`
- Modify: `docs/07-roadmap.md`

- [x] **Step 1: Add browser QA before implementation**

Browser QA now requires the overview readiness panel to expose a single progress value, a single current action block, five step hints, and zero overflow for both the guide shell and readiness items.

- [x] **Step 2: Render a human-friendly guide state**

The overview readiness panel now presents `初始化向导`, `可测进度`, the next recommended action, contextual copy, and a direct module jump button instead of only listing raw readiness states.

- [x] **Step 3: Add per-step guidance**

Each readiness step includes a concise hint for what to do next or how to review the ready state, so the user can complete the real-test loop without guessing which module to open.

- [x] **Step 4: Fix and assert Calm Ops stylesheet delivery**

The server now serves `/assets/calm-ops.css`, and browser QA asserts the stylesheet response is `200` so modern layout fixes cannot silently fail.

- [x] **Step 5: Verify through build, full QA, cleanup, and public scan**

Run `scripts/qa/local-suite.sh`, `scripts/dev/build.sh`, `scripts/dev/test.sh`, `scripts/qa/cleanup.sh`, `scripts/qa/public-scan.sh`, and `git diff --check`; expected result is zero exit status with guide overflow counts at 0.
