#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

INDEX_FILE="$ROOT_DIR/internal/httpapi/static/index.html"
APP_FILE="$ROOT_DIR/internal/httpapi/static/app.js"
CSS_FILE="$ROOT_DIR/internal/httpapi/static/styles.css"
SERVER_FILE="$ROOT_DIR/internal/httpapi/server.go"
SCREENSHOT_SCRIPT="$ROOT_DIR/scripts/qa/screenshot.sh"

require_pattern() {
  local file="$1"
  local pattern="$2"
  local message="$3"
  if ! rg -q -- "$pattern" "$file"; then
    log "$message"
    exit 1
  fi
}

require_absent() {
  local file="$1"
  local pattern="$2"
  local message="$3"
  if rg -q -- "$pattern" "$file"; then
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

log "running static UI contract checks (v3 quiet console)"

# ---- 样式层：单一文件，设计令牌 ----
require_pattern "$INDEX_FILE" 'href="/assets/styles.css"' "main stylesheet must be loaded"
require_absent "$INDEX_FILE" 'calm-ops' "legacy calm-ops stylesheet must not be referenced"
require_absent "$SERVER_FILE" 'calm-ops' "legacy calm-ops route must not be served"
require_pattern "$CSS_FILE" 'FluxGate UI v3' "v3 design layer must be documented in CSS"
require_pattern "$CSS_FILE" '--accent:' "design tokens must define a single accent color"
require_pattern "$CSS_FILE" '--border:' "design tokens must define the hairline border color"
require_pattern "$CSS_FILE" '--fg-mobile-content-safe:' "mobile shell must define a shared bottom safe area for the dock"
require_pattern "$CSS_FILE" '--fg-mobile-overlay-bottom:' "mobile overlays must align with the dock safe area"
require_pattern "$CSS_FILE" 'prefers-reduced-motion' "motion must respect the user reduced-motion preference"
require_pattern "$CSS_FILE" 'overflow-wrap: anywhere;' "long text must not escape cards"

# ---- 导航与骨架 ----
require_count "$INDEX_FILE" 'data-view-nav=' 14 "desktop nav (7), mobile dock (4) and more-menu (3) must cover all modules"
require_count "$INDEX_FILE" 'side-nav-section-title' 2 "desktop sidebar must group modules into scannable sections"
require_pattern "$INDEX_FILE" 'data-mobile-more-toggle' "mobile dock must expose the overflow menu toggle"
require_pattern "$INDEX_FILE" 'id="mobile-more-menu"' "mobile overflow menu must exist"
require_pattern "$APP_FILE" 'document\.body\.classList\.toggle\("is-authenticated"' "authenticated shell must mark body state"
require_pattern "$CSS_FILE" 'body\.is-authenticated \.topbar' "authenticated app must hide the login topbar"

# ---- 命令区：单行头部，无叠层 ----
require_pattern "$INDEX_FILE" 'data-view-command-center' "workspace header must keep its stable QA hook"
require_pattern "$INDEX_FILE" 'id="view-title"' "workspace header must expose the module title"
require_pattern "$INDEX_FILE" 'id="view-description"' "workspace header must expose the module description"
require_pattern "$INDEX_FILE" 'workspace-session-controls' "session status and logout must live inside the workspace actions"
require_absent "$INDEX_FILE" 'workspace-insight' "legacy insight strip must stay removed from the header"
require_absent "$INDEX_FILE" 'view-context' "legacy context chip strip must stay removed from the header"
require_absent "$INDEX_FILE" 'view-rail' "legacy module rail must stay removed from the header"
require_absent "$APP_FILE" 'updateWorkspaceInsight' "legacy insight renderer must stay deleted"
require_absent "$APP_FILE" 'updateViewRail' "legacy rail renderer must stay deleted"
require_absent "$APP_FILE" 'updateViewContext' "legacy context renderer must stay deleted"

# ---- 概览 ----
require_pattern "$INDEX_FILE" 'id="overview-hero"' "overview must expose the first-screen summary strip"
require_pattern "$INDEX_FILE" 'id="overview-insights"' "overview must expose the insight grid"
require_count "$INDEX_FILE" 'data-view-jump-target=' 4 "overview quick cards must deep-link to concrete module targets"
require_pattern "$APP_FILE" 'renderOverviewHero\(' "overview summary must render from live state"
require_pattern "$APP_FILE" 'renderOverviewInsights\(' "overview insights must render from live state"
require_pattern "$APP_FILE" 'getJSON\("/api/delivery/readiness"\)' "overview guide must load delivery readiness"
require_pattern "$APP_FILE" 'guide-action-strip' "overview guide must expose direct real-test actions"
require_pattern "$APP_FILE" 'navigateToDashboardTarget\(' "quick cards must navigate and focus target panels"

# ---- 模块面板与工作台 ----
require_count "$INDEX_FILE" 'data-content-panel=' 9 "core management sections must use the unified content panel shell"
require_count "$INDEX_FILE" 'data-content-panel-meta' 9 "core management sections must expose short panel context"
require_pattern "$APP_FILE" 'renderSourceWorkbench\(' "source list must render its workbench stats"
require_pattern "$APP_FILE" 'renderTokenWorkbench\(' "token list must render its delivery workbench"
require_pattern "$APP_FILE" 'renderIdentityWorkbench\(' "identity module must render its workbench"
require_pattern "$APP_FILE" 'renderNodeWorkbench\(' "node pool must render its workbench"
require_pattern "$APP_FILE" 'renderVirtualNodeWorkbench\(' "virtual gateway list must render its workbench"
require_pattern "$APP_FILE" 'renderPolicyWorkbench\(' "policy list must render its workbench"
require_pattern "$APP_FILE" 'renderTrafficWorkbench\(' "traffic module must render its workbench"
require_pattern "$APP_FILE" 'renderOpsWorkbench\(' "ops module must render its delivery workbench"

# ---- 数据健壮性 ----
require_pattern "$APP_FILE" 'const asList = \(value\) => \(Array\.isArray\(value\) \? value : \[\]\);' "list endpoints must be normalized against JSON null"

# ---- 图标系统 ----
require_pattern "$APP_FILE" 'moduleIconMarkup\(' "module glyphs must render from the shared icon helper"
require_pattern "$APP_FILE" 'actionIconMarkup\(' "action buttons must render from the shared icon helper"
require_pattern "$APP_FILE" 'symbol-fallback' "module glyphs must keep a hidden text fallback"
require_pattern "$APP_FILE" 'button-symbol-fallback' "action icons must keep a hidden text fallback"
require_pattern "$CSS_FILE" '\.symbol-fallback,' "icon text fallbacks must be visually hidden"
require_pattern "$CSS_FILE" 'stroke: currentColor;' "linear icons must inherit the surrounding text color"

# ---- 表单抽屉（侧滑表单页） ----
require_pattern "$APP_FILE" 'syncFormDrawerShellState\(' "form drawers must synchronize the page shell state"
require_pattern "$APP_FILE" 'document\.body\.classList\.toggle\("has-form-drawer-open"' "opening a form drawer must mark the page shell"
require_pattern "$APP_FILE" 'aria-modal' "expanded form drawers must expose dialog semantics"
require_pattern "$CSS_FILE" '\.form-drawer:not\(\.is-collapsed\) \{' "expanded form drawers must be visually governed"
require_pattern "$CSS_FILE" 'body\.has-form-drawer-open::before' "open form sheets must calm the background with a backdrop"

# ---- 可达性 ----
require_pattern "$APP_FILE" 'actionLabelAttrs\(' "icon-only actions must expose hover and accessible labels"
require_pattern "$APP_FILE" 'aria-label="\$\{safeLabel\}" title="\$\{safeLabel\}"' "action labels must be mirrored to aria-label and title"
require_pattern "$APP_FILE" 'aria-label="编辑节点"' "node edit icon button must have an accessible name"

# ---- 移动端 ----
require_pattern "$CSS_FILE" '@media \(max-width: 760px\)' "mobile layout must adapt below 760px"
require_pattern "$CSS_FILE" '\.mobile-dock \{' "mobile dock must be visually governed"
require_pattern "$CSS_FILE" 'var\(--fg-mobile-content-safe\)' "mobile workspace must reserve the dock safe area"
require_pattern "$CSS_FILE" 'var\(--fg-mobile-overlay-bottom\)' "mobile overlays must sit above the dock"

# ---- 截图矩阵 ----
require_pattern "$SCREENSHOT_SCRIPT" 'SCREENSHOT_VIEWPORTS=' "screenshot QA must expose a configurable viewport matrix"
require_pattern "$SCREENSHOT_SCRIPT" '1440x900' "screenshot QA must capture the wide desktop viewport"
require_pattern "$SCREENSHOT_SCRIPT" '1280x720' "screenshot QA must capture the compact desktop viewport"
require_pattern "$SCREENSHOT_SCRIPT" '390x844' "screenshot QA must capture the mobile viewport"
require_pattern "$SCREENSHOT_SCRIPT" 'capture_viewport "\$viewport" \|\| return \$\?' "screenshot QA must stop the matrix on the first failed viewport"

log "static UI contract checks passed"
