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

require_multiline_pattern() {
  local file="$1"
  local pattern="$2"
  local message="$3"
  if ! rg -Uq "$pattern" "$file"; then
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
require_multiline_pattern "$CALM_CSS_FILE" '\.node-card \{[^}]*position: relative;' "node cards must anchor actions without adding empty vertical space"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \{[^}]*position: absolute;' "node card actions must be docked instead of consuming a full card row"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \{[^}]*top: 12px;' "node card actions must stay close to the card title"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \.table-button \{[^}]*flex: 0 0 auto;' "node card action buttons must not stretch into full-width blocks on desktop"
require_pattern "$CALM_CSS_FILE" 'padding-right: 64px;' "node card content must reserve only the compact icon action lane"
require_pattern "$APP_FILE" 'aria-label="编辑节点"' "node edit icon button must have an accessible label"
require_pattern "$APP_FILE" 'title="恢复自动命名"' "node reset icon button must explain the action on hover"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \.button-label \{[^}]*clip-path: inset\(50%\);' "desktop node icon actions must hide truncated text labels visually"
require_multiline_pattern "$CALM_CSS_FILE" '\.node-actions \.table-button \{[^}]*aspect-ratio: 1;' "desktop node icon actions must be stable square controls"

log "static UI contract checks passed"
