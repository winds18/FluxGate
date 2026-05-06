#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/lib/common.sh"

DEPLOY_CONFIG="${DEPLOY_CONFIG:-$ROOT_DIR/.env.deploy.local}"
if [[ -f "$DEPLOY_CONFIG" ]]; then
  # shellcheck disable=SC1090
  source "$DEPLOY_CONFIG"
fi

BASE_URL="${1:-${BASE_URL:-${REMOTE_BROWSER_BASE_URL:-${BROWSER_BASE_URL:-http://127.0.0.1:8080}}}}"
ADMIN_USERNAME="${ADMIN_USERNAME:-${ADMIN_BOOTSTRAP_USERNAME:-${REMOTE_ADMIN_BOOTSTRAP_USERNAME:-}}}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-${ADMIN_BOOTSTRAP_PASSWORD:-${REMOTE_ADMIN_BOOTSTRAP_PASSWORD:-}}}"
KEEP_ARTIFACTS="${KEEP_ARTIFACTS:-false}"
OUT_DIR="${OUT_DIR:-$ROOT_DIR/logs/qa/browser-login/$(timestamp)}"
TEST_FILE="$ROOT_DIR/tmp/browser-login.spec.cjs"
CONFIG_FILE="$ROOT_DIR/tmp/playwright.browser-login.config.cjs"

if [[ -z "$ADMIN_USERNAME" || -z "$ADMIN_PASSWORD" ]]; then
  log "ADMIN_USERNAME and ADMIN_PASSWORD are required for browser login QA"
  exit 64
fi

if [[ -z "${PLAYWRIGHT_BROWSER_CHANNEL:-}" && -d "/Applications/Google Chrome.app" ]]; then
  PLAYWRIGHT_BROWSER_CHANNEL="chrome"
fi

ensure_dir "$ROOT_DIR/tmp"
ensure_dir "$OUT_DIR"

cleanup() {
  rm -f "$TEST_FILE" "$CONFIG_FILE"
  rm -rf "$ROOT_DIR/tmp/playwright-browser-login-output"
  if [[ "$KEEP_ARTIFACTS" != "true" ]]; then
    rm -rf "$OUT_DIR"
  fi
}
trap cleanup EXIT

cat >"$TEST_FILE" <<'JS'
const { expect, test } = require("@playwright/test");

