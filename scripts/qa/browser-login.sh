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
