#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

INDEX_FILE="$ROOT_DIR/internal/httpapi/static/index.html"
APP_FILE="$ROOT_DIR/internal/httpapi/static/app.js"
CALM_CSS_FILE="$ROOT_DIR/internal/httpapi/static/calm-ops.css"

require_pattern() {
  local file="$1"
  local pattern="$2"
  local message="$3"
  if ! rg -q -- "$pattern" "$file"; then
    log "$message"
    exit 1
  fi
}

require_multiline_pattern() {
  local file="$1"
  local pattern="$2"
  local message="$3"
  if ! rg -Uq -- "$pattern" "$file"; then
    log "$message"
    exit 1
  fi
}

require_count() {
  local file="$1"
  local pattern="$2"
  local expected="$3"
  local message="$4"
  local count
  count="$(rg -c -- "$pattern" "$file" || true)"
  if [[ "$count" != "$expected" ]]; then
    log "$message: expected=$expected actual=$count"
    exit 1
  fi
}

log "running static UI contract checks"

require_pattern "$INDEX_FILE" 'href="/assets/calm-ops.css"' "calm ops stylesheet must be loaded"
require_count "$INDEX_FILE" 'data-view-jump-target=' 4 "overview quick cards must deep-link to concrete module targets"
require_pattern "$APP_FILE" 'navigateToDashboardTarget\(' "overview quick cards must navigate and focus target panels"
require_pattern "$APP_FILE" 'button\.dataset\.viewJumpTarget' "overview quick cards must pass concrete target panels"
require_pattern "$APP_FILE" 'getJSON\("/api/delivery/readiness"\)' "overview guide must load delivery readiness for real-test guidance"
require_pattern "$APP_FILE" '订阅分发' "overview guide must include subscription distribution as a first-class test step"
require_pattern "$APP_FILE" '发布检查' "overview guide must include gateway publishing/readiness as a first-class test step"
require_pattern "$APP_FILE" 'guide-action-strip' "overview guide must expose direct real-test actions"
require_pattern "$APP_FILE" '去复制订阅' "overview guide must provide a direct subscription copy entry"
require_pattern "$APP_FILE" 'target: "config-publish"' "overview guide must deep-link to gateway publish action"
require_pattern "$INDEX_FILE" 'id="overview-hero"' "overview must expose a first-screen command summary strip"
require_pattern "$APP_FILE" 'renderOverviewHero\(' "overview command summary must be rendered from live state"
require_pattern "$APP_FILE" 'data-overview-hero-action' "overview command summary must provide a direct humane next action"
require_pattern "$INDEX_FILE" 'id="overview-insights"' "overview must expose a live control-room insight grid"
require_pattern "$APP_FILE" 'renderOverviewInsights\(' "overview insights must be rendered from live state"
require_pattern "$APP_FILE" 'overviewMetricCard\(' "overview metrics must use readiness-oriented KPI cards"
require_pattern "$APP_FILE" 'renderSourceWorkbench\(' "source list must render an access health workbench"
require_pattern "$APP_FILE" 'sourceWorkbenchTypeCard\(' "source workbench must expose source type distribution cards"
require_pattern "$APP_FILE" 'renderTokenWorkbench\(' "token list must render a subscription delivery workbench"
require_pattern "$APP_FILE" 'tokenWorkbenchFormatCard\(' "token delivery workbench must expose per-client format cards"
require_pattern "$APP_FILE" 'renderNodeWorkbench\(' "node pool must render a first-screen workbench"
require_pattern "$APP_FILE" 'data-node-workbench' "node pool workbench must expose a stable QA hook"
require_pattern "$INDEX_FILE" 'id="workspace-insight"' "workspace header must expose a humane module insight strip"
require_pattern "$APP_FILE" 'updateWorkspaceInsight\(' "workspace insight strip must be refreshed from live state"
require_pattern "$APP_FILE" 'workspaceInsightForView\(' "workspace insight strip must derive per-module guidance"
require_pattern "$APP_FILE" 'data-workspace-insight-action' "workspace insight strip must expose a direct contextual action"
require_pattern "$CALM_CSS_FILE" 'FluxGate UI v2 shell' "UI v2 shell layer must be documented in CSS"
require_pattern "$CALM_CSS_FILE" '--fg-sidebar-bg: #ffffff;' "UI v2 shell must use a light sidebar surface"
require_pattern "$CALM_CSS_FILE" '--fg-panel-shadow:' "UI v2 shell must define a restrained panel shadow"
require_pattern "$CALM_CSS_FILE" '\.app-shell \{' "UI v2 shell must govern the application frame"
require_pattern "$CALM_CSS_FILE" '\.dashboard-sidebar \{' "UI v2 shell must govern the sidebar"
require_pattern "$CALM_CSS_FILE" '\.overview-hero \{' "UI v2 shell must govern the real-test overview strip"
require_pattern "$CALM_CSS_FILE" '\.overview-insights \{' "overview insights must have a governed grid"
require_pattern "$CALM_CSS_FILE" '\.overview-insight-panel \{' "overview insight panels must be visually governed"
require_pattern "$CALM_CSS_FILE" '\.overview-distribution-row \{' "overview region distribution rows must be stable"
require_pattern "$CALM_CSS_FILE" '\.overview-sync-row \{' "overview recent sync rows must be stable"
require_pattern "$CALM_CSS_FILE" '\.source-workbench \{' "source health workbench must be visually governed"
require_pattern "$CALM_CSS_FILE" '\.source-workbench-type-card \{' "source workbench type cards must be visually governed"
require_pattern "$CALM_CSS_FILE" '\.token-workbench \{' "token subscription workbench must be visually governed"
require_pattern "$CALM_CSS_FILE" '\.token-workbench-format-card \{' "token subscription format cards must be visually governed"
require_pattern "$CALM_CSS_FILE" '\.node-workbench \{' "node pool workbench must be visually governed"
require_pattern "$CALM_CSS_FILE" '\.node-workbench-region-card \{' "node pool workbench region cards must be visually governed"
require_pattern "$CALM_CSS_FILE" 'overflow-wrap: anywhere;' "UI v2 shell must protect long text from escaping cards"
require_pattern "$CALM_CSS_FILE" 'text-overflow: ellipsis;' "UI v2 shell must provide single-line truncation where needed"
require_pattern "$CALM_CSS_FILE" '\.workspace-header \{' "workspace command header styles are missing"
require_pattern "$CALM_CSS_FILE" 'position: sticky;' "workspace header must keep module actions reachable on desktop"
require_pattern "$CALM_CSS_FILE" 'scroll-snap-type: x proximity;' "module command strips must remain horizontally scannable"
require_pattern "$CALM_CSS_FILE" 'scrollbar-width: none;' "module command strips should hide visual scrollbar noise"
require_pattern "$CALM_CSS_FILE" 'flex: 0 0 auto;' "command chips and rail items must not shrink into unreadable pills"
require_pattern "$CALM_CSS_FILE" '\.overview-hero \{' "overview command summary must have a governed visual shell"
require_pattern "$CALM_CSS_FILE" '\.overview-hero-stat \{' "overview command summary must expose stable status cells"
require_pattern "$CALM_CSS_FILE" '\.workspace-insight \{' "workspace insight strip must have a governed visual shell"
require_pattern "$CALM_CSS_FILE" '\.workspace-insight-action \{' "workspace insight action must stay visually stable"
require_pattern "$CALM_CSS_FILE" '\.workspace-actions \{' "workspace actions must be governed by the calm ops layer"
require_pattern "$CALM_CSS_FILE" 'grid-template-columns: repeat\(2, minmax\(0, 1fr\)\);' "mobile workspace actions must remain balanced and tappable"
require_pattern "$CALM_CSS_FILE" '\.source-actions,' "card action toolbars must be governed by the calm ops layer"
require_pattern "$CALM_CSS_FILE" 'justify-content: flex-end;' "desktop card action toolbars must not look like full-width primary blocks"
require_pattern "$CALM_CSS_FILE" 'flex: 0 1 auto;' "desktop card action buttons must stay compact"
require_pattern "$CALM_CSS_FILE" 'button\[data-source-action="edit"\]' "low-risk inline edit actions must be quiet secondary controls"
require_pattern "$CALM_CSS_FILE" 'button\[data-source-action="save"\]' "inline save actions must keep a clear primary affordance"
require_pattern "$CALM_CSS_FILE" '\.source-actions \.table-button,' "mobile card action bars must retain tappable full-width behavior"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-card \{[^}]*position: relative;' "node cards must anchor actions without adding empty vertical space"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \{[^}]*position: absolute;' "node card actions must be docked instead of consuming a full card row"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \{[^}]*top: 12px;' "node card actions must stay close to the card title"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \.table-button \{[^}]*flex: 0 0 auto;' "node card action buttons must not stretch into full-width blocks on desktop"
require_pattern "$CALM_CSS_FILE" 'padding-right: 64px;' "node card content must reserve only the compact icon action lane"
require_pattern "$APP_FILE" 'aria-label="编辑节点"' "node edit icon button must have an accessible label"
require_pattern "$APP_FILE" 'title="恢复自动命名"' "node reset icon button must explain the action on hover"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \.button-label \{[^}]*clip-path: inset\(50%\);' "desktop node icon actions must hide truncated text labels visually"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \.table-button \{[^}]*aspect-ratio: 1;' "desktop node icon actions must be stable square controls"
require_pattern "$APP_FILE" 'syncFormDrawerShellState\(' "form drawers must synchronize a real side-sheet shell state"
require_pattern "$APP_FILE" 'document\.body\.classList\.toggle\("has-form-drawer-open"' "opening a form drawer must mark the page shell"
require_pattern "$APP_FILE" 'aria-modal' "expanded form drawers must expose dialog semantics"
require_multiline_pattern "$CALM_CSS_FILE" '\.form-drawer:not\(\.is-collapsed\) \{[^}]*position: fixed;' "expanded form drawers must float as side sheets instead of pushing the page"
require_multiline_pattern "$CALM_CSS_FILE" '\.form-drawer:not\(\.is-collapsed\) \{[^}]*width: min\(520px, calc\(100vw - 36px\)\);' "desktop form side sheets must use a stable humane width"
require_multiline_pattern "$CALM_CSS_FILE" 'body\.has-form-drawer-open::before \{[^}]*backdrop-filter: blur\(2px\);' "open form side sheets must calm the background instead of competing with content"

log "static UI contract checks passed"
