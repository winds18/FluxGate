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

// FluxGate UI v3 行为验收：登录、模块切换、表单抽屉、节点详情、移动 Dock。
// 只断言行为与稳定钩子，不锁定像素与文案细节。

const VIEWS = ["overview", "access", "nodes", "identity", "policies", "traffic", "ops"];

test("admin login reaches dashboard and core flows work", async ({ page }) => {
  const baseURL = process.env.BASE_URL;
  const username = process.env.ADMIN_USERNAME;
  const password = Buffer.from(process.env.LOGIN_CREDENTIAL_B64 || "", "base64").toString("utf8");
  const screenshotPath = process.env.SCREENSHOT_PATH;

  let stylesheetStatus = 0;
  const pageErrors = [];
  const consoleErrors = [];
  page.on("response", (response) => {
    if (response.url().endsWith("/assets/styles.css")) {
      stylesheetStatus = response.status();
    }
  });
  page.on("pageerror", (error) => pageErrors.push(String(error.message)));
  page.on("console", (message) => {
    if (message.type() !== "error") return;
    const text = message.text();
    const url = message.location()?.url || "";
    // 未登录时的 401 会话探测是预期行为
    if (text.includes("401") && (url.includes("/api/auth/session") || url === "")) return;
    consoleErrors.push(`${text} (${url})`);
  });

  // 1. 登录
  await page.goto(baseURL, { waitUntil: "domcontentloaded" });
  await page.waitForSelector("#login-form", { state: "visible", timeout: 15000 });
  expect(await page.locator('link[href="/assets/styles.css"]').count()).toBe(1);
  expect(stylesheetStatus).toBe(200);
  await page.fill("#login-username", username);
  await page.fill("#login-password", password);
  await page.click('#login-form button[type="submit"]');
  await page.waitForSelector("#app-view:not([hidden])", { timeout: 15000 });
  await expect(page.locator("#login-view")).toBeHidden();
  await expect(page.locator("body")).toHaveClass(/is-authenticated/);
  await page.waitForFunction(
    () => document.querySelector("#status")?.dataset.statusTone === "success",
    { timeout: 15000 },
  );

  // 2. 七个模块都能切换并渲染
  for (const view of VIEWS) {
    await page.click(`.side-nav [data-view-nav="${view}"]`);
    await expect(page.locator(`[data-dashboard-view="${view}"]`)).toBeVisible();
    const title = (await page.locator("#view-title").textContent())?.trim();
    expect(title?.length).toBeGreaterThan(0);
  }

  // 3. 表单抽屉：展开为对话框、遮罩生效、Escape 收起
  await page.click('.side-nav [data-view-nav="access"]');
  await page.click("#source-form [data-form-drawer-toggle]");
  await expect(page.locator("#source-form")).toHaveAttribute("aria-modal", "true");
  await expect(page.locator("body")).toHaveClass(/has-form-drawer-open/);
  await page.keyboard.press("Escape");
  await expect(page.locator("body")).not.toHaveClass(/has-form-drawer-open/);

  // 4. 节点地区下钻 + 详情抽屉（有节点数据时）
  await page.click('.side-nav [data-view-nav="nodes"]');
  const regionCard = page.locator('[data-node-region-action="open"]').first();
  if (await regionCard.count()) {
    await regionCard.click();
    const detailButton = page.locator('[data-node-action="detail"]').first();
    if (await detailButton.count()) {
      await detailButton.click();
      await expect(page.locator(".node-detail-drawer")).toBeVisible();
      await page.click('[data-node-action="close-detail"]');
      await expect(page.locator(".node-detail-drawer")).toHaveCount(0);
    }
  }

  // 5. 桌面截图
  await page.click('.side-nav [data-view-nav="overview"]');
  if (screenshotPath) {
    await page.screenshot({ path: screenshotPath, fullPage: false });
  }

  // 6. 移动端：侧栏隐藏、Dock 可用、更多菜单可开合
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(page.locator(".dashboard-sidebar")).toBeHidden();
  await expect(page.locator(".mobile-dock")).toBeVisible();
  await page.click("[data-mobile-more-toggle]");
  await expect(page.locator("#mobile-more-menu")).toBeVisible();
  await page.click('#mobile-more-menu [data-view-nav="ops"]');
  await expect(page.locator("#mobile-more-menu")).toBeHidden();
  await expect(page.locator('[data-dashboard-view="ops"]')).toBeVisible();

  // 7. 全程无未预期的脚本错误
  expect(pageErrors, `page errors: ${pageErrors.join("; ")}`).toHaveLength(0);
  expect(consoleErrors, `console errors: ${consoleErrors.join("; ")}`).toHaveLength(0);
});
JS

cat >"$CONFIG_FILE" <<'JS'
const path = require("path");

module.exports = {
  testDir: __dirname,
  testMatch: "browser-login.spec.cjs",
  outputDir: path.join(__dirname, "playwright-browser-login-output"),
  reporter: [["line"]],
  timeout: 60000,
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
