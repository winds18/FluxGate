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
  if ! rg -q "$pattern" "$file"; then
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
  count="$(rg -c "$pattern" "$file" || true)"
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
require_pattern "$CALM_CSS_FILE" '\.workspace-header \{' "workspace command header styles are missing"
require_pattern "$CALM_CSS_FILE" 'position: sticky;' "workspace header must keep module actions reachable on desktop"
require_pattern "$CALM_CSS_FILE" 'scroll-snap-type: x proximity;' "module command strips must remain horizontally scannable"
require_pattern "$CALM_CSS_FILE" 'scrollbar-width: none;' "module command strips should hide visual scrollbar noise"
require_pattern "$CALM_CSS_FILE" 'flex: 0 0 auto;' "command chips and rail items must not shrink into unreadable pills"
require_pattern "$CALM_CSS_FILE" '\.workspace-actions \{' "workspace actions must be governed by the calm ops layer"
require_pattern "$CALM_CSS_FILE" 'grid-template-columns: repeat\(2, minmax\(0, 1fr\)\);' "mobile workspace actions must remain balanced and tappable"
require_pattern "$CALM_CSS_FILE" '\.source-actions,' "card action toolbars must be governed by the calm ops layer"
require_pattern "$CALM_CSS_FILE" 'justify-content: flex-end;' "desktop card action toolbars must not look like full-width primary blocks"
require_pattern "$CALM_CSS_FILE" 'flex: 0 1 auto;' "desktop card action buttons must stay compact"
require_pattern "$CALM_CSS_FILE" 'button\[data-source-action="edit"\]' "low-risk inline edit actions must be quiet secondary controls"
require_pattern "$CALM_CSS_FILE" 'button\[data-source-action="save"\]' "inline save actions must keep a clear primary affordance"
require_pattern "$CALM_CSS_FILE" '\.source-actions \.table-button,' "mobile card action bars must retain tappable full-width behavior"

log "static UI contract checks passed"
