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
  const calmOpsStylesheetCount = await page.locator('link[href="/assets/calm-ops.css"]').count();
  await page.waitForSelector("#login-form", { state: "visible", timeout: 15000 });
  const loginButtonSymbolCount = await page.locator('#login-form button[type="submit"] .button-symbol').count();
  const loginButtonOverflowCount = await page.evaluate(() => {
    const button = document.querySelector('#login-form button[type="submit"]');
    if (!button) return 1;
    const buttonBox = button.getBoundingClientRect();
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(button.querySelectorAll(".button-symbol, .button-label"))
      .filter((element) => element.offsetParent !== null)
      .reduce((total, element) => {
        const box = element.getBoundingClientRect();
        return box.width > 0 && box.height > 0 && outside(box, buttonBox) ? total + 1 : total;
      }, 0);
  });
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
  const formDrawerToggleSymbolCount = await page.locator("[data-form-drawer] [data-form-drawer-toggle] .button-symbol").count();
  const formSubmitButtonSymbolCount = await page.locator('[data-form-drawer] button[type="submit"] .button-symbol').count();
  const refreshButtonSymbolCount = await page.locator("#refresh .button-symbol").count();
  const logoutButtonSymbolCount = await page.locator("#logout .button-symbol").count();
  const topbarButtonOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll(".topbar-actions button")).reduce((total, button) => {
      const buttonBox = button.getBoundingClientRect();
      const elements = Array.from(button.querySelectorAll(".button-symbol, .button-label")).filter(
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
  const refreshButtonOverflowCount = await page.evaluate(() => {
    const button = document.querySelector("#refresh");
    if (!button) return 1;
    const parent = button.getBoundingClientRect();
    const outside = (child) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(button.querySelectorAll(".button-symbol, .button-label"))
      .filter((element) => element.offsetParent !== null)
      .reduce((total, element) => {
        const box = element.getBoundingClientRect();
        return total + (box.width > 0 && box.height > 0 && outside(box) ? 1 : 0);
      }, 0);
  });
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
        const elements = Array.from(
          header.querySelectorAll(".drawer-title, .drawer-symbol, h2, .drawer-toggle, .drawer-toggle .button-symbol, .drawer-toggle .button-label"),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const parent = element.closest(".drawer-toggle") && !element.classList.contains("drawer-toggle")
            ? element.closest(".drawer-toggle").getBoundingClientRect()
            : headerBox;
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, parent)) {
            total += 1;
          }
        }
        return total;
      }, 0);
  });
  const visibleFormSubmitOverflow = async () => page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll('[data-form-drawer] button[type="submit"]'))
      .filter((button) => button.offsetParent !== null)
      .reduce((total, button) => {
        const buttonBox = button.getBoundingClientRect();
        const elements = Array.from(button.querySelectorAll(".button-symbol, .button-label")).filter(
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
  const formSubmitOverflow = {};
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
    formSubmitOverflow[view] = await visibleFormSubmitOverflow();
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
  const overviewNextStepActionSymbolCount = await page.locator("#overview-next-step .next-step-action .button-symbol").count();
  const overviewNextStepOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const card = document.querySelector("#overview-next-step [data-overview-next-step-card]");
    if (!card) return 1;
    const cardBox = card.getBoundingClientRect();
    return Array.from(card.querySelectorAll(".next-step-symbol, .next-step-body, .next-step-action, .next-step-action .button-symbol, .next-step-action .button-label"))
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
  const trafficTokenSummaryChipCount = await page.locator("#traffic-tokens [data-traffic-summary-chip]").count();
  const trafficOutboundSummaryChipCount = await page.locator("#traffic-outbounds [data-traffic-summary-chip]").count();
  const trafficSectionCount = await page.locator('[data-dashboard-view="traffic"] [data-traffic-card-section]').count();
  const trafficSectionSymbolCount = await page.locator('[data-dashboard-view="traffic"] [data-traffic-section-symbol]').count();
  const trafficSectionBadgeCount = await page.locator('[data-dashboard-view="traffic"] [data-traffic-section-count]').count();
  const trafficSectionHeaderOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll('[data-dashboard-view="traffic"] .traffic-card-section-heading')).reduce(
      (total, heading) => {
        const headingBox = heading.getBoundingClientRect();
        const elements = Array.from(
          heading.querySelectorAll(".traffic-card-section-symbol, .traffic-card-section-title, .traffic-card-section-count"),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, headingBox)) {
            total += 1;
          }
        }
        return total;
      },
      0,
    );
  });
  const trafficChartPanelCount = await page.locator('[data-dashboard-view="traffic"] [data-traffic-chart-panel]').count();
  const trafficChartSymbolCount = await page.locator('[data-dashboard-view="traffic"] [data-traffic-chart-symbol]').count();
  const trafficChartSummaryChipCount = await page
    .locator('[data-dashboard-view="traffic"] [data-traffic-chart-summary-chip]')
    .count();
  const trafficChartOverflow = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const details = [];
    Array.from(document.querySelectorAll('[data-dashboard-view="traffic"] [data-traffic-chart-panel]')).forEach(
      (panel, panelIndex) => {
        const panelBox = panel.getBoundingClientRect();
        const elements = Array.from(
          panel.querySelectorAll(
            "[data-traffic-chart-symbol], .traffic-chart-title, .traffic-chart-summary, [data-traffic-chart-summary-chip]",
          ),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, panelBox)) {
            details.push({
              panelIndex,
              selector:
                element.getAttribute("data-traffic-chart-symbol") !== null
                  ? "symbol"
                  : element.getAttribute("data-traffic-chart-summary-chip") !== null
                    ? "summary-chip"
                    : element.className || element.tagName,
              text: (element.textContent || "").trim().replace(/\s+/g, " ").slice(0, 80),
              panel: {
                left: Math.round(panelBox.left),
                right: Math.round(panelBox.right),
                top: Math.round(panelBox.top),
                bottom: Math.round(panelBox.bottom),
              },
              box: {
                left: Math.round(box.left),
                right: Math.round(box.right),
                top: Math.round(box.top),
                bottom: Math.round(box.bottom),
              },
            });
          }
        }
      },
    );
    return { count: details.length, details };
  });
  const trafficChartOverflowCount = trafficChartOverflow.count;
  const trafficOutboundEmptyStateCount = await page.locator("#traffic-outbounds [data-empty-state]").count();
  const trafficEmptyStateCount = await page.locator('[data-dashboard-view="traffic"] [data-empty-state]').count();
  const trafficEmptyStateSymbolCount = await page.locator('[data-dashboard-view="traffic"] [data-empty-state-symbol]').count();
  const trafficEmptyStateOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll('[data-dashboard-view="traffic"] [data-empty-state]')).reduce((total, empty) => {
      const parentBox = empty.getBoundingClientRect();
      const elements = Array.from(
        empty.querySelectorAll("[data-empty-state-symbol], .empty-state-copy, .empty-state-title, .empty-state-hint"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, parentBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
  const trafficVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#traffic-tokens .traffic-card, #traffic-outbounds .traffic-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(
          ".traffic-card-heading, .traffic-card-heading > code, .traffic-card-summary, [data-traffic-summary-chip], .traffic-card-meter, .traffic-card-field",
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
  const tokenResultSubscriptionProfileCount = await page.locator("#token-result [data-token-subscription-profile]").count();
  const tokenResultSubscriptionProfiles = await page
    .locator("#token-result [data-token-subscription-profile]")
    .allTextContents();
  const tokenResultSubscriptionOriginCount = await page.locator("#token-result [data-token-subscription-origin]").count();
  const tokenResultSubscriptionOrigins = await page
    .locator("#token-result [data-token-subscription-origin]")
    .allTextContents();
  const tokenResultSubscriptionOpenCount = await page.locator("#token-result a[data-token-subscription-link]").count();
  const tokenResultSubscriptionActionSymbolCount = await page.locator("#token-result .token-subscription-actions .button-symbol").count();
  const tokenResultHeadingSymbolCount = await page.locator("#token-result [data-token-result-symbol]").count();
  const tokenResultHeadingMetaCount = await page.locator("#token-result [data-token-result-meta]").count();
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
      resultCard.querySelectorAll(".token-result-heading, .token-result-symbol, .token-result-copy, .token-result-meta, .token-subscription-item, .token-subscription-heading, .token-subscription-kind, .token-subscription-meta, .token-subscription-profile, .token-subscription-origin, code, button, a"),
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
  const tokenCommandGroupCount = await page.locator("#tokens [data-token-command-group]").count();
  const tokenActionButtonSymbolCount = await page.locator("#tokens .token-card-actions .button-symbol").count();
  const tokenSectionHeadingCount = await page.locator("#tokens [data-token-section-heading]").count();
  const tokenSectionSymbolCount = await page.locator("#tokens [data-token-section-symbol]").count();
  const tokenSectionMetaCount = await page.locator("#tokens .token-card-section-meta").count();
  const tokenSectionOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#tokens [data-token-section-heading]")).reduce((total, heading) => {
      const headingBox = heading.getBoundingClientRect();
      const elements = Array.from(
        heading.querySelectorAll(".token-card-section-symbol, .token-card-section-title, .token-card-section-meta"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, headingBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
  const tokenQuotaMeterCount = await page.locator("#tokens [data-token-quota-meter] .quota-meter").count();
  const tokenSummaryChipCount = await page.locator("#tokens [data-token-summary-chip]").count();
  const tokenSubscriptionItemCount = await page.locator("#tokens .token-subscription-item").count();
  const tokenSubscriptionKindCount = await page.locator("#tokens [data-token-subscription-kind]").count();
  const tokenSubscriptionProfileCount = await page.locator("#tokens [data-token-subscription-profile]").count();
  const tokenSubscriptionOriginCount = await page.locator("#tokens [data-token-subscription-origin]").count();
  const tokenSubscriptionCopyCount = await page.locator("#tokens button[data-token-action='copy-subscription']").count();
  const tokenSubscriptionOpenCount = await page.locator("#tokens a[data-token-subscription-link]").count();
  const tokenSubscriptionActionSymbolCount = await page.locator("#tokens .token-subscription-actions .button-symbol").count();
  const tokenRotateSubscriptionCount = await page.locator("#tokens button[data-token-action='rotate-subscription']").count();
  let tokenCopyFeedbackVisible = false;
  let tokenCopySymbolRestored = false;
  let confirmDialogVisible = false;
  let confirmDialogTitle = "";
  let confirmDialogButtonCount = 0;
  let confirmDialogCancelled = false;
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
  if (tokenRotateSubscriptionCount > 0) {
    await page.locator("#tokens button[data-token-action='rotate-subscription']").first().click();
    const confirmDialog = page.locator("[data-confirm-dialog]").first();
    await expect(confirmDialog).toBeVisible({ timeout: 5000 });
    confirmDialogVisible = true;
    confirmDialogTitle = (await confirmDialog.locator("[data-confirm-title]").textContent())?.trim() || "";
    confirmDialogButtonCount = await confirmDialog.locator("button").count();
    await confirmDialog.locator("[data-confirm-cancel]").click();
    await expect(confirmDialog).toBeHidden({ timeout: 5000 });
    await expect(page.locator("#status")).toContainText("已取消", { timeout: 5000 });
    confirmDialogCancelled = true;
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
          ".token-card-heading, .token-card-summary, [data-token-summary-chip], .token-card-field, .token-card-section-heading, .token-card-section-symbol, .token-card-section-title, .token-card-section-meta, .token-card-meter, .token-card-subscriptions, .token-card-control, .token-subscription-item, .token-subscription-heading, .token-subscription-kind, .token-subscription-meta, .token-subscription-profile, .token-subscription-origin, .token-card-actions",
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
  const tokenActionsScrollOverflowCount = await page.evaluate(() =>
    Array.from(document.querySelectorAll("#tokens .token-card-actions")).reduce((total, element) => {
      if (element.scrollWidth > element.clientWidth + 1) return total + 1;
      return total;
    }, 0),
  );
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
  const identitySummaryChipCount = await page.locator('[data-dashboard-view="identity"] [data-identity-summary-chip]').count();
  const identityActionButtonSymbolCount = await page
    .locator("#teams .identity-card-actions .button-symbol, #users .identity-card-actions .button-symbol")
    .count();
  const identityVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#teams .identity-card, #users .identity-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(
          ".identity-card-heading, .identity-card-summary, [data-identity-summary-chip], .identity-card-badge, .identity-card-field, .identity-card-actions, .identity-card-actions button, .identity-card-actions .button-symbol, .identity-card-actions .button-label",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const parent = element.matches(".identity-card-actions button")
          ? (element.closest(".identity-card-actions") || card).getBoundingClientRect()
          : element.closest(".identity-card-actions") && !element.classList.contains("identity-card-actions")
            ? (element.closest("button") || element.closest(".identity-card-actions") || card).getBoundingClientRect()
            : cardBox;
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, parent)) {
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
  const sourceSummaryChipCount = await page.locator("#sources [data-source-summary-chip]").count();
  const sourceActionButtonSymbolCount = await page.locator("#sources .source-actions .button-symbol").count();
  const sourceVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#sources .source-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(
          ".source-card-heading, .source-card-summary, [data-source-summary-chip], .source-type-badge, .source-card-field, .source-actions, .source-actions button, .source-actions .button-symbol, .source-actions .button-label",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const parent = element.matches(".source-actions button")
          ? (element.closest(".source-actions") || card).getBoundingClientRect()
          : element.closest(".source-actions") && !element.classList.contains("source-actions")
            ? (element.closest("button") || element.closest(".source-actions") || card).getBoundingClientRect()
            : cardBox;
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, parent)) {
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
  const virtualNodeSummaryChipCount = await page.locator("#virtual-nodes [data-virtual-node-summary-chip]").count();
  const virtualNodeActionButtonSymbolCount = await page.locator("#virtual-nodes .virtual-node-actions .button-symbol").count();
  const virtualNodeVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#virtual-nodes .virtual-node-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(
          ".virtual-node-card-heading, .virtual-node-card-summary, [data-virtual-node-summary-chip], .virtual-node-listen-badge, .virtual-node-card-field, .virtual-node-actions, .virtual-node-actions button, .virtual-node-actions .button-symbol, .virtual-node-actions .button-label",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const parent = element.matches(".virtual-node-actions button")
          ? (element.closest(".virtual-node-actions") || card).getBoundingClientRect()
          : element.closest(".virtual-node-actions") && !element.classList.contains("virtual-node-actions")
            ? (element.closest("button") || element.closest(".virtual-node-actions") || card).getBoundingClientRect()
            : cardBox;
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, parent)) {
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
  const policySummaryChipCount = await page.locator("#policies [data-policy-summary-chip]").count();
  const policyActionButtonSymbolCount = await page.locator("#policies .policy-actions .button-symbol").count();
  const policyVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#policies .policy-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(
          ".policy-card-heading, .policy-card-summary, [data-policy-summary-chip], .policy-scope-badge, .policy-card-field, .policy-actions, .policy-actions button, .policy-actions .button-symbol, .policy-actions .button-label",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const parent = element.matches(".policy-actions button")
          ? (element.closest(".policy-actions") || card).getBoundingClientRect()
          : element.closest(".policy-actions") && !element.classList.contains("policy-actions")
            ? (element.closest("button") || element.closest(".policy-actions") || card).getBoundingClientRect()
            : cardBox;
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, parent)) {
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
  const nodeRegionSymbolCount = await page.locator("#nodes .node-region-symbol").count();
  const nodeRegionBadgeCount = await page.locator("#nodes .node-region-badge").count();
  const nodeRegionStatChipCount = await page.locator("#nodes [data-node-region-stat]").count();
  const nodeRegionOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#nodes .node-region-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(".node-region-heading, .node-region-symbol, .node-region-copy, .node-region-name, strong, .node-region-badge, .node-region-stats, [data-node-region-stat]"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const parent = element.classList.contains("node-region-heading") || element.classList.contains("node-region-stats")
          ? cardBox
          : (element.closest(".node-region-heading") || element.closest(".node-region-stats") || card).getBoundingClientRect();
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, parent)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
  let nodeCardCount = 0;
  let nodeCardProtocolSymbolCount = 0;
  let nodeCardChipCount = 0;
  let nodeCardEndpointCount = 0;
  let nodeActionButtonSymbolCount = 0;
  let nodeControlButtonSymbolCount = 0;
  let nodeVisualOverflowCount = 0;
  let nodeRegionSummaryCount = 0;
  let nodeRegionSummaryChipCount = 0;
  let nodeRegionSummaryOverflowCount = 0;
  if (nodeRegionCount > 0) {
    await page.locator("#nodes button[data-node-region-action='open']").first().click();
    await expect(page.locator("#nodes .node-card").first()).toBeVisible({ timeout: 5000 });
    nodeRegionSummaryCount = await page.locator("#nodes [data-node-region-summary]").count();
    nodeRegionSummaryChipCount = await page.locator("#nodes [data-node-region-summary-chip]").count();
    nodeCardCount = await page.locator("#nodes .node-card").count();
    nodeCardProtocolSymbolCount = await page.locator("#nodes .node-card-protocol-symbol").count();
    nodeCardChipCount = await page.locator("#nodes [data-node-card-chip]").count();
    nodeCardEndpointCount = await page.locator("#nodes [data-node-card-endpoint]").count();
    nodeActionButtonSymbolCount = await page.locator("#nodes .node-actions .button-symbol").count();
    nodeControlButtonSymbolCount = await page
      .locator(
        "#nodes button[data-node-region-action='back'] .button-symbol, #nodes button[data-node-filter-action='clear'] .button-symbol",
      )
      .count();
    await page.evaluate(() => {
      const summaryTitle = document.querySelector("#nodes [data-node-region-summary] .node-region-summary-copy strong");
      if (summaryTitle) {
        summaryTitle.textContent = "🇨🇳中国|香港-联通/移动/广港 IEPL 长地区名称用于摘要验收";
      }
      const firstTitle = document.querySelector("#nodes .node-card-title");
      if (firstTitle) {
        firstTitle.textContent = "[gougou-joy] 🇦🇷 38阿根廷-联通/移动(AnyTLS)-超长节点名称用于布局验收";
      }
    });
    nodeRegionSummaryOverflowCount = await page.evaluate(() => {
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      return Array.from(document.querySelectorAll("#nodes [data-node-region-summary]")).reduce((total, summary) => {
        const summaryBox = summary.getBoundingClientRect();
        const elements = Array.from(
          summary.querySelectorAll(".node-region-summary-symbol, .node-region-summary-copy, .node-region-summary-copy strong, .node-region-summary-copy span, .node-region-summary-chips, [data-node-region-summary-chip]"),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const parent = element.classList.contains("node-region-summary-copy") || element.classList.contains("node-region-summary-chips")
            ? summaryBox
            : (element.closest(".node-region-summary-copy") || element.closest(".node-region-summary-chips") || summary).getBoundingClientRect();
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, parent)) {
            total += 1;
          }
        }
        return total;
      }, 0);
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
          card.querySelectorAll(".node-card-main, .node-card-heading, .node-card-protocol-symbol, .node-card-copy, .node-card-title, .node-card-subtitle, .node-card-tags, .node-card-chip-row, [data-node-card-chip], .node-card-endpoint, .node-card-endpoint span, .node-card-endpoint code, .node-actions, .node-actions button, .node-actions .button-symbol, .node-actions .button-label"),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const parent = element.classList.contains("node-card-main") || element.classList.contains("node-actions")
            ? cardBox
            : element.matches(".node-actions button")
              ? (element.closest(".node-actions") || card).getBoundingClientRect()
            : element.closest(".node-actions")
              ? (element.closest("button") || element.closest(".node-actions") || card).getBoundingClientRect()
            : element.classList.contains("node-card-endpoint")
              ? (element.closest(".node-card-main") || card).getBoundingClientRect()
              : (element.closest(".node-card-heading") ||
                  element.closest(".node-card-chip-row") ||
                  element.closest(".node-card-endpoint") ||
                  element.closest(".node-card-main") ||
                  card).getBoundingClientRect();
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
  let nodeDetailDrawerVisible = false;
  let nodeDetailSummaryVisible = false;
  let nodeDetailChipCount = 0;
  let nodeDetailSummaryOverflowCount = 0;
  let nodeDetailCodeOverflowCount = 0;
  let nodeDetailCopyVisible = false;
  let nodeDetailCopyFeedbackVisible = false;
  let nodeEditSaveSymbolCount = 0;
  let nodeEditFieldCount = 0;
  let nodeEditInputNames = [];
  let nodeEditFormOverflowCount = 0;
  if (nodeEditCount > 0) {
    await page.locator("#nodes button[data-node-action='edit']").first().click();
    await expect(page.locator("#nodes form[data-node-edit-form]").first()).toBeVisible({ timeout: 5000 });
    nodeEditFormVisible = true;
    nodeEditSaveSymbolCount = await page.locator("#nodes form[data-node-edit-form] button[type='submit'] .button-symbol").count();
    nodeEditFieldCount = await page.locator("#nodes form[data-node-edit-form] .node-edit-field").count();
    nodeEditInputNames = await page
      .locator("#nodes form[data-node-edit-form] input[name], #nodes form[data-node-edit-form] select[name]")
      .evaluateAll((elements) => elements.map((element) => element.getAttribute("name")).sort());
    nodeEditFormOverflowCount = await page.evaluate(() => {
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      return Array.from(document.querySelectorAll("#nodes form[data-node-edit-form]")).reduce((total, form) => {
        const formBox = form.getBoundingClientRect();
        for (const element of Array.from(form.querySelectorAll("label, input, select, button")).filter((element) => element.offsetParent !== null)) {
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, formBox)) {
            total += 1;
          }
        }
        return total;
      }, 0);
    });
    await page.locator("#nodes button[data-node-action='cancel']").first().click();
    await page.locator("#nodes .node-card-main").first().click();
    await expect(page.locator("#nodes .node-detail-drawer .detail-code").first()).toBeVisible({ timeout: 5000 });
    nodeDetailVisible = true;
    nodeDetailDrawerVisible = await page.locator("#nodes .node-detail-drawer").first().isVisible();
    await expect(page.locator("#nodes .node-detail-summary").first()).toBeVisible({ timeout: 5000 });
    nodeDetailSummaryVisible = true;
    nodeDetailChipCount = await page.locator("#nodes [data-node-detail-chip]").count();
    nodeDetailSummaryOverflowCount = await page.evaluate(() => {
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      return Array.from(document.querySelectorAll("#nodes .node-detail-summary")).reduce((total, summary) => {
        const summaryBox = summary.getBoundingClientRect();
        const elements = Array.from(
          summary.querySelectorAll(".node-detail-summary-title, .node-card-protocol-symbol, strong, .node-detail-summary-title span, .node-detail-summary-chips, [data-node-detail-chip]"),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const parent = element.classList.contains("node-detail-summary-title") || element.classList.contains("node-detail-summary-chips")
            ? summaryBox
            : (element.closest(".node-detail-summary-title") || element.closest(".node-detail-summary-chips") || summary).getBoundingClientRect();
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, parent)) {
            total += 1;
          }
        }
        return total;
      }, 0);
    });
    const nodeDetailCopyButton = page.locator("#nodes button[data-node-action='copy-uri']").first();
    await expect(nodeDetailCopyButton).toBeVisible({ timeout: 5000 });
    nodeDetailCopyVisible = true;
    await nodeDetailCopyButton.click();
    await expect(nodeDetailCopyButton).toHaveText("已复制", { timeout: 5000 });
    await expect(page.locator("#status")).toContainText("节点 URI 已复制", { timeout: 5000 });
    nodeDetailCopyFeedbackVisible = true;
    nodeDetailCodeOverflowCount = await page.evaluate(() =>
      Array.from(document.querySelectorAll("#nodes .node-detail-drawer .detail-code")).filter(
        (element) => element.scrollWidth > element.clientWidth + 2,
      ).length,
    );
  }

  await switchView("ops");
  const opsActionCardCount = await page.locator("#ops-actions .ops-action-card").count();
  const opsActionSymbolCount = await page.locator("#ops-actions .ops-action-symbol").count();
  const opsActionChipCount = await page.locator("#ops-actions [data-ops-action-chip]").count();
  const opsActionButtonSymbolCount = await page.locator("#ops-actions button .button-symbol").count();
  await page.locator("#config-check").click();
  await expect(page.locator("#config-check-result .ops-result-card")).toBeVisible({ timeout: 5000 });
  await page.locator("#delivery-readiness").click();
  await expect(page.locator("#config-check-result .ops-result-card")).toContainText(/交付/, { timeout: 5000 });
  const opsResultVisible = await page.locator("#config-check-result .ops-result-card").isVisible();
  const opsResultFieldCount = await page.locator("#config-check-result .ops-result-field").count();
  const opsResultSymbolCount = await page.locator("#config-check-result .ops-result-symbol").count();
  const opsResultChipCount = await page.locator("#config-check-result [data-ops-result-chip]").count();
  const opsVisualOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#ops-actions .ops-action-card, #config-check-result .ops-result-card")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(
          "button, .ops-action-symbol, [data-ops-action-chip], .ops-result-heading, .ops-result-symbol, [data-ops-result-chip], .ops-result-time, .ops-result-field, code, strong",
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
  state.formDrawerToggleSymbolCount = formDrawerToggleSymbolCount;
  state.formSubmitButtonSymbolCount = formSubmitButtonSymbolCount;
  state.loginButtonSymbolCount = loginButtonSymbolCount;
  state.loginButtonOverflowCount = loginButtonOverflowCount;
  state.refreshButtonSymbolCount = refreshButtonSymbolCount;
  state.refreshButtonOverflowCount = refreshButtonOverflowCount;
  state.logoutButtonSymbolCount = logoutButtonSymbolCount;
  state.topbarButtonOverflowCount = topbarButtonOverflowCount;
  state.formDrawerSymbols = formDrawerSymbols;
  state.formDrawerHeaderOverflow = formDrawerHeaderOverflow;
  state.formSubmitOverflow = formSubmitOverflow;
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
  state.overviewNextStepActionSymbolCount = overviewNextStepActionSymbolCount;
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
  state.identitySummaryChipCount = identitySummaryChipCount;
  state.identityActionButtonSymbolCount = identityActionButtonSymbolCount;
  state.identityVisualOverflowCount = identityVisualOverflowCount;
  state.sourceCardCount = sourceCardCount;
  state.sourceSummaryChipCount = sourceSummaryChipCount;
  state.sourceActionButtonSymbolCount = sourceActionButtonSymbolCount;
  state.sourceVisualOverflowCount = sourceVisualOverflowCount;
  state.sourceEditCount = sourceEditCount;
  state.sourceEditFieldsVisible = sourceEditFieldsVisible;
  state.virtualNodeCardCount = virtualNodeCardCount;
  state.virtualNodeSummaryChipCount = virtualNodeSummaryChipCount;
  state.virtualNodeActionButtonSymbolCount = virtualNodeActionButtonSymbolCount;
  state.virtualNodeVisualOverflowCount = virtualNodeVisualOverflowCount;
  state.virtualNodeEditCount = virtualNodeEditCount;
  state.virtualNodeEditFieldsVisible = virtualNodeEditFieldsVisible;
  state.policyCardCount = policyCardCount;
  state.policySummaryChipCount = policySummaryChipCount;
  state.policyActionButtonSymbolCount = policyActionButtonSymbolCount;
  state.policyVisualOverflowCount = policyVisualOverflowCount;
  state.policyEditCount = policyEditCount;
  state.policyEditFieldsVisible = policyEditFieldsVisible;
  state.nodeSearchVisible = nodeSearchVisible;
  state.nodeFilterValue = nodeFilterValue;
  state.nodeFilteredRegionCount = nodeFilteredRegionCount;
  state.nodeSearchClears = nodeSearchClears;
  state.nodeRegionCount = nodeRegionCount;
  state.nodeRegionSymbolCount = nodeRegionSymbolCount;
  state.nodeRegionBadgeCount = nodeRegionBadgeCount;
  state.nodeRegionStatChipCount = nodeRegionStatChipCount;
  state.nodeRegionOverflowCount = nodeRegionOverflowCount;
  state.nodeRegionSummaryCount = nodeRegionSummaryCount;
  state.nodeRegionSummaryChipCount = nodeRegionSummaryChipCount;
  state.nodeRegionSummaryOverflowCount = nodeRegionSummaryOverflowCount;
  state.nodeCardCount = nodeCardCount;
  state.nodeCardProtocolSymbolCount = nodeCardProtocolSymbolCount;
  state.nodeCardChipCount = nodeCardChipCount;
  state.nodeCardEndpointCount = nodeCardEndpointCount;
  state.nodeActionButtonSymbolCount = nodeActionButtonSymbolCount;
  state.nodeControlButtonSymbolCount = nodeControlButtonSymbolCount;
  state.nodeVisualOverflowCount = nodeVisualOverflowCount;
  state.nodeEditCount = nodeEditCount;
  state.nodeEditFormVisible = nodeEditFormVisible;
  state.nodeDetailVisible = nodeDetailVisible;
  state.nodeDetailDrawerVisible = nodeDetailDrawerVisible;
  state.nodeDetailSummaryVisible = nodeDetailSummaryVisible;
  state.nodeDetailChipCount = nodeDetailChipCount;
  state.nodeDetailSummaryOverflowCount = nodeDetailSummaryOverflowCount;
  state.nodeDetailCopyVisible = nodeDetailCopyVisible;
  state.nodeDetailCopyFeedbackVisible = nodeDetailCopyFeedbackVisible;
  state.nodeDetailCodeOverflowCount = nodeDetailCodeOverflowCount;
  state.nodeEditSaveSymbolCount = nodeEditSaveSymbolCount;
  state.nodeEditFieldCount = nodeEditFieldCount;
  state.nodeEditInputNames = nodeEditInputNames;
  state.nodeEditFormOverflowCount = nodeEditFormOverflowCount;
  state.east8TimeSamples = east8TimeSamples;
  state.trafficTokenRowCount = trafficTokenRowCount;
  state.trafficTokenCardCount = trafficTokenCardCount;
  state.trafficOutboundRowCount = trafficOutboundRowCount;
  state.trafficOutboundCardCount = trafficOutboundCardCount;
  state.trafficTokenSummaryChipCount = trafficTokenSummaryChipCount;
  state.trafficOutboundSummaryChipCount = trafficOutboundSummaryChipCount;
  state.trafficSectionCount = trafficSectionCount;
  state.trafficSectionSymbolCount = trafficSectionSymbolCount;
  state.trafficSectionBadgeCount = trafficSectionBadgeCount;
  state.trafficSectionHeaderOverflowCount = trafficSectionHeaderOverflowCount;
  state.trafficChartPanelCount = trafficChartPanelCount;
  state.trafficChartSymbolCount = trafficChartSymbolCount;
  state.trafficChartSummaryChipCount = trafficChartSummaryChipCount;
  state.trafficChartOverflowCount = trafficChartOverflowCount;
  state.trafficChartOverflowDetails = trafficChartOverflow.details;
  state.trafficOutboundEmptyStateCount = trafficOutboundEmptyStateCount;
  state.trafficEmptyStateCount = trafficEmptyStateCount;
  state.trafficEmptyStateSymbolCount = trafficEmptyStateSymbolCount;
  state.trafficEmptyStateOverflowCount = trafficEmptyStateOverflowCount;
  state.trafficVisualOverflowCount = trafficVisualOverflowCount;
  state.tokenRowCount = tokenRowCount;
  state.tokenCardCount = tokenCardCount;
  state.tokenResultSubscriptionItemCount = tokenResultSubscriptionItemCount;
  state.tokenResultSubscriptionKindCount = tokenResultSubscriptionKindCount;
  state.tokenResultSubscriptionProfileCount = tokenResultSubscriptionProfileCount;
  state.tokenResultSubscriptionProfiles = tokenResultSubscriptionProfiles;
  state.tokenResultSubscriptionOriginCount = tokenResultSubscriptionOriginCount;
  state.tokenResultSubscriptionOrigins = tokenResultSubscriptionOrigins;
  state.tokenResultSubscriptionOpenCount = tokenResultSubscriptionOpenCount;
  state.tokenResultSubscriptionActionSymbolCount = tokenResultSubscriptionActionSymbolCount;
  state.tokenResultHeadingSymbolCount = tokenResultHeadingSymbolCount;
  state.tokenResultHeadingMetaCount = tokenResultHeadingMetaCount;
  state.tokenResultVisualOverflowCount = tokenResultVisualOverflowCount;
  state.quotaMeterCount = quotaMeterCount;
  state.quotaUsageVisible = quotaUsageVisible;
  state.tokenExtendInputCount = tokenExtendInputCount;
  state.tokenQuotaInputCount = tokenQuotaInputCount;
  state.tokenActionFieldCount = tokenActionFieldCount;
  state.tokenCommandGroupCount = tokenCommandGroupCount;
  state.tokenActionButtonSymbolCount = tokenActionButtonSymbolCount;
  state.tokenSectionHeadingCount = tokenSectionHeadingCount;
  state.tokenSectionSymbolCount = tokenSectionSymbolCount;
  state.tokenSectionMetaCount = tokenSectionMetaCount;
  state.tokenSectionOverflowCount = tokenSectionOverflowCount;
  state.tokenQuotaMeterCount = tokenQuotaMeterCount;
  state.tokenSummaryChipCount = tokenSummaryChipCount;
  state.tokenSubscriptionItemCount = tokenSubscriptionItemCount;
  state.tokenSubscriptionKindCount = tokenSubscriptionKindCount;
  state.tokenSubscriptionProfileCount = tokenSubscriptionProfileCount;
  state.tokenSubscriptionOriginCount = tokenSubscriptionOriginCount;
  state.tokenSubscriptionCopyCount = tokenSubscriptionCopyCount;
  state.tokenSubscriptionOpenCount = tokenSubscriptionOpenCount;
  state.tokenSubscriptionActionSymbolCount = tokenSubscriptionActionSymbolCount;
  state.tokenCopyFeedbackVisible = tokenCopyFeedbackVisible;
  state.tokenCopySymbolRestored = tokenCopySymbolRestored;
  state.tokenRotateSubscriptionCount = tokenRotateSubscriptionCount;
  state.confirmDialogVisible = confirmDialogVisible;
  state.confirmDialogTitle = confirmDialogTitle;
  state.confirmDialogButtonCount = confirmDialogButtonCount;
  state.confirmDialogCancelled = confirmDialogCancelled;
  state.tokenVisualOverflowCount = tokenVisualOverflowCount;
  state.tokenActionsScrollOverflowCount = tokenActionsScrollOverflowCount;
  state.tokenVisualOverlapCount = tokenVisualOverlapCount;
  state.opsActionCardCount = opsActionCardCount;
  state.opsActionSymbolCount = opsActionSymbolCount;
  state.opsActionChipCount = opsActionChipCount;
  state.opsActionButtonSymbolCount = opsActionButtonSymbolCount;
  state.opsResultVisible = opsResultVisible;
  state.opsResultFieldCount = opsResultFieldCount;
  state.opsResultSymbolCount = opsResultSymbolCount;
  state.opsResultChipCount = opsResultChipCount;
  state.opsVisualOverflowCount = opsVisualOverflowCount;
  state.calmOpsStylesheetCount = calmOpsStylesheetCount;
  if (tokenRowCount > 0 && (tokenExtendInputCount !== tokenRowCount || tokenQuotaInputCount !== tokenRowCount)) {
    throw new Error(`token custom controls missing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && tokenActionFieldCount !== tokenRowCount * 2) {
    throw new Error(`token action labels missing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && (tokenCommandGroupCount !== tokenRowCount || tokenActionButtonSymbolCount < tokenRowCount * 5)) {
    throw new Error(`token action panels missing symbols or command groups: ${JSON.stringify(state)}`);
  }
  if (
    tokenRowCount > 0 &&
    (tokenSectionHeadingCount !== tokenRowCount * 3 ||
      tokenSectionSymbolCount !== tokenRowCount * 3 ||
      tokenSectionMetaCount !== tokenRowCount * 3 ||
      tokenSectionOverflowCount > 0)
  ) {
    throw new Error(`token section headings are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && tokenQuotaMeterCount !== tokenRowCount) {
    throw new Error(`token quota meters missing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && tokenSummaryChipCount !== tokenRowCount * 4) {
    throw new Error(`token summary chips missing: ${JSON.stringify(state)}`);
  }
  if (
    tokenResultSubscriptionItemCount !== 3 ||
    tokenResultSubscriptionKindCount !== 3 ||
    tokenResultSubscriptionProfileCount !== 3 ||
    tokenResultSubscriptionOriginCount !== 3 ||
    ["URI · 通用", "YAML · Mihomo", "JSON · sing-box"].some(
      (profile, index) => tokenResultSubscriptionProfiles[index] !== profile,
    ) ||
    tokenResultSubscriptionOrigins.some((origin) => !["当前访问域名", "配置域名", "相对地址"].includes(origin)) ||
    tokenResultSubscriptionOpenCount !== 3 ||
    tokenResultSubscriptionActionSymbolCount !== 6 ||
    tokenResultHeadingSymbolCount !== 1 ||
    tokenResultHeadingMetaCount !== 1 ||
    tokenResultVisualOverflowCount > 0
  ) {
    throw new Error(`token subscription result card is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && tokenRotateSubscriptionCount !== tokenRowCount) {
    throw new Error(`token subscription rotate controls missing: ${JSON.stringify(state)}`);
  }
  if (tokenRotateSubscriptionCount > 0 && (!confirmDialogVisible || confirmDialogTitle !== "重置订阅地址" || confirmDialogButtonCount !== 2 || !confirmDialogCancelled)) {
    throw new Error(`custom confirm dialog is missing or not cancellable: ${JSON.stringify(state)}`);
  }
  if (
    tokenSubscriptionCopyCount > 0 &&
    (tokenSubscriptionItemCount !== tokenSubscriptionCopyCount ||
      tokenSubscriptionKindCount !== tokenSubscriptionCopyCount ||
      tokenSubscriptionProfileCount !== tokenSubscriptionCopyCount ||
      tokenSubscriptionOriginCount !== tokenSubscriptionCopyCount ||
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
  if (tokenVisualOverflowCount > 0 || tokenVisualOverlapCount > 0 || tokenActionsScrollOverflowCount > 0) {
    throw new Error(`token controls visually overflow or overlap: ${JSON.stringify(state)}`);
  }
  if (
    (trafficTokenRowCount > 0 && trafficTokenCardCount !== trafficTokenRowCount) ||
    (trafficOutboundRowCount > 0 && trafficOutboundCardCount !== trafficOutboundRowCount) ||
    (trafficTokenCardCount > 0 && trafficTokenSummaryChipCount !== trafficTokenCardCount * 4) ||
    (trafficOutboundCardCount > 0 && trafficOutboundSummaryChipCount !== trafficOutboundCardCount * 4) ||
    trafficSectionCount !== 2 ||
    trafficSectionSymbolCount !== 2 ||
    trafficSectionBadgeCount !== 2 ||
    trafficSectionHeaderOverflowCount > 0 ||
    trafficChartPanelCount !== 2 ||
    trafficChartSymbolCount !== 2 ||
    trafficChartSummaryChipCount !== 6 ||
    trafficChartOverflowCount > 0 ||
    (trafficOutboundRowCount === 0 && trafficOutboundEmptyStateCount !== 1) ||
    trafficEmptyStateCount !== trafficEmptyStateSymbolCount ||
    trafficEmptyStateOverflowCount > 0 ||
    trafficVisualOverflowCount > 0
  ) {
    throw new Error(`traffic cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    (teamEditCount > 0 && teamCardCount !== teamEditCount) ||
    (userEditCount > 0 && userCardCount !== userEditCount) ||
    identitySummaryChipCount !== (teamCardCount + userCardCount) * 4 ||
    identityActionButtonSymbolCount !== teamCardCount + userCardCount ||
    identityVisualOverflowCount > 0
  ) {
    throw new Error(`identity cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    sourceEditCount > 0 &&
    (sourceCardCount !== sourceEditCount ||
      sourceSummaryChipCount !== sourceCardCount * 4 ||
      sourceActionButtonSymbolCount !== sourceCardCount * 3 ||
      sourceVisualOverflowCount > 0)
  ) {
    throw new Error(`source cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    virtualNodeEditCount > 0 &&
    (virtualNodeCardCount !== virtualNodeEditCount ||
      virtualNodeSummaryChipCount !== virtualNodeCardCount * 4 ||
      virtualNodeActionButtonSymbolCount !== virtualNodeCardCount ||
      virtualNodeVisualOverflowCount > 0)
  ) {
    throw new Error(`virtual node cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    policyEditCount > 0 &&
    (policyCardCount !== policyEditCount ||
      policySummaryChipCount !== policyCardCount * 4 ||
      policyActionButtonSymbolCount !== policyCardCount ||
      policyVisualOverflowCount > 0)
  ) {
    throw new Error(`policy cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    nodeCardCount > 0 &&
    (nodeCardProtocolSymbolCount !== nodeCardCount ||
      nodeCardChipCount !== nodeCardCount * 3 ||
      nodeCardEndpointCount !== nodeCardCount ||
      nodeActionButtonSymbolCount !== nodeCardCount * 2 ||
      nodeVisualOverflowCount > 0)
  ) {
    throw new Error(`node cards are incomplete or visually overflow their parent: ${JSON.stringify(state)}`);
  }
  if (
    nodeRegionCount > 0 &&
    (nodeRegionSummaryCount !== 1 || nodeRegionSummaryChipCount < 4 || nodeRegionSummaryOverflowCount > 0)
  ) {
    throw new Error(`node region summary is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (nodeRegionCount > 0 && nodeControlButtonSymbolCount !== 2) {
    throw new Error(`node browser controls are missing action symbols: ${JSON.stringify(state)}`);
  }
  if (
    nodeRegionCount > 0 &&
    (nodeRegionSymbolCount !== nodeRegionCount ||
      nodeRegionBadgeCount !== nodeRegionCount ||
      nodeRegionStatChipCount !== nodeRegionCount * 2 ||
      nodeRegionOverflowCount > 0)
  ) {
    throw new Error(`node region cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (!nodeSearchVisible || nodeFilterValue !== "香港" || nodeFilteredRegionCount < 1 || !nodeSearchClears) {
    throw new Error(`node search filter failed: ${JSON.stringify(state)}`);
  }
  if (nodeDetailCodeOverflowCount > 0) {
    throw new Error(`node detail code blocks overflow horizontally: ${JSON.stringify(state)}`);
  }
  if (
    nodeEditCount > 0 &&
    (nodeEditSaveSymbolCount < 1 ||
      nodeEditFieldCount < 4 ||
      ["display_name", "name_mode", "region", "tags"].some((name) => !nodeEditInputNames.includes(name)) ||
      nodeEditFormOverflowCount > 0)
  ) {
    throw new Error(`node edit form is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (nodeDetailVisible && (!nodeDetailDrawerVisible || !nodeDetailSummaryVisible || nodeDetailChipCount < 4 || nodeDetailSummaryOverflowCount > 0)) {
    throw new Error(`node detail summary is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (nodeDetailVisible && (!nodeDetailCopyVisible || !nodeDetailCopyFeedbackVisible)) {
    throw new Error(`node detail copy feedback missing: ${JSON.stringify(state)}`);
  }
  if (
    opsActionCardCount !== 5 ||
    opsActionSymbolCount !== 5 ||
    opsActionChipCount !== 5 ||
    opsActionButtonSymbolCount !== 5 ||
    !opsResultVisible ||
    opsResultFieldCount < 4 ||
    opsResultSymbolCount !== 1 ||
    opsResultChipCount !== 1 ||
    opsVisualOverflowCount > 0
  ) {
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
  const overflowingFormSubmit = Object.entries(formSubmitOverflow).find(([, overflow]) => overflow > 0);
  if (
    formDrawerCount !== 7 ||
    collapsedFormDrawerCount !== 7 ||
    formDrawerSymbolCount !== 7 ||
    formDrawerToggleSymbolCount !== 7 ||
    formSubmitButtonSymbolCount !== 7 ||
    loginButtonSymbolCount !== 1 ||
    loginButtonOverflowCount > 0 ||
    refreshButtonSymbolCount !== 1 ||
    logoutButtonSymbolCount !== 1 ||
    refreshButtonOverflowCount > 0 ||
    topbarButtonOverflowCount > 0 ||
    expectedFormDrawerSymbols.some((symbol, index) => formDrawerSymbols[index] !== symbol) ||
    overflowingFormDrawerHeader ||
    overflowingFormSubmit
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
  if (
    overviewNextStepCardCount !== 1 ||
    overviewNextStepActionCount !== 1 ||
    overviewNextStepActionSymbolCount !== 1 ||
    overviewNextStepOverflowCount > 0
  ) {
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
    calmOpsStylesheetCount !== 1 ||
    !state.loginHidden ||
    state.loginVisible ||
    state.appHidden ||
    !state.appVisible ||
    !["已连接", "配置可用", "交付闭环可测"].includes(state.status)
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