test("admin login reaches dashboard", async ({ page, context }) => {
  const baseURL = process.env.BASE_URL;
  const username = process.env.ADMIN_USERNAME;
  const password = Buffer.from(process.env.LOGIN_CREDENTIAL_B64 || "", "base64").toString("utf8");
  const screenshotPath = process.env.SCREENSHOT_PATH;
  const responses = [];
  const consoleMessages = [];

  page.on("response", (response) => {
    const url = response.url();
    if (url.includes("/api/")) {
      responses.push(`${response.status()} ${url}`);
    }
  });
  page.on("console", (message) => {
    if (message.type() === "error" || message.type() === "warning") {
      consoleMessages.push(`${message.type()}: ${message.text()}`);
    }
  });
  page.on("pageerror", (error) => {
    consoleMessages.push(`pageerror: ${error.message}`);
  });

  await page.goto(baseURL, { waitUntil: "domcontentloaded" });
  await page.waitForSelector("#login-form", { state: "visible", timeout: 15000 });
  await page.fill("#login-username", username);
  await page.fill("#login-password", password);
  await page.click("#login-form button[type=submit]");
  await expect(page.locator("#login-form")).toBeHidden({ timeout: 10000 });
  await expect(page.locator("#app-view")).toBeVisible({ timeout: 10000 });
  await page.waitForTimeout(500);

  const east8TimeSamples = await page.evaluate(() => ({
    rfc3339: typeof formatCell === "function" ? formatCell("2026-04-29T00:00:00Z", "updated_at") : "",
    sqlite: typeof formatCell === "function" ? formatCell("2026-04-29 00:00:00", "created_at") : "",
    hour: typeof formatDateTimeForDisplay === "function" ? formatDateTimeForDisplay("2026-04-29T00:00:00Z") : "",
  }));
  if (
    !east8TimeSamples.rfc3339.includes("2026-04-29 08:00:00") ||
    !east8TimeSamples.sqlite.includes("2026-04-29 08:00:00") ||
    east8TimeSamples.hour !== "2026-04-29 08:00:00"
  ) {
    throw new Error(`east8 time display failed: ${JSON.stringify(east8TimeSamples)}`);
  }
  const trafficTokenRowCount = await page.locator("#traffic-tokens tbody tr").count();
  const quotaMeterCount = await page.locator("#traffic-tokens .quota-meter").count();
  const tokenCardCount = await page.locator("#tokens .token-card").count();
  const tokenRowCount = tokenCardCount || (await page.locator("#tokens tbody tr").count());
  const tokenExtendInputCount = await page.locator("#tokens input[data-token-extend-days]").count();
  const tokenQuotaInputCount = await page.locator("#tokens input[data-token-quota-mib]").count();
  const tokenSubscriptionCopyCount = await page.locator("#tokens button[data-token-action='copy-subscription']").count();
  const tokenRotateSubscriptionCount = await page.locator("#tokens button[data-token-action='rotate-subscription']").count();
  const tokenVisualOverlapCount = await page.evaluate(() => {
    const intersects = (a, b) =>
      a.left < b.right - 1 &&
      a.right > b.left + 1 &&
      a.top < b.bottom - 1 &&
      a.bottom > b.top + 1;
    return Array.from(document.querySelectorAll("#tokens .token-card")).reduce((total, card) => {
      const boxes = Array.from(card.querySelectorAll("button, input, .token-subscription-item code"))
        .filter((element) => element.offsetParent !== null)
        .map((element) => element.getBoundingClientRect())
        .filter((box) => box.width > 0 && box.height > 0);
      for (let left = 0; left < boxes.length; left += 1) {
        for (let right = left + 1; right < boxes.length; right += 1) {
          if (intersects(boxes[left], boxes[right])) {
            total += 1;
          }
        }
      }
      return total;
    }, 0);
  });
  let quotaUsageVisible = false;
  if (trafficTokenRowCount > 0) {
    await expect(page.locator("#traffic-tokens .quota-meter").first()).toBeVisible({ timeout: 5000 });
    if (quotaMeterCount !== trafficTokenRowCount) {
      throw new Error(`quota meter count mismatch: rows=${trafficTokenRowCount} meters=${quotaMeterCount}`);
    }
    quotaUsageVisible = true;
  }

  const teamEditCount = await page.locator("#teams button[data-team-action='edit']").count();
  let teamEditFieldsVisible = false;
  if (teamEditCount > 0) {
    await page.locator("#teams button[data-team-action='edit']").first().click();
    await expect(page.locator('#teams [data-team-field="name"]').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#teams [data-team-field="description"]').first()).toBeVisible({ timeout: 5000 });
    teamEditFieldsVisible = true;
    await page.locator("#teams button[data-team-action='cancel']").first().click();
  }

  const userEditCount = await page.locator("#users button[data-user-action='edit']").count();
  let userEditFieldsVisible = false;
  if (userEditCount > 0) {
    await page.locator("#users button[data-user-action='edit']").first().click();
    await expect(page.locator('#users [data-user-field="name"]').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#users [data-user-field="email"]').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#users [data-user-field="status"]').first()).toBeVisible({ timeout: 5000 });
    userEditFieldsVisible = true;
    await page.locator("#users button[data-user-action='cancel']").first().click();
  }

  const sourceEditCount = await page.locator("#sources button[data-source-action='edit']").count();
  let sourceEditFieldsVisible = false;
  if (sourceEditCount > 0) {
    await page.locator("#sources button[data-source-action='edit']").first().click();
    await expect(page.locator('#sources [data-source-field="name"]').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#sources [data-source-field="display_prefix"]').first()).toBeVisible({ timeout: 5000 });
    sourceEditFieldsVisible = true;
    await page.locator("#sources button[data-source-action='cancel']").first().click();
  }

  const virtualNodeEditCount = await page.locator("#virtual-nodes button[data-virtual-node-action='edit']").count();
  let virtualNodeEditFieldsVisible = false;
  if (virtualNodeEditCount > 0) {
    await page.locator("#virtual-nodes button[data-virtual-node-action='edit']").first().click();
    await expect(page.locator('#virtual-nodes [data-virtual-node-field="name"]').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#virtual-nodes [data-virtual-node-field="listen_port"]').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#virtual-nodes [data-virtual-node-field="tag_selector"]').first()).toBeVisible({ timeout: 5000 });
    virtualNodeEditFieldsVisible = true;
    await page.locator("#virtual-nodes button[data-virtual-node-action='cancel']").first().click();
  }

  const policyEditCount = await page.locator("#policies button[data-policy-action='edit']").count();
  let policyEditFieldsVisible = false;
  if (policyEditCount > 0) {
    await page.locator("#policies button[data-policy-action='edit']").first().click();
    await expect(page.locator('#policies [data-policy-field="name"]').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#policies [data-policy-field="scope_type"]').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#policies [data-policy-field="allowed_virtual_nodes"]').first()).toBeVisible({ timeout: 5000 });
    policyEditFieldsVisible = true;
    await page.locator("#policies button[data-policy-action='cancel']").first().click();
  }

  const nodeRegionCount = await page.locator("#nodes button[data-node-region-action='open']").count();
  let nodeCardCount = 0;
  let nodeVisualOverflowCount = 0;
  if (nodeRegionCount > 0) {
    await page.locator("#nodes button[data-node-region-action='open']").first().click();
    await expect(page.locator("#nodes .node-card").first()).toBeVisible({ timeout: 5000 });
    nodeCardCount = await page.locator("#nodes .node-card").count();
    await page.evaluate(() => {
      const firstTitle = document.querySelector("#nodes .node-card-title");
      if (firstTitle) {
        firstTitle.textContent = "[gougou-joy] 🇦🇷 38阿根廷-联通/移动(AnyTLS)-超长节点名称用于布局验收";
      }
    });
    nodeVisualOverflowCount = await page.evaluate(() => {
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      return Array.from(document.querySelectorAll("#nodes .node-card")).reduce((total, card) => {
        const cardBox = card.getBoundingClientRect();
        const elements = Array.from(
          card.querySelectorAll(".node-card-main, .node-card-title, .node-card-subtitle, .node-card-tags, .node-card-meta, .node-actions"),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const parent = element.classList.contains("node-card-main") || element.classList.contains("node-actions")
            ? cardBox
            : (element.closest(".node-card-main") || card).getBoundingClientRect();
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, parent)) {
            total += 1;
          }
        }
        return total;
      }, 0);
    });
  }

  const nodeEditCount = await page.locator("#nodes button[data-node-action='edit']").count();
  let nodeEditFormVisible = false;
  let nodeDetailVisible = false;
  if (nodeEditCount > 0) {
    await page.locator("#nodes button[data-node-action='edit']").first().click();
    await expect(page.locator("#nodes form[data-node-edit-form]").first()).toBeVisible({ timeout: 5000 });
    nodeEditFormVisible = true;
    await page.locator("#nodes button[data-node-action='cancel']").first().click();
    await page.locator("#nodes .node-card-main").first().click();
    await expect(page.locator("#nodes .node-detail-row .detail-code").first()).toBeVisible({ timeout: 5000 });
    nodeDetailVisible = true;
  }

  const state = await page.evaluate(() => ({
    loginHidden: document.querySelector("#login-view")?.hidden,
    appHidden: document.querySelector("#app-view")?.hidden,
    loginVisible: Boolean(
      document.querySelector("#login-form") &&
        getComputedStyle(document.querySelector("#login-form")).display !== "none" &&
        document.querySelector("#login-form").getClientRects().length > 0,
    ),
    appVisible: Boolean(
      document.querySelector("#app-view") &&
        getComputedStyle(document.querySelector("#app-view")).display !== "none" &&
        document.querySelector("#app-view").getClientRects().length > 0,
    ),
    status: document.querySelector("#status")?.textContent,
    error: document.querySelector("#login-error")?.textContent,
    url: window.location.href,
  }));
  state.teamEditCount = teamEditCount;
  state.teamEditFieldsVisible = teamEditFieldsVisible;
  state.userEditCount = userEditCount;
  state.userEditFieldsVisible = userEditFieldsVisible;
  state.sourceEditCount = sourceEditCount;
  state.sourceEditFieldsVisible = sourceEditFieldsVisible;
  state.virtualNodeEditCount = virtualNodeEditCount;
  state.virtualNodeEditFieldsVisible = virtualNodeEditFieldsVisible;
  state.policyEditCount = policyEditCount;
  state.policyEditFieldsVisible = policyEditFieldsVisible;
  state.nodeRegionCount = nodeRegionCount;
  state.nodeCardCount = nodeCardCount;
  state.nodeVisualOverflowCount = nodeVisualOverflowCount;
  state.nodeEditCount = nodeEditCount;
  state.nodeEditFormVisible = nodeEditFormVisible;
  state.nodeDetailVisible = nodeDetailVisible;
  state.east8TimeSamples = east8TimeSamples;
  state.trafficTokenRowCount = trafficTokenRowCount;
  state.tokenRowCount = tokenRowCount;
  state.tokenCardCount = tokenCardCount;
  state.quotaMeterCount = quotaMeterCount;
  state.quotaUsageVisible = quotaUsageVisible;
  state.tokenExtendInputCount = tokenExtendInputCount;
  state.tokenQuotaInputCount = tokenQuotaInputCount;
  state.tokenSubscriptionCopyCount = tokenSubscriptionCopyCount;
  state.tokenRotateSubscriptionCount = tokenRotateSubscriptionCount;
  state.tokenVisualOverlapCount = tokenVisualOverlapCount;
  if (tokenRowCount > 0 && (tokenExtendInputCount !== tokenRowCount || tokenQuotaInputCount !== tokenRowCount)) {
    throw new Error(`token custom controls missing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && tokenRotateSubscriptionCount !== tokenRowCount) {
    throw new Error(`token subscription rotate controls missing: ${JSON.stringify(state)}`);
  }
  if (tokenVisualOverlapCount > 0) {
    throw new Error(`token controls visually overlap: ${JSON.stringify(state)}`);
  }
  if (nodeVisualOverflowCount > 0) {
    throw new Error(`node cards visually overflow their parent: ${JSON.stringify(state)}`);
  }
  const cookies = await context.cookies(baseURL);
  await page.screenshot({ path: screenshotPath, fullPage: true });

  const result = {
    state,
    cookies: cookies.map((cookie) => ({
      name: cookie.name,
      domain: cookie.domain,
      path: cookie.path,
      secure: cookie.secure,
      httpOnly: cookie.httpOnly,
      sameSite: cookie.sameSite,
    })),
    responses,
    consoleMessages,
  };

  if (!state.loginHidden || state.loginVisible || state.appHidden || !state.appVisible || state.status !== "已连接") {
    throw new Error(JSON.stringify(result, null, 2));
  }

  console.log(JSON.stringify(result, null, 2));
});
JS

cat >"$CONFIG_FILE" <<'JS'
const path = require("path");

module.exports = {
  testDir: __dirname,
  testMatch: "browser-login.spec.cjs",
  outputDir: path.join(__dirname, "playwright-browser-login-output"),
  reporter: [["line"]],
  timeout: 30000,
  use: {
    channel: process.env.PLAYWRIGHT_BROWSER_CHANNEL || undefined,
    headless: true,
    trace: "off",
    video: "off",
    screenshot: "off",
  },
};
JS

log "running browser login QA against $BASE_URL"
LOGIN_CREDENTIAL_B64="$(printf '%s' "$ADMIN_PASSWORD" | base64 | tr -d '\n')"
SCREENSHOT_PATH="$OUT_DIR/fluxgate-login.png" \
  BASE_URL="$BASE_URL" \
  ADMIN_USERNAME="$ADMIN_USERNAME" \
  LOGIN_CREDENTIAL_B64="$LOGIN_CREDENTIAL_B64" \
  CONFIG_FILE="$CONFIG_FILE" \
  PLAYWRIGHT_BROWSER_CHANNEL="${PLAYWRIGHT_BROWSER_CHANNEL:-}" \
  npx --yes -p @playwright/test -c 'export NODE_PATH="$(dirname "$(dirname "$(which playwright)")")${NODE_PATH:+:$NODE_PATH}"; playwright test --config "$CONFIG_FILE"'

if [[ "$KEEP_ARTIFACTS" == "true" ]]; then
  log "browser login QA passed; screenshot: $OUT_DIR/fluxgate-login.png"
else
  log "browser login QA passed; temporary artifacts will be removed"
fi
