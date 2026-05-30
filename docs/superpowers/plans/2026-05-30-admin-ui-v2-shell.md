# Admin UI v2 Shell Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the FluxGate admin visual shell so the existing control panel looks like a modern, stable operations console without changing backend behavior.

**Architecture:** Keep the current static HTML/CSS/JS frontend and add the v2 visual system as a focused CSS/QA increment. Preserve existing module rendering and APIs, but replace the heavy Calm Ops visual treatment with a lighter dashboard shell, consistent controls, safer text containment, and screenshot-verifiable layout rules.

**Tech Stack:** Static HTML, vanilla JavaScript, CSS, Bash QA scripts, Playwright-based browser QA.

---

## File Map

- `docs/superpowers/specs/2026-05-30-admin-ui-v2-visual-system.md`: locked visual target and acceptance criteria.
- `docs/superpowers/plans/2026-05-30-admin-ui-v2-shell.md`: this implementation plan.
- `internal/httpapi/static/calm-ops.css`: primary implementation surface for visual shell, controls, cards, text containment, responsive rules.
- `scripts/qa/ui-static.sh`: static contract checks for v2 shell tokens and overflow protections.
- `scripts/qa/browser-login.sh`: existing browser QA; this increment must keep it passing.

## Task 1: Static Contract For UI v2 Shell

**Files:**
- Modify: `scripts/qa/ui-static.sh`

- [ ] **Step 1: Add v2 shell contract checks**

Add checks that prove the CSS contains a named v2 shell layer, light sidebar treatment, compact work surface, light overview status strip, and explicit text containment utilities.

Expected checks:

```bash
require_pattern "$CALM_CSS_FILE" 'FluxGate UI v2 shell' "UI v2 shell layer must be documented in CSS"
require_pattern "$CALM_CSS_FILE" '--fg-sidebar-bg: #ffffff;' "UI v2 shell must use a light sidebar surface"
require_pattern "$CALM_CSS_FILE" '--fg-panel-shadow:' "UI v2 shell must define a restrained panel shadow"
require_pattern "$CALM_CSS_FILE" '\.app-shell \{' "UI v2 shell must govern the application frame"
require_pattern "$CALM_CSS_FILE" '\.dashboard-sidebar \{' "UI v2 shell must govern the sidebar"
require_pattern "$CALM_CSS_FILE" '\.overview-hero \{' "UI v2 shell must govern the real-test overview strip"
require_pattern "$CALM_CSS_FILE" 'overflow-wrap: anywhere;' "UI v2 shell must protect long text from escaping cards"
require_pattern "$CALM_CSS_FILE" 'text-overflow: ellipsis;' "UI v2 shell must provide single-line truncation where needed"
```

- [ ] **Step 2: Run static QA and verify it fails before CSS work**

Run: `scripts/qa/ui-static.sh`

Expected: FAIL with `UI v2 shell layer must be documented in CSS`.

## Task 2: Rebuild Visual Shell In CSS

**Files:**
- Modify: `internal/httpapi/static/calm-ops.css`

- [ ] **Step 1: Add v2 design tokens**

Add a top comment and token set in `:root`:

```css
/* FluxGate UI v2 shell */
:root {
  --fg-page-bg: #f5f7fb;
  --fg-sidebar-bg: #ffffff;
  --fg-sidebar-border: #e5eaf0;
  --fg-panel-shadow: 0 1px 2px rgba(15, 23, 42, 0.05), 0 14px 38px rgba(15, 23, 42, 0.06);
  --fg-soft-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
}
```

- [ ] **Step 2: Replace heavy shell treatment**

Update `body`, `.topbar`, `.app-shell`, `.dashboard-sidebar`, `.sidebar-brand`, `.side-nav`, `.nav-item`, `.nav-symbol`, `.workspace`, `.workspace-header`, `.workspace-actions`, `.status`, `.mobile-dock`, and `.mobile-more-panel` to use the v2 light operations console treatment.

Requirements:

- Sidebar is light, compact, and bordered.
- Active nav is blue-tinted with a left blue accent.
- Topbar is quieter and does not compete with the workspace.
- Workspace header is a compact sticky panel, not a large card stack.
- Buttons and status chips keep stable heights.

- [ ] **Step 3: Replace heavy overview strip**

Update `.overview-hero` and children so the real-test overview area is a light readiness strip instead of a dark hero block.

Requirements:

- No dark marketing-style hero.
- Status dot, title, copy, stat cells, and action button fit in one compact panel on desktop.
- Mobile stacks cleanly.

- [ ] **Step 4: Strengthen text containment**

Add or tighten selectors for cards, headings, chips, code blocks, tables, and subscription cards:

```css
.panel,
.overview-panel,
.token-card,
.source-card,
.identity-card,
.virtual-node-card,
.policy-card,
.traffic-card-section,
.node-card,
.node-region-card {
  min-width: 0;
  overflow: hidden;
}

.panel *,
.token-card *,
.source-card *,
.node-card *,
.node-detail-drawer *,
.token-subscription-item * {
  min-width: 0;
}
```

Keep existing targeted `overflow-wrap: anywhere;` and `text-overflow: ellipsis;` protections.

- [ ] **Step 5: Run static QA and verify it passes**

Run: `scripts/qa/ui-static.sh`

Expected: PASS with `static UI contract checks passed`.

## Task 3: Browser QA And Visual Smoke

**Files:**
- Test only: no implementation files unless QA exposes issues.

- [ ] **Step 1: Run browser login QA**

Run: `scripts/qa/browser-login.sh`

Expected: PASS.

- [ ] **Step 2: Run local usable probe**

Run: `scripts/qa/usable-probe.sh`

Expected: PASS.

- [ ] **Step 3: Run cleanup and public scan**

Run:

```bash
scripts/qa/cleanup.sh
scripts/qa/public-scan.sh
```

Expected: both pass.

## Task 4: Commit And Deploy

**Files:**
- Stage and commit the design spec, plan, QA script, and CSS changes.

- [ ] **Step 1: Review changed files**

Run: `git status --short`

Expected changed files:

- `docs/superpowers/specs/2026-05-30-admin-ui-v2-visual-system.md`
- `docs/superpowers/plans/2026-05-30-admin-ui-v2-shell.md`
- `internal/httpapi/static/calm-ops.css`
- `scripts/qa/ui-static.sh`

- [ ] **Step 2: Commit**

Run:

```bash
git add docs/superpowers/specs/2026-05-30-admin-ui-v2-visual-system.md docs/superpowers/plans/2026-05-30-admin-ui-v2-shell.md internal/httpapi/static/calm-ops.css scripts/qa/ui-static.sh
git commit -m "Refine admin UI v2 shell"
```

- [ ] **Step 3: Push and deploy**

Run:

```bash
git push
scripts/deploy/push-and-deploy.sh
```

Expected: remote deploy completes and reports a deploy id.
