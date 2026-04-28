#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

require_cmd git
require_cmd rg

cd "$ROOT_DIR"

fail=0

report_matches() {
  local label="$1"
  local regex="$2"
  local matches
  matches="$(
    {
      git ls-files -z |
      xargs -0 rg -n --pcre2 --color=never "$regex" 2>/dev/null |
      awk -F: '{print $1 ":" $2}' |
      sort -u
    } || true
  )"

  if [[ -n "$matches" ]]; then
    log "public scan failed: $label"
    printf '%s\n' "$matches"
    fail=1
  fi
}

report_secret_env_matches() {
  local matches
  matches="$(
    {
      git ls-files -z |
      xargs -0 rg -n --pcre2 --color=never '^[[:space:]]*[A-Z0-9_]*(SECRET|TOKEN|PASSWORD|API[_-]?KEY)[A-Z0-9_]*[[:space:]]*=[[:space:]]*["'\'']?(?!change-me|changeme|example|dev-|qa-|<|\$)[^[:space:]#]+' 2>/dev/null |
      rg -v '\$\{' |
      awk -F: '{print $1 ":" $2}' |
      sort -u
    } || true
  )"

  if [[ -n "$matches" ]]; then
    log "public scan failed: non-placeholder secret env assignment"
    printf '%s\n' "$matches"
    fail=1
  fi
}

tracked_env_files="$(git ls-files | rg '(^|/)\.env($|\.)' | rg -v '(^|/)\.env\.example$' || true)"
if [[ -n "$tracked_env_files" ]]; then
  log "public scan failed: tracked env files are not allowed"
  printf '%s\n' "$tracked_env_files"
  fail=1
fi

git_author_email="$(git config user.email || true)"
if [[ -z "$git_author_email" || "$git_author_email" =~ (@localhost$|\.local$|\.lan$) ]]; then
  log "public scan failed: configure a non-local Git user.email before committing"
  fail=1
fi

report_matches "private host or local home paths" '/(home|Users)/wings'
report_matches "private SSH alias" '(^|[^0-9A-Za-z_.-])66\.10([^0-9.]|$)|ssh[[:space:]]+66\.10'
report_matches "private key material" '-----BEGIN [A-Z ]*PRIVATE KEY-----'
report_matches "GitHub access token pattern" 'github_pat_[A-Za-z0-9_]+|gh[pousr]_[A-Za-z0-9_]{20,}'
report_matches "OpenAI-style API key pattern" 'sk-[A-Za-z0-9_-]{20,}'
report_secret_env_matches

if [[ "$fail" -ne 0 ]]; then
  log "public scan failed; keep sensitive values in ignored local config, server .env, or secret stores"
  exit 1
fi

log "public scan passed"
