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
  const expectedViewSymbols = {
    overview: "概",
    access: "源",
    nodes: "点",
    identity: "身",
    policies: "策",
    traffic: "量",
    ops: "运",
  };

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
  await page.waitForFunction(() => {
    const status = document.querySelector("#status")?.textContent?.trim();
    return Boolean(status && !["登录中", "已登录", "刷新中"].includes(status));
  }, null, { timeout: 15000 });
  const dashboardLoadStatus = (await page.locator("#status").textContent())?.trim();
  if (dashboardLoadStatus !== "已连接") {
    const loadDiagnostics = await page.evaluate(() => ({
      status: document.querySelector("#status")?.textContent?.trim(),
      metricsText: document.querySelector("#metrics")?.textContent?.trim(),
    }));
    throw new Error(
      `dashboard data did not load: ${JSON.stringify({
        ...loadDiagnostics,
        responses,
        consoleMessages,
      })}`,
    );
  }

  const viewNames = ["overview", "access", "nodes", "identity", "policies", "traffic", "ops"];
  const switchView = async (view) => {
    const sidebarButton = page.locator(`.dashboard-sidebar [data-view-nav="${view}"]`).first();
    if (await sidebarButton.isVisible()) {
      await sidebarButton.click();
    } else {
      await page.locator(`.mobile-dock [data-view-nav="${view}"]`).first().click();
    }
    await expect(page.locator(`[data-dashboard-view="${view}"]`)).toBeVisible({ timeout: 5000 });
    const activeNavCount = await page.locator(`[data-view-nav="${view}"].is-active`).count();
    if (activeNavCount === 0) {
      throw new Error(`active navigation missing for ${view}`);
    }
  };
  const pageHorizontalOverflow = async () =>
    page.evaluate(() => Math.max(0, document.documentElement.scrollWidth - window.innerWidth));
  const dashboardNavCount = await page.locator(".dashboard-sidebar [data-view-nav]").count();
  const dashboardNavSymbolCount = await page.locator(".dashboard-sidebar .nav-symbol").count();
  const dashboardNavBadgeCount = await page.locator(".dashboard-sidebar [data-view-count]").count();
  const mobileDockCount = await page.locator(".mobile-dock [data-view-nav]").count();
  const mobileDockSymbolCount = await page.locator(".mobile-dock .mobile-dock-symbol").count();
  const mobileDockBadgeCount = await page.locator(".mobile-dock [data-view-count]").count();
  let populatedNavBadgeCount = 0;
  let navBadgeValues = {};
  const dashboardViewCount = await page.locator("[data-dashboard-view]").count();
  const formDrawerCount = await page.locator("[data-form-drawer]").count();
  const collapsedFormDrawerCount = await page.locator("[data-form-drawer].is-collapsed").count();
  const formDrawerSymbolCount = await page.locator("[data-form-drawer] .drawer-symbol").count();
  const formDrawerSymbols = await page
    .locator("[data-form-drawer] .drawer-symbol")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const visibleFormDrawerHeaderOverflow = async () => page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("[data-form-drawer] .panel-header"))
      .filter((header) => header.offsetParent !== null)
      .reduce((total, header) => {
        const headerBox = header.getBoundingClientRect();
        const elements = Array.from(header.querySelectorAll(".drawer-title, .drawer-symbol, h2, .drawer-toggle")).filter(
          (element) => element.offsetParent !== null,
        );
        for (const element of elements) {
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, headerBox)) {
            total += 1;
          }
        }
        return total;
      }, 0);
  });
  const panelCountCount = await page.locator(".panel-count").count();
  const populatedPanelCountCount = await page
    .locator(".panel-count")
    .evaluateAll((elements) => elements.filter((element) => element.textContent.trim().length > 0).length);
  const panelSymbolCount = await page.locator(".panel-title .panel-symbol").count();
  const panelSymbols = await page
    .locator(".panel-title .panel-symbol")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const viewOverflow = {};
  const formDrawerHeaderOverflow = {};
  const panelTitleOverflow = {};
  const viewContextChipCounts = {};
  const viewContextOverflow = {};
  const viewRailButtonCounts = {};
  const viewRailSymbolCounts = {};
  const viewRailSymbols = {};
  const viewRailBadgeCounts = {};
  const viewRailOverflow = {};
  const viewHeaderSymbols = {};
  for (const view of viewNames) {
    await switchView(view);
    viewOverflow[view] = await pageHorizontalOverflow();
    formDrawerHeaderOverflow[view] = await visibleFormDrawerHeaderOverflow();
    viewHeaderSymbols[view] = (await page.locator("#view-symbol").textContent())?.trim();
    viewContextChipCounts[view] = await page.locator("#view-context .context-chip").count();
    viewContextOverflow[view] = await page.evaluate(() => {
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      return Array.from(document.querySelectorAll("#view-context .context-chip")).reduce((total, chip) => {
        const chipBox = chip.getBoundingClientRect();
        const elements = Array.from(chip.querySelectorAll("span, strong")).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, chipBox)) {
            total += 1;
          }
        }
        return total;
      }, 0);
    });
    viewRailButtonCounts[view] = await page.locator("#view-rail .view-rail-button").count();
    viewRailSymbolCounts[view] = await page.locator("#view-rail .view-rail-symbol").count();
    viewRailSymbols[view] = await page
      .locator("#view-rail .view-rail-symbol")
      .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
    viewRailBadgeCounts[view] = await page.locator("#view-rail [data-view-rail-count]").count();
    viewRailOverflow[view] = await page.evaluate(() => {
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      return Array.from(document.querySelectorAll("#view-rail .view-rail-button")).reduce((total, button) => {
        const buttonBox = button.getBoundingClientRect();
        const elements = Array.from(button.querySelectorAll(".view-rail-symbol, .view-rail-label, .view-rail-count")).filter(
          (element) => element.offsetParent !== null,
        );
        for (const element of elements) {
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, buttonBox)) {
            total += 1;
          }
        }
        return total;
      }, 0);
    });
    panelTitleOverflow[view] = await page.evaluate((viewName) => {
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      return Array.from(document.querySelectorAll(`[data-dashboard-view="${viewName}"] .panel-title`)).reduce((total, title) => {
        const titleBox = title.getBoundingClientRect();
        const elements = Array.from(title.querySelectorAll(".panel-symbol, h2, .panel-count"))
          .filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, titleBox)) {
            total += 1;
          }
        }
        return total;
      }, 0);
    }, view);
  }
  await switchView("overview");
  await expect(page.locator("#metrics .metric")).toHaveCount(7, { timeout: 5000 });
  const overviewMetricCount = await page.locator("#metrics .metric").count();
  const overviewMetricSymbolCount = await page.locator("#metrics .metric-symbol").count();
  const overviewMetricValueCount = await page.locator("#metrics .metric-value").count();
  const overviewMetricOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#metrics .metric")).reduce((total, metric) => {
      const metricBox = metric.getBoundingClientRect();
      const elements = Array.from(metric.querySelectorAll(".metric-heading, .metric-symbol, .metric-label, .metric-value"))
        .filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, metricBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
  const overviewReadinessCount = await page.locator("#overview-readiness .readiness-item").count();
  const overviewReadinessIndexCount = await page.locator("#overview-readiness .readiness-index").count();
  const overviewReadinessSymbolCount = await page.locator("#overview-readiness .readiness-symbol").count();
  const overviewReadinessStateCount = await page.locator("#overview-readiness .readiness-state").count();
  const overviewReadinessStates = await page
    .locator("#overview-readiness .readiness-state")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const overviewReadinessSymbols = await page
    .locator("#overview-readiness .readiness-symbol")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const overviewReadinessOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#overview-readiness .readiness-item")).reduce((total, item) => {
      const itemBox = item.getBoundingClientRect();
      const elements = Array.from(
        item.querySelectorAll(
          ".readiness-symbol-wrap, .readiness-index, .readiness-symbol, .readiness-dot, .readiness-body, .readiness-title-row, .readiness-state, strong, small",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, itemBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
  const overviewNextStepButtonCount = await page.locator("#overview-next-step [data-overview-jump]").count();
  const overviewNextStepCardCount = await page.locator("#overview-next-step [data-overview-next-step-card]").count();
  const overviewNextStepActionCount = await page.locator("#overview-next-step .next-step-action").count();
  const overviewNextStepOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const card = document.querySelector("#overview-next-step [data-overview-next-step-card]");
    if (!card) return 1;
    const cardBox = card.getBoundingClientRect();
    return Array.from(card.querySelectorAll(".next-step-symbol, .next-step-body, .next-step-action"))
      .filter((element) => element.offsetParent !== null)
      .reduce((total, element) => {
        const box = element.getBoundingClientRect();
        return box.width > 0 && box.height > 0 && outside(box, cardBox) ? total + 1 : total;
      }, 0);
  });
  const overviewQuickCardCount = await page.locator(".quick-card").count();
  const overviewQuickCardSymbolCount = await page.locator(".quick-card .quick-card-symbol").count();
  const overviewQuickCardSymbols = await page
    .locator(".quick-card .quick-card-symbol")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const overviewQuickCardBadgeCount = await page.locator(".quick-card [data-overview-card-count]").count();
  const overviewQuickCardBadgeValues = await page
    .locator(".quick-card [data-overview-card-count]")
    .evaluateAll((elements) => Object.fromEntries(elements.map((element) => [element.dataset.overviewCardCount, element.textContent.trim()])));
  const overviewQuickCardOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll(".quick-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(card.querySelectorAll(".quick-card-heading, .quick-card-symbol, strong, .quick-card-badge"))
        .filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, cardBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });

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
  await switchView("traffic");
  const trafficTokenTableRowCount = await page.locator("#traffic-tokens tbody tr").count();
  const trafficTokenCardCount = await page.locator("#traffic-tokens .traffic-token-card").count();
  const trafficTokenRowCount = trafficTokenCardCount || trafficTokenTableRowCount;
  const trafficOutboundTableRowCount = await page.locator("#traffic-outbounds tbody tr").count();
  const trafficOutboundCardCount = await page.locator("#traffic-outbounds .traffic-outbound-card").count();
  const trafficOutboundRowCount = trafficOutboundCardCount || trafficOutboundTableRowCount;
  const quotaMeterCount = await page.locator("#traffic-tokens .quota-meter").count();
  const trafficVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#traffic-tokens .traffic-card, #traffic-outbounds .traffic-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(".traffic-card-heading, .traffic-card-meter, .traffic-card-field"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, cardBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
  let quotaUsageVisible = false;
  if (trafficTokenRowCount > 0) {
    await expect(page.locator("#traffic-tokens .quota-meter").first()).toBeVisible({ timeout: 5000 });
    if (quotaMeterCount !== trafficTokenRowCount) {
      throw new Error(`quota meter count mismatch: cards=${trafficTokenRowCount} meters=${quotaMeterCount}`);
    }
    quotaUsageVisible = true;
  }

  await switchView("identity");
  let moduleRailOpensTokenForm = false;
  const tokenFormRailButton = page.locator('#view-rail [data-view-rail-target="token-form"]').first();
  if (await tokenFormRailButton.isVisible()) {
    await tokenFormRailButton.click();
    await expect(page.locator("#token-form")).not.toHaveClass(/is-collapsed/, { timeout: 5000 });
    moduleRailOpensTokenForm = true;
  }
  await page.evaluate(() => {
    if (typeof showTokenSubscriptionResult !== "function") return;
    showTokenSubscriptionResult(
      {
        id: 999,
        subscription: "https://fluxgate.example.test/sub/fg_browser_qa_default_subscription_token",
        subscriptions: {
          clash: "https://fluxgate.example.test/sub/fg_browser_qa_default_subscription_token?target=clash",
          sing_box: "https://fluxgate.example.test/sub/fg_browser_qa_default_subscription_token?target=sing-box",
        },
      },
      "订阅地址预览",
    );
  });
  await expect(page.locator("#token-result .token-result-card")).toBeVisible({ timeout: 5000 });
  const tokenResultSubscriptionItemCount = await page.locator("#token-result .token-subscription-item").count();
  const tokenResultSubscriptionKindCount = await page.locator("#token-result [data-token-subscription-kind]").count();
  const tokenResultSubscriptionOpenCount = await page.locator("#token-result a[data-token-subscription-link]").count();
  const tokenResultSubscriptionActionSymbolCount = await page.locator("#token-result .token-subscription-actions .button-symbol").count();
  const tokenResultVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const resultCard = document.querySelector("#token-result .token-result-card");
    if (!resultCard) return 1;
    const resultBox = resultCard.getBoundingClientRect();
    return Array.from(
      resultCard.querySelectorAll(".token-result-heading, .token-subscription-item, .token-subscription-heading, .token-subscription-kind, code, button, a"),
    ).reduce((total, element) => {
      if (element.offsetParent === null) return total;
      const box = element.getBoundingClientRect();
      if (box.width > 0 && box.height > 0 && outside(box, resultBox)) {
        return total + 1;
      }
      return total;
    }, 0);
  });
  const tokenCardCount = await page.locator("#tokens .token-card").count();
  const tokenRowCount = tokenCardCount || (await page.locator("#tokens tbody tr").count());
  const tokenExtendInputCount = await page.locator("#tokens input[data-token-extend-days]").count();
  const tokenQuotaInputCount = await page.locator("#tokens input[data-token-quota-mib]").count();
  const tokenActionFieldCount = await page.locator("#tokens .token-action-field").count();
  const tokenQuotaMeterCount = await page.locator("#tokens [data-token-quota-meter] .quota-meter").count();
  const tokenSubscriptionItemCount = await page.locator("#tokens .token-subscription-item").count();
  const tokenSubscriptionKindCount = await page.locator("#tokens [data-token-subscription-kind]").count();
  const tokenSubscriptionCopyCount = await page.locator("#tokens button[data-token-action='copy-subscription']").count();
  const tokenSubscriptionOpenCount = await page.locator("#tokens a[data-token-subscription-link]").count();
  const tokenSubscriptionActionSymbolCount = await page.locator("#tokens .token-subscription-actions .button-symbol").count();
  const tokenRotateSubscriptionCount = await page.locator("#tokens button[data-token-action='rotate-subscription']").count();
  let tokenCopyFeedbackVisible = false;
  let tokenCopySymbolRestored = false;
  if (tokenSubscriptionCopyCount > 0) {
    const firstCopyButton = page.locator("#tokens button[data-token-action='copy-subscription']").first();
    await firstCopyButton.click();
    await expect(firstCopyButton).toHaveText("已复制", { timeout: 5000 });
    await expect(page.locator("#status")).toContainText("订阅地址已复制", { timeout: 5000 });
    tokenCopyFeedbackVisible = true;
    await page.waitForTimeout(1700);
    await expect(firstCopyButton.locator(".button-symbol")).toHaveText("⧉", { timeout: 5000 });
    await expect(firstCopyButton.locator(".button-label")).toHaveText("复制", { timeout: 5000 });
    tokenCopySymbolRestored = true;
  }
  const tokenVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#tokens .token-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(
          ".token-card-heading, .token-card-field, .token-card-meter, .token-card-subscriptions, .token-subscription-item, .token-subscription-heading, .token-subscription-kind, .token-card-actions",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, cardBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
  const tokenVisualOverlapCount = await page.evaluate(() => {
    const intersects = (a, b) =>
      a.left < b.right - 1 &&
      a.right > b.left + 1 &&
      a.top < b.bottom - 1 &&
      a.bottom > b.top + 1;
    return Array.from(document.querySelectorAll("#tokens .token-card")).reduce((total, card) => {
      const boxes = Array.from(card.querySelectorAll("button, a, input, .token-subscription-item code"))
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

  const teamCardCount = await page.locator("#teams .identity-card").count();
  const userCardCount = await page.locator("#users .identity-card").count();
  const identityVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#teams .identity-card, #users .identity-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(".identity-card-heading, .identity-card-badge, .identity-card-field, .identity-card-actions"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, cardBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
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

  await switchView("access");
  const sourceCardCount = await page.locator("#sources .source-card").count();
  const sourceVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#sources .source-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(".source-card-heading, .source-type-badge, .source-card-field, .source-actions"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, cardBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
  const sourceEditCount = await page.locator("#sources button[data-source-action='edit']").count();
  let sourceEditFieldsVisible = false;
  if (sourceEditCount > 0) {
    await page.locator("#sources button[data-source-action='edit']").first().click();
    await expect(page.locator('#sources [data-source-field="name"]').first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#sources [data-source-field="display_prefix"]').first()).toBeVisible({ timeout: 5000 });
    sourceEditFieldsVisible = true;
    await page.locator("#sources button[data-source-action='cancel']").first().click();
  }

  await switchView("nodes");
  const virtualNodeCardCount = await page.locator("#virtual-nodes .virtual-node-card").count();
  const virtualNodeVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#virtual-nodes .virtual-node-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(".virtual-node-card-heading, .virtual-node-listen-badge, .virtual-node-card-field, .virtual-node-actions"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, cardBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
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

  await switchView("policies");
  const policyCardCount = await page.locator("#policies .policy-card").count();
  const policyVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#policies .policy-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(".policy-card-heading, .policy-scope-badge, .policy-card-field, .policy-actions"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, cardBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
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

  await switchView("nodes");
  const nodeSearchVisible = await page.locator("#nodes [data-node-filter]").first().isVisible();
  await page.fill("#nodes [data-node-filter]", "香港");
  await page.waitForTimeout(100);
  const nodeFilteredRegionCount = await page.locator("#nodes button[data-node-region-action='open']").count();
  const nodeFilterValue = await page.locator("#nodes [data-node-filter]").first().inputValue();
  await page.locator("#nodes button[data-node-filter-action='clear']").first().click();
  await expect(page.locator("#nodes [data-node-filter]").first()).toHaveValue("", { timeout: 5000 });
  const nodeSearchClears = (await page.locator("#nodes [data-node-filter]").first().inputValue()) === "";
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
  let nodeDetailCodeOverflowCount = 0;
  let nodeDetailCopyVisible = false;
  let nodeDetailCopyFeedbackVisible = false;
  if (nodeEditCount > 0) {
    await page.locator("#nodes button[data-node-action='edit']").first().click();
    await expect(page.locator("#nodes form[data-node-edit-form]").first()).toBeVisible({ timeout: 5000 });
    nodeEditFormVisible = true;
    await page.locator("#nodes button[data-node-action='cancel']").first().click();
    await page.locator("#nodes .node-card-main").first().click();
    await expect(page.locator("#nodes .node-detail-row .detail-code").first()).toBeVisible({ timeout: 5000 });
    nodeDetailVisible = true;
    const nodeDetailCopyButton = page.locator("#nodes button[data-node-action='copy-uri']").first();
    await expect(nodeDetailCopyButton).toBeVisible({ timeout: 5000 });
    nodeDetailCopyVisible = true;
    await nodeDetailCopyButton.click();
    await expect(nodeDetailCopyButton).toHaveText("已复制", { timeout: 5000 });
    await expect(page.locator("#status")).toContainText("节点 URI 已复制", { timeout: 5000 });
    nodeDetailCopyFeedbackVisible = true;
    nodeDetailCodeOverflowCount = await page.evaluate(() =>
      Array.from(document.querySelectorAll("#nodes .node-detail-row .detail-code")).filter(
        (element) => element.scrollWidth > element.clientWidth + 2,
      ).length,
    );
  }

  await switchView("ops");
  const opsActionCardCount = await page.locator("#ops-actions .ops-action-card").count();
  await page.locator("#config-check").click();
  await expect(page.locator("#config-check-result .ops-result-card")).toBeVisible({ timeout: 5000 });
  const opsResultVisible = await page.locator("#config-check-result .ops-result-card").isVisible();
  const opsResultFieldCount = await page.locator("#config-check-result .ops-result-field").count();
  const opsVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#ops-actions .ops-action-card, #config-check-result .ops-result-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll("button, .ops-result-heading, .ops-result-field, code, strong"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, cardBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });

  await page.setViewportSize({ width: 390, height: 844 });
  await switchView("overview");
  const mobileDockVisible = await page.locator(".mobile-dock").isVisible();
  const mobileSidebarVisible = await page.locator(".dashboard-sidebar").isVisible();
  const mobileOverviewOverflow = await pageHorizontalOverflow();
  const mobileDockOverflow = await page.locator(".mobile-dock").evaluate((dock) =>
    Math.max(0, dock.scrollWidth - dock.clientWidth),
  );
  await switchView("policies");
  const mobilePoliciesVisible = await page.locator('[data-dashboard-view="policies"]').isVisible();
  const mobilePoliciesOverflow = await pageHorizontalOverflow();
  await switchView("nodes");
  const mobileNodesVisible = await page.locator('[data-dashboard-view="nodes"]').isVisible();
  const mobileNodesOverflow = await pageHorizontalOverflow();
  await switchView("ops");
  const mobileOpsVisible = await page.locator('[data-dashboard-view="ops"]').isVisible();
  const mobileOpsOverflow = await pageHorizontalOverflow();
  await page.setViewportSize({ width: 1280, height: 720 });
  await switchView("nodes");
  populatedNavBadgeCount = await page
    .locator("[data-view-count]")
    .evaluateAll((elements) => elements.filter((element) => element.textContent.trim().length > 0 && element.textContent.trim() !== "--").length);
  navBadgeValues = await page
    .locator(".dashboard-sidebar [data-view-count]")
    .evaluateAll((elements) => Object.fromEntries(elements.map((element) => [element.dataset.viewCount, element.textContent.trim()])));

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
    viewTitle: document.querySelector("#view-title")?.textContent,
  }));
  state.dashboardNavCount = dashboardNavCount;
  state.dashboardNavSymbolCount = dashboardNavSymbolCount;
  state.dashboardNavBadgeCount = dashboardNavBadgeCount;
  state.mobileDockCount = mobileDockCount;
  state.mobileDockSymbolCount = mobileDockSymbolCount;
  state.mobileDockBadgeCount = mobileDockBadgeCount;
  state.populatedNavBadgeCount = populatedNavBadgeCount;
  state.navBadgeValues = navBadgeValues;
  state.dashboardViewCount = dashboardViewCount;
  state.formDrawerCount = formDrawerCount;
  state.collapsedFormDrawerCount = collapsedFormDrawerCount;
  state.formDrawerSymbolCount = formDrawerSymbolCount;
  state.formDrawerSymbols = formDrawerSymbols;
  state.formDrawerHeaderOverflow = formDrawerHeaderOverflow;
  state.panelCountCount = panelCountCount;
  state.populatedPanelCountCount = populatedPanelCountCount;
  state.panelSymbolCount = panelSymbolCount;
  state.panelSymbols = panelSymbols;
  state.panelTitleOverflow = panelTitleOverflow;
  state.viewContextChipCounts = viewContextChipCounts;
  state.viewContextOverflow = viewContextOverflow;
  state.viewRailButtonCounts = viewRailButtonCounts;
  state.viewRailSymbolCounts = viewRailSymbolCounts;
  state.viewRailSymbols = viewRailSymbols;
  state.viewRailBadgeCounts = viewRailBadgeCounts;
  state.viewRailOverflow = viewRailOverflow;
  state.viewHeaderSymbols = viewHeaderSymbols;
  state.moduleRailOpensTokenForm = moduleRailOpensTokenForm;
  state.overviewMetricCount = overviewMetricCount;
  state.overviewMetricSymbolCount = overviewMetricSymbolCount;
  state.overviewMetricValueCount = overviewMetricValueCount;
  state.overviewMetricOverflowCount = overviewMetricOverflowCount;
  state.overviewReadinessCount = overviewReadinessCount;
  state.overviewReadinessIndexCount = overviewReadinessIndexCount;
  state.overviewReadinessSymbolCount = overviewReadinessSymbolCount;
  state.overviewReadinessStateCount = overviewReadinessStateCount;
  state.overviewReadinessStates = overviewReadinessStates;
  state.overviewReadinessSymbols = overviewReadinessSymbols;
  state.overviewReadinessOverflowCount = overviewReadinessOverflowCount;
  state.overviewNextStepButtonCount = overviewNextStepButtonCount;
  state.overviewNextStepCardCount = overviewNextStepCardCount;
  state.overviewNextStepActionCount = overviewNextStepActionCount;
  state.overviewNextStepOverflowCount = overviewNextStepOverflowCount;
  state.overviewQuickCardCount = overviewQuickCardCount;
  state.overviewQuickCardSymbolCount = overviewQuickCardSymbolCount;
  state.overviewQuickCardSymbols = overviewQuickCardSymbols;
  state.overviewQuickCardBadgeCount = overviewQuickCardBadgeCount;
  state.overviewQuickCardBadgeValues = overviewQuickCardBadgeValues;
  state.overviewQuickCardOverflowCount = overviewQuickCardOverflowCount;
  state.viewOverflow = viewOverflow;
  state.mobileDockVisible = mobileDockVisible;
  state.mobileSidebarVisible = mobileSidebarVisible;
  state.mobileOverviewOverflow = mobileOverviewOverflow;
  state.mobileDockOverflow = mobileDockOverflow;
  state.mobilePoliciesVisible = mobilePoliciesVisible;
  state.mobilePoliciesOverflow = mobilePoliciesOverflow;
  state.mobileNodesVisible = mobileNodesVisible;
  state.mobileNodesOverflow = mobileNodesOverflow;
  state.mobileOpsVisible = mobileOpsVisible;
  state.mobileOpsOverflow = mobileOpsOverflow;
  state.teamCardCount = teamCardCount;
  state.teamEditCount = teamEditCount;
  state.teamEditFieldsVisible = teamEditFieldsVisible;
  state.userCardCount = userCardCount;
  state.userEditCount = userEditCount;
  state.userEditFieldsVisible = userEditFieldsVisible;
  state.identityVisualOverflowCount = identityVisualOverflowCount;
  state.sourceCardCount = sourceCardCount;
  state.sourceVisualOverflowCount = sourceVisualOverflowCount;
  state.sourceEditCount = sourceEditCount;
  state.sourceEditFieldsVisible = sourceEditFieldsVisible;
  state.virtualNodeCardCount = virtualNodeCardCount;
  state.virtualNodeVisualOverflowCount = virtualNodeVisualOverflowCount;
  state.virtualNodeEditCount = virtualNodeEditCount;
  state.virtualNodeEditFieldsVisible = virtualNodeEditFieldsVisible;
  state.policyCardCount = policyCardCount;
  state.policyVisualOverflowCount = policyVisualOverflowCount;
  state.policyEditCount = policyEditCount;
  state.policyEditFieldsVisible = policyEditFieldsVisible;
  state.nodeSearchVisible = nodeSearchVisible;
  state.nodeFilterValue = nodeFilterValue;
  state.nodeFilteredRegionCount = nodeFilteredRegionCount;
  state.nodeSearchClears = nodeSearchClears;
  state.nodeRegionCount = nodeRegionCount;
  state.nodeCardCount = nodeCardCount;
  state.nodeVisualOverflowCount = nodeVisualOverflowCount;
  state.nodeEditCount = nodeEditCount;
  state.nodeEditFormVisible = nodeEditFormVisible;
  state.nodeDetailVisible = nodeDetailVisible;
  state.nodeDetailCopyVisible = nodeDetailCopyVisible;
  state.nodeDetailCopyFeedbackVisible = nodeDetailCopyFeedbackVisible;
  state.nodeDetailCodeOverflowCount = nodeDetailCodeOverflowCount;
  state.east8TimeSamples = east8TimeSamples;
  state.trafficTokenRowCount = trafficTokenRowCount;
  state.trafficTokenCardCount = trafficTokenCardCount;
  state.trafficOutboundRowCount = trafficOutboundRowCount;
  state.trafficOutboundCardCount = trafficOutboundCardCount;
  state.trafficVisualOverflowCount = trafficVisualOverflowCount;
  state.tokenRowCount = tokenRowCount;
  state.tokenCardCount = tokenCardCount;
  state.tokenResultSubscriptionItemCount = tokenResultSubscriptionItemCount;
  state.tokenResultSubscriptionKindCount = tokenResultSubscriptionKindCount;
  state.tokenResultSubscriptionOpenCount = tokenResultSubscriptionOpenCount;
  state.tokenResultSubscriptionActionSymbolCount = tokenResultSubscriptionActionSymbolCount;
  state.tokenResultVisualOverflowCount = tokenResultVisualOverflowCount;
  state.quotaMeterCount = quotaMeterCount;
  state.quotaUsageVisible = quotaUsageVisible;
  state.tokenExtendInputCount = tokenExtendInputCount;
  state.tokenQuotaInputCount = tokenQuotaInputCount;
  state.tokenActionFieldCount = tokenActionFieldCount;
  state.tokenQuotaMeterCount = tokenQuotaMeterCount;
  state.tokenSubscriptionItemCount = tokenSubscriptionItemCount;
  state.tokenSubscriptionKindCount = tokenSubscriptionKindCount;
  state.tokenSubscriptionCopyCount = tokenSubscriptionCopyCount;
  state.tokenSubscriptionOpenCount = tokenSubscriptionOpenCount;
  state.tokenSubscriptionActionSymbolCount = tokenSubscriptionActionSymbolCount;
  state.tokenCopyFeedbackVisible = tokenCopyFeedbackVisible;
  state.tokenCopySymbolRestored = tokenCopySymbolRestored;
  state.tokenRotateSubscriptionCount = tokenRotateSubscriptionCount;
  state.tokenVisualOverflowCount = tokenVisualOverflowCount;
  state.tokenVisualOverlapCount = tokenVisualOverlapCount;
  state.opsActionCardCount = opsActionCardCount;
  state.opsResultVisible = opsResultVisible;
  state.opsResultFieldCount = opsResultFieldCount;
  state.opsVisualOverflowCount = opsVisualOverflowCount;
  if (tokenRowCount > 0 && (tokenExtendInputCount !== tokenRowCount || tokenQuotaInputCount !== tokenRowCount)) {
    throw new Error(`token custom controls missing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && tokenActionFieldCount !== tokenRowCount * 2) {
    throw new Error(`token action labels missing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && tokenQuotaMeterCount !== tokenRowCount) {
    throw new Error(`token quota meters missing: ${JSON.stringify(state)}`);
  }
  if (
    tokenResultSubscriptionItemCount !== 3 ||
    tokenResultSubscriptionKindCount !== 3 ||
    tokenResultSubscriptionOpenCount !== 3 ||
    tokenResultSubscriptionActionSymbolCount !== 6 ||
    tokenResultVisualOverflowCount > 0
  ) {
    throw new Error(`token subscription result card is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && tokenRotateSubscriptionCount !== tokenRowCount) {
    throw new Error(`token subscription rotate controls missing: ${JSON.stringify(state)}`);
  }
  if (
    tokenSubscriptionCopyCount > 0 &&
    (tokenSubscriptionItemCount !== tokenSubscriptionCopyCount ||
      tokenSubscriptionKindCount !== tokenSubscriptionCopyCount ||
      tokenSubscriptionOpenCount !== tokenSubscriptionCopyCount ||
      tokenSubscriptionActionSymbolCount !== tokenSubscriptionCopyCount * 2)
  ) {
    throw new Error(`token subscription cards are incomplete: ${JSON.stringify(state)}`);
  }
  if (tokenSubscriptionCopyCount > 0 && !tokenCopyFeedbackVisible) {
    throw new Error(`token subscription copy feedback missing: ${JSON.stringify(state)}`);
  }
  if (tokenSubscriptionCopyCount > 0 && !tokenCopySymbolRestored) {
    throw new Error(`token subscription copy symbol restore missing: ${JSON.stringify(state)}`);
  }
  if (tokenVisualOverflowCount > 0 || tokenVisualOverlapCount > 0) {
    throw new Error(`token controls visually overflow or overlap: ${JSON.stringify(state)}`);
  }
  if (
    (trafficTokenRowCount > 0 && trafficTokenCardCount !== trafficTokenRowCount) ||
    (trafficOutboundRowCount > 0 && trafficOutboundCardCount !== trafficOutboundRowCount) ||
    trafficVisualOverflowCount > 0
  ) {
    throw new Error(`traffic cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    (teamEditCount > 0 && teamCardCount !== teamEditCount) ||
    (userEditCount > 0 && userCardCount !== userEditCount) ||
    identityVisualOverflowCount > 0
  ) {
    throw new Error(`identity cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (sourceEditCount > 0 && (sourceCardCount !== sourceEditCount || sourceVisualOverflowCount > 0)) {
    throw new Error(`source cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (virtualNodeEditCount > 0 && (virtualNodeCardCount !== virtualNodeEditCount || virtualNodeVisualOverflowCount > 0)) {
    throw new Error(`virtual node cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (policyEditCount > 0 && (policyCardCount !== policyEditCount || policyVisualOverflowCount > 0)) {
    throw new Error(`policy cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (nodeVisualOverflowCount > 0) {
    throw new Error(`node cards visually overflow their parent: ${JSON.stringify(state)}`);
  }
  if (!nodeSearchVisible || nodeFilterValue !== "香港" || nodeFilteredRegionCount < 1 || !nodeSearchClears) {
    throw new Error(`node search filter failed: ${JSON.stringify(state)}`);
  }
  if (nodeDetailCodeOverflowCount > 0) {
    throw new Error(`node detail code blocks overflow horizontally: ${JSON.stringify(state)}`);
  }
  if (nodeDetailVisible && (!nodeDetailCopyVisible || !nodeDetailCopyFeedbackVisible)) {
    throw new Error(`node detail copy feedback missing: ${JSON.stringify(state)}`);
  }
  if (opsActionCardCount !== 4 || !opsResultVisible || opsResultFieldCount < 4 || opsVisualOverflowCount > 0) {
    throw new Error(`ops cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    !mobileDockVisible ||
    mobileSidebarVisible ||
    !mobilePoliciesVisible ||
    !mobileNodesVisible ||
    !mobileOpsVisible ||
    mobileDockOverflow > 2 ||
    mobileOverviewOverflow > 2 ||
    mobilePoliciesOverflow > 2 ||
    mobileNodesOverflow > 2 ||
    mobileOpsOverflow > 2
  ) {
    throw new Error(`mobile dashboard layout failed: ${JSON.stringify(state)}`);
  }
  if (
    dashboardNavCount !== viewNames.length ||
    dashboardNavSymbolCount !== viewNames.length ||
    dashboardNavBadgeCount !== viewNames.length ||
    mobileDockCount !== viewNames.length ||
    mobileDockSymbolCount !== viewNames.length ||
    mobileDockBadgeCount !== viewNames.length ||
    populatedNavBadgeCount !== viewNames.length * 2 ||
    dashboardViewCount !== viewNames.length
  ) {
    throw new Error(`dashboard navigation is incomplete: ${JSON.stringify(state)}`);
  }
  const expectedFormDrawerSymbols = ["源", "导", "网", "团", "员", "钥", "策"];
  const overflowingFormDrawerHeader = Object.entries(formDrawerHeaderOverflow).find(([, overflow]) => overflow > 0);
  if (
    formDrawerCount !== 7 ||
    collapsedFormDrawerCount !== 7 ||
    formDrawerSymbolCount !== 7 ||
    expectedFormDrawerSymbols.some((symbol, index) => formDrawerSymbols[index] !== symbol) ||
    overflowingFormDrawerHeader
  ) {
    throw new Error(`form drawers should be collapsed and visually stable by default: ${JSON.stringify(state)}`);
  }
  const expectedPanelSymbols = ["源", "点", "网", "团", "员", "钥", "策", "量", "运"];
  const overflowingPanelTitle = Object.entries(panelTitleOverflow).find(([, overflow]) => overflow > 0);
  if (
    panelCountCount !== 9 ||
    populatedPanelCountCount !== 9 ||
    panelSymbolCount !== 9 ||
    expectedPanelSymbols.some((symbol, index) => panelSymbols[index] !== symbol) ||
    overflowingPanelTitle
  ) {
    throw new Error(`module panel count badges are incomplete: ${JSON.stringify(state)}`);
  }
  const sparseContextView = Object.entries(viewContextChipCounts).find(([, count]) => count < 3);
  const overflowingContextView = Object.entries(viewContextOverflow).find(([, overflow]) => overflow > 0);
  if (sparseContextView || overflowingContextView) {
    throw new Error(`dashboard view context is incomplete: ${JSON.stringify(state)}`);
  }
  const sparseRailView = Object.entries(viewRailButtonCounts).find(([, count]) => count < 1);
  if (sparseRailView || !moduleRailOpensTokenForm) {
    throw new Error(`dashboard module rail is incomplete: ${JSON.stringify(state)}`);
  }
  const expectedRailSymbols = {
    overview: ["概", "闭", "步"],
    access: ["源", "加", "导"],
    nodes: ["点", "网", "建"],
    identity: ["钥", "团", "员", "钥", "团", "员"],
    policies: ["策", "建"],
    traffic: ["量", "时", "日", "出"],
    ops: ["运"],
  };
  const railSymbolMismatch = Object.entries(viewRailButtonCounts).find(([view, count]) => {
    const expected = expectedRailSymbols[view] || [];
    const actual = viewRailSymbols[view] || [];
    return count !== viewRailSymbolCounts[view] || expected.some((symbol, index) => actual[index] !== symbol);
  });
  if (railSymbolMismatch) {
    throw new Error(`dashboard module rail symbols are incomplete: ${JSON.stringify(state)}`);
  }
  const overflowingRailView = Object.entries(viewRailOverflow).find(([, overflow]) => overflow > 0);
  if (overflowingRailView) {
    throw new Error(`dashboard module rail elements overflow: ${JSON.stringify(state)}`);
  }
  const railBadgeMismatch = Object.entries(viewRailButtonCounts).find(
    ([view, count]) => count !== viewRailBadgeCounts[view],
  );
  if (railBadgeMismatch) {
    throw new Error(`dashboard module rail badges are incomplete: ${JSON.stringify(state)}`);
  }
  const headerSymbolMismatch = Object.entries(expectedViewSymbols).find(
    ([view, symbol]) => viewHeaderSymbols[view] !== symbol,
  );
  if (headerSymbolMismatch) {
    throw new Error(`dashboard view header symbol is stale: ${JSON.stringify(state)}`);
  }
  if (
    overviewMetricCount !== 7 ||
    overviewMetricSymbolCount !== 7 ||
    overviewMetricValueCount !== 7 ||
    overviewMetricOverflowCount > 0
  ) {
    throw new Error(`overview metric cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  const expectedOverviewReadinessSymbols = ["源", "点", "网", "身", "策"];
  if (
    overviewReadinessCount !== 5 ||
    overviewReadinessIndexCount !== 5 ||
    overviewReadinessSymbolCount !== 5 ||
    overviewReadinessStateCount !== 5 ||
    overviewReadinessStates.some((state) => !["就绪", "待补"].includes(state)) ||
    expectedOverviewReadinessSymbols.some((symbol, index) => overviewReadinessSymbols[index] !== symbol) ||
    overviewReadinessOverflowCount > 0 ||
    overviewNextStepButtonCount !== 1
  ) {
    throw new Error(`overview readiness board is incomplete: ${JSON.stringify(state)}`);
  }
  if (overviewNextStepCardCount !== 1 || overviewNextStepActionCount !== 1 || overviewNextStepOverflowCount > 0) {
    throw new Error(`overview next step action card is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  const expectedOverviewQuickCardSymbols = ["源", "点", "身", "运"];
  if (
    overviewQuickCardCount !== 4 ||
    overviewQuickCardSymbolCount !== 4 ||
    expectedOverviewQuickCardSymbols.some((symbol, index) => overviewQuickCardSymbols[index] !== symbol) ||
    overviewQuickCardBadgeCount !== 4 ||
    overviewQuickCardOverflowCount > 0 ||
    Object.values(overviewQuickCardBadgeValues).some((value) => !value || value === "--")
  ) {
    throw new Error(`overview quick cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  const overflowingView = Object.entries(viewOverflow).find(([, overflow]) => overflow > 2);
  if (overflowingView) {
    throw new Error(`dashboard view creates page horizontal overflow: ${JSON.stringify(state)}`);
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

  if (
    !state.loginHidden ||
    state.loginVisible ||
    state.appHidden ||
    !state.appVisible ||
    !["已连接", "配置可用"].includes(state.status)
  ) {
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
