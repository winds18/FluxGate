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
  let calmOpsResponseStatus = 0;
  let sourceInlineSaveFeedback = null;
  let sourceInlineSaveRestored = null;
  let logoutButtonPendingFeedback = {};
  let logoutButtonCompletedFeedback = {};
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
    if (url.endsWith("/assets/calm-ops.css")) {
      calmOpsResponseStatus = response.status();
    }
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
  await page.evaluate(() => {
    const originalFetch = window.fetch.bind(window);
    window.__logoutProbeSeen = false;
    window.__releaseLogoutProbe = null;
    window.fetch = (input, init = {}) => {
      const url = typeof input === "string" ? input : input?.url || "";
      const method = String(init?.method || input?.method || "GET").toUpperCase();
      if (method === "POST" && String(url).endsWith("/api/auth/logout")) {
        window.__logoutProbeSeen = true;
        return new Promise((resolve) => {
          window.__releaseLogoutProbe = () => {
            window.fetch = originalFetch;
            resolve(new Response("{}", { status: 200, headers: { "content-type": "application/json" } }));
          };
        });
      }
      return originalFetch(input, init);
    };
  });
  await page.locator("#logout").click();
  await expect.poll(() => page.evaluate(() => Boolean(window.__logoutProbeSeen)), { timeout: 1000 }).toBe(true);
  await page.waitForTimeout(100);
  logoutButtonPendingFeedback = await page.evaluate(() => {
    const button = document.querySelector("#logout");
    const appView = document.querySelector("#app-view");
    const parent = button?.getBoundingClientRect();
    const outside = (child) =>
      parent &&
      (child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1);
    const overflow = parent
      ? Array.from(button.querySelectorAll(".button-symbol, .button-label"))
          .filter((element) => element.offsetParent !== null)
          .reduce((total, element) => {
            const box = element.getBoundingClientRect();
            return total + (box.width > 0 && box.height > 0 && outside(box) ? 1 : 0);
          }, 0)
      : 1;
    return {
      disabled: button?.disabled ? 1 : 0,
      pending: button?.classList.contains("is-pending") ? 1 : 0,
      label: button?.querySelector(".button-label")?.textContent?.trim() || "",
      symbol: button?.querySelector(".button-symbol")?.textContent?.trim() || "",
      status: document.querySelector("#status")?.textContent?.trim() || "",
      tone: document.querySelector("#status")?.dataset.statusTone || "",
      ariaBusy: appView?.getAttribute("aria-busy") || "",
      overflow,
    };
  });
  await page.evaluate(() => window.__releaseLogoutProbe?.());
  await expect(page.locator("#login-form")).toBeVisible({ timeout: 10000 });
  logoutButtonCompletedFeedback = await page.evaluate(() => ({
    loginVisible: Boolean(
      document.querySelector("#login-form") &&
        getComputedStyle(document.querySelector("#login-form")).display !== "none" &&
        document.querySelector("#login-form").getClientRects().length > 0,
    ),
    appHidden: document.querySelector("#app-view")?.hidden ? 1 : 0,
    logoutHidden: document.querySelector("#logout")?.hidden ? 1 : 0,
    logoutDisabled: document.querySelector("#logout")?.disabled ? 1 : 0,
    status: document.querySelector("#status")?.textContent?.trim() || "",
    tone: document.querySelector("#status")?.dataset.statusTone || "",
    ariaBusy: document.querySelector("#app-view")?.getAttribute("aria-busy") || "",
  }));
  await page.fill("#login-username", username);
  await page.fill("#login-password", password);
  await page.click("#login-form button[type=submit]");
  await expect(page.locator("#login-form")).toBeHidden({ timeout: 10000 });
  await expect(page.locator("#app-view")).toBeVisible({ timeout: 10000 });
  await page.waitForFunction(() => {
    const status = document.querySelector("#status")?.textContent?.trim();
    return Boolean(status && !["登录中", "已登录", "刷新中"].includes(status));
  }, null, { timeout: 15000 });
  const reloginLoadStatus = (await page.locator("#status").textContent())?.trim();
  if (reloginLoadStatus !== "已连接") {
    throw new Error(`dashboard did not reload after logout probe: ${reloginLoadStatus}`);
  }

  const viewNames = ["overview", "access", "nodes", "identity", "policies", "traffic", "ops"];
  const primaryMobileViews = ["overview", "access", "nodes", "identity"];
  const overflowMobileViews = ["policies", "traffic", "ops"];
  const switchView = async (view) => {
    const sidebarButton = page.locator(`.dashboard-sidebar [data-view-nav="${view}"]`).first();
    if (await sidebarButton.isVisible()) {
      await sidebarButton.click();
    } else {
      const dockButton = page.locator(`.mobile-dock [data-view-nav="${view}"]`).first();
      if (await dockButton.count()) {
        await dockButton.click();
      } else {
        await page.locator("[data-mobile-more-toggle]").click();
        await expect(page.locator("#mobile-more-menu")).toBeVisible({ timeout: 3000 });
        await page.locator(`#mobile-more-menu [data-view-nav="${view}"]`).first().click();
        await expect(page.locator("#mobile-more-menu")).toBeHidden({ timeout: 3000 });
      }
    }
    await expect(page.locator(`[data-dashboard-view="${view}"]`)).toBeVisible({ timeout: 5000 });
    const activeNavCount = await page.locator(`[data-view-nav="${view}"].is-active`).count();
    if (activeNavCount === 0) {
      throw new Error(`active navigation missing for ${view}`);
    }
  };
  const workbenchStyleEntries = [];
  const collectWorkbenchStyles = async (selectors) => {
    const entries = await page.evaluate((requestedSelectors) => {
      return requestedSelectors.flatMap((selector) =>
        Array.from(document.querySelectorAll(selector)).map((element) => {
          const style = getComputedStyle(element);
          return {
            selector,
            backgroundColor: style.backgroundColor,
            backgroundImage: style.backgroundImage,
            borderColor: style.borderColor,
            borderRadius: style.borderRadius,
            boxShadow: style.boxShadow,
          };
        }),
      );
    }, selectors);
    workbenchStyleEntries.push(...entries);
  };
  const pageHorizontalOverflow = async () =>
    page.evaluate(() => Math.max(0, document.documentElement.scrollWidth - window.innerWidth));
  const dashboardNavCount = await page.locator(".dashboard-sidebar [data-view-nav]").count();
  const dashboardNavSymbolCount = await page.locator(".dashboard-sidebar .nav-symbol").count();
  const dashboardNavBadgeCount = await page.locator(".dashboard-sidebar [data-view-count]").count();
  const mobileDockCount = await page.locator(".mobile-dock > button").count();
  const mobileDockDirectNavCount = await page.locator(".mobile-dock [data-view-nav]").count();
  const mobileDockSymbolCount = await page.locator(".mobile-dock .mobile-dock-symbol").count();
  const mobileDockBadgeCount = await page.locator(".mobile-dock .mobile-dock-badge").count();
  const mobileMoreToggleCount = await page.locator("[data-mobile-more-toggle]").count();
  const mobileMoreNavCount = await page.locator("#mobile-more-menu [data-view-nav]").count();
  const mobileMoreBadgeCount = await page.locator("#mobile-more-menu [data-view-count]").count();
  const workspaceCommandCenterCount = await page.locator("[data-view-command-center]").count();
  const workspaceCommandCenterStyle = await page.evaluate(() => {
    const element = document.querySelector("[data-view-command-center]");
    if (!element) return null;
    const style = getComputedStyle(element);
    return {
      backgroundColor: style.backgroundColor,
      backgroundImage: style.backgroundImage,
      borderColor: style.borderColor,
      borderRadius: style.borderRadius,
      boxShadow: style.boxShadow,
      display: style.display,
      gridTemplateColumns: style.gridTemplateColumns,
    };
  });
  let populatedNavBadgeCount = 0;
  let navBadgeValues = {};
  const dashboardViewCount = await page.locator("[data-dashboard-view]").count();
  const formDrawerCount = await page.locator("[data-form-drawer]").count();
  const collapsedFormDrawerCount = await page.locator("[data-form-drawer].is-collapsed").count();
  const formDrawerSymbolCount = await page.locator("[data-form-drawer] .drawer-symbol").count();
  const formDrawerToggleSymbolCount = await page.locator("[data-form-drawer] [data-form-drawer-toggle] .button-symbol").count();
  const formDrawerCancelCount = await page.locator("[data-form-drawer] [data-form-drawer-cancel]").count();
  const formDrawerCancelSymbolCount = await page.locator("[data-form-drawer] [data-form-drawer-cancel] .button-symbol").count();
  const formDrawerDraftCount = await page.locator("[data-form-drawer] [data-form-draft]").count();
  const formSubmitButtonSymbolCount = await page.locator('[data-form-drawer] button[type="submit"] .button-symbol').count();
  const refreshButtonSymbolCount = await page.locator("#refresh .button-symbol").count();
  const logoutButtonSymbolCount = await page.locator("#logout .button-symbol").count();
  const authenticatedShellLayout = await page.evaluate(() => {
    const topbar = document.querySelector(".topbar");
    const sidebar = document.querySelector(".dashboard-sidebar");
    const workspace = document.querySelector(".workspace");
    const commandCenter = document.querySelector("[data-view-command-center]");
    const topbarStyle = topbar ? getComputedStyle(topbar) : null;
    const sidebarStyle = sidebar ? getComputedStyle(sidebar) : null;
    const topbarBox = topbar?.getBoundingClientRect();
    const sidebarBox = sidebar?.getBoundingClientRect();
    const workspaceBox = workspace?.getBoundingClientRect();
    const commandBox = commandCenter?.getBoundingClientRect();
    return {
      bodyClass: document.body.classList.contains("is-authenticated") ? 1 : 0,
      topbarPosition: topbarStyle?.position || "",
      topbarBrandVisible: topbar?.querySelector(".brand")?.offsetParent !== null ? 1 : 0,
      topbarLeft: Math.round(topbarBox?.left || 0),
      topbarRightGap: Math.round(Math.max(0, window.innerWidth - (topbarBox?.right || 0))),
      topbarHeight: Math.round(topbarBox?.height || 0),
      sidebarPosition: sidebarStyle?.position || "",
      sidebarTop: Math.round(sidebarBox?.top || 0),
      sidebarHeight: Math.round(sidebarBox?.height || 0),
      viewportHeight: window.innerHeight,
      workspaceTop: Math.round(workspaceBox?.top || 0),
      commandCenterTop: Math.round(commandBox?.top || 0),
      overlapsCommandCenter: topbarBox && commandBox && topbarBox.bottom > commandBox.top + 1 ? 1 : 0,
    };
  });
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
          header.querySelectorAll(".drawer-title, .drawer-symbol, h2, [data-form-draft], .drawer-actions, .drawer-toggle, .drawer-toggle .button-symbol, .drawer-toggle .button-label, [data-form-drawer-cancel], [data-form-drawer-cancel] .button-symbol, [data-form-drawer-cancel] .button-label"),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const parentButton = element.closest(".drawer-toggle, [data-form-drawer-cancel]");
          const parent = parentButton && !element.classList.contains("drawer-toggle") && !element.hasAttribute("data-form-drawer-cancel")
            ? parentButton.getBoundingClientRect()
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
  const visibleFormDrawerNarrowCount = async () => page.evaluate(() => {
    const isVisible = (element) => {
      if (!element) return false;
      const style = getComputedStyle(element);
      const box = element.getBoundingClientRect();
      return style.display !== "none" && style.visibility !== "hidden" && box.width > 0 && box.height > 0;
    };
    return Array.from(document.querySelectorAll("[data-dashboard-view]:not([hidden]) .action-grid")).reduce((total, grid) => {
      const gridBox = grid.getBoundingClientRect();
      if (gridBox.width <= 0) return total;
      return total + Array.from(grid.querySelectorAll("[data-form-drawer]"))
        .filter(isVisible)
        .filter((drawer) => drawer.classList.contains("is-collapsed"))
        .reduce((count, drawer) => {
          const drawerBox = drawer.getBoundingClientRect();
          return count + (drawerBox.width / gridBox.width < 0.92 ? 1 : 0);
        }, 0);
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
  const contentPanelCount = await page.locator("[data-content-panel]").count();
  const contentPanelMetaCount = await page.locator("[data-content-panel-meta]").count();
  const contentPanelStyles = await page.evaluate(() =>
    Array.from(document.querySelectorAll("[data-content-panel]")).map((panel) => {
      const style = getComputedStyle(panel);
      const header = panel.querySelector(".panel-header");
      const headerStyle = header ? getComputedStyle(header) : null;
      return {
        panel: panel.getAttribute("data-content-panel") || "",
        backgroundColor: style.backgroundColor,
        backgroundImage: style.backgroundImage,
        borderColor: style.borderColor,
        borderRadius: style.borderRadius,
        boxShadow: style.boxShadow,
        display: style.display,
        headerBackgroundColor: headerStyle?.backgroundColor || "",
      };
    }),
  );
  const visibleContentPanelHeaderOverflow = async () => page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("[data-content-panel] .panel-title"))
      .filter((title) => title.offsetParent !== null)
      .reduce((total, title) => {
        const titleBox = title.getBoundingClientRect();
        const elements = Array.from(title.querySelectorAll(".panel-symbol, h2, .panel-count, .panel-meta"))
          .filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, titleBox)) {
            total += 1;
          }
        }
        return total;
      }, 0);
  });
  const viewOverflow = {};
  const formDrawerHeaderOverflow = {};
  const formSubmitOverflow = {};
  const formDrawerNarrowCount = {};
  const panelTitleOverflow = {};
  const contentPanelHeaderOverflow = {};
  const viewContextChipCounts = {};
  const viewContextOverflow = {};
  const viewRailButtonCounts = {};
  const viewRailLabels = {};
  const viewRailSymbolCounts = {};
  const viewRailSymbols = {};
  const viewRailBadgeCounts = {};
  const viewRailOverflow = {};
  const viewHeaderSymbols = {};
  const workspaceInsightCounts = {};
  const workspaceInsightTitles = {};
  const workspaceInsightActionCounts = {};
  const workspaceInsightSymbolCounts = {};
  const workspaceInsightActionSymbolCounts = {};
  const workspaceInsightOverflow = {};
  const workspaceCommandCenterOverflow = {};
  const workspaceCommandRailScrollOverflow = {};
  const workspacePrimaryActionCounts = {};
  const workspacePrimaryActionLabels = {};
  const workspacePrimaryActionSymbolCounts = {};
  const workspacePrimaryActionOverflow = {};
  const workspacePrimaryActionFocus = {};
  const workspacePrimaryActionHighlight = {};
  const workspacePrimaryActionStatus = {};
  const workspacePrimaryActionStatusTone = {};
  const activeRailTargets = {};
  let formDrawerCancelFeedback = {};
  let formDrawerDraftFeedback = {};
  let formDrawerEscapeFeedback = {};
  let formDrawerSideSheetFeedback = {};
  let formDrawerSubmitFeedback = {};
  let formDrawerSubmitRestored = {};
  let refreshButtonPendingFeedback = {};
  let refreshButtonPendingRestored = {};
  for (const view of viewNames) {
    await switchView(view);
    viewOverflow[view] = await pageHorizontalOverflow();
    workspaceCommandCenterOverflow[view] = await page.evaluate(() => {
      const center = document.querySelector("[data-view-command-center]");
      if (!center) return 1;
      const centerBox = center.getBoundingClientRect();
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      return Array.from(
        center.querySelectorAll(
          ".workspace-command-main, .workspace-command-title, .workspace-title-row, .workspace-view-symbol, #view-title, #view-description, .workspace-insight, .workspace-command-rail, #view-context, #view-rail, .workspace-actions, #view-primary-action, #refresh",
        ),
      )
        .filter((element) => element.offsetParent !== null)
        .reduce((total, element) => {
          const buttonParent = element.closest("#view-primary-action, #refresh");
          const parent = buttonParent && !element.matches("#view-primary-action, #refresh")
            ? buttonParent.getBoundingClientRect()
            : centerBox;
          const box = element.getBoundingClientRect();
          return total + (box.width > 0 && box.height > 0 && outside(box, parent) ? 1 : 0);
        }, 0);
    });
    workspaceCommandRailScrollOverflow[view] = await page.evaluate(() => {
      const rail = document.querySelector(".workspace-command-rail");
      if (!rail) return 1;
      return rail.scrollWidth > rail.clientWidth + 1 ? 1 : 0;
    });
    formDrawerHeaderOverflow[view] = await visibleFormDrawerHeaderOverflow();
    formSubmitOverflow[view] = await visibleFormSubmitOverflow();
    formDrawerNarrowCount[view] = await visibleFormDrawerNarrowCount();
    viewHeaderSymbols[view] = (await page.locator("#view-symbol").textContent())?.trim();
    workspaceInsightCounts[view] = await page.locator("#workspace-insight:not([hidden])").count();
    workspaceInsightTitles[view] = workspaceInsightCounts[view] > 0
      ? (await page.locator("#workspace-insight .workspace-insight-title").first().textContent())?.trim() || ""
      : "";
    workspaceInsightActionCounts[view] = await page.locator("#workspace-insight [data-workspace-insight-action]").count();
    workspaceInsightSymbolCounts[view] = await page.locator("#workspace-insight .workspace-insight-symbol").count();
    workspaceInsightActionSymbolCounts[view] = await page.locator("#workspace-insight .workspace-insight-action .button-symbol").count();
    workspaceInsightOverflow[view] = await page.evaluate(() => {
      const insight = document.querySelector("#workspace-insight:not([hidden])");
      if (!insight) return 1;
      const insightBox = insight.getBoundingClientRect();
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      return Array.from(
        insight.querySelectorAll(
          ".workspace-insight-symbol, .workspace-insight-copy, .workspace-insight-kicker, .workspace-insight-title, .workspace-insight-detail, .workspace-insight-action, .workspace-insight-action .button-symbol, .workspace-insight-action .button-label",
        ),
      )
        .filter((element) => element.offsetParent !== null)
        .reduce((total, element) => {
          const parent = element.closest(".workspace-insight-action") && !element.classList.contains("workspace-insight-action")
            ? element.closest(".workspace-insight-action").getBoundingClientRect()
            : insightBox;
          const box = element.getBoundingClientRect();
          return total + (box.width > 0 && box.height > 0 && outside(box, parent) ? 1 : 0);
        }, 0);
    });
    workspacePrimaryActionCounts[view] = await page.locator("#view-primary-action").count();
    workspacePrimaryActionLabels[view] = workspacePrimaryActionCounts[view] > 0
      ? (await page.locator("#view-primary-action .button-label").first().textContent())?.trim() || ""
      : "";
    workspacePrimaryActionSymbolCounts[view] = await page.locator("#view-primary-action .button-symbol").count();
    workspacePrimaryActionOverflow[view] = await page.evaluate(() => {
      const button = document.querySelector("#view-primary-action");
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
          return total + (box.width > 0 && box.height > 0 && outside(box, buttonBox) ? 1 : 0);
        }, 0);
    });
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
    viewRailLabels[view] = await page
      .locator("#view-rail .view-rail-label")
      .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
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
        const elements = Array.from(title.querySelectorAll(".panel-symbol, h2, .panel-count, .panel-meta"))
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
    contentPanelHeaderOverflow[view] = await visibleContentPanelHeaderOverflow();
  }
  let workspaceInsightActionNavigates = false;
  await switchView("overview");
  const workspaceInsightAction = page.locator("#workspace-insight [data-workspace-insight-action]").first();
  if (await workspaceInsightAction.isVisible()) {
    const targetID = (await workspaceInsightAction.getAttribute("data-workspace-insight-target")) || "";
    await workspaceInsightAction.click();
    if (targetID) {
      await expect(page.locator(`#${targetID}`).first()).toBeVisible({ timeout: 5000 });
    }
    await expect(page.locator("#status")).toContainText("已定位", { timeout: 5000 });
    workspaceInsightActionNavigates = true;
  }
  await switchView("access");
  await page.locator("#view-primary-action").click();
  await expect(page.locator("#source-form:not(.is-collapsed)")).toBeVisible({ timeout: 5000 });
  await expect(page.locator("#source-form.is-target-highlighted")).toHaveCount(1, { timeout: 1000 });
  formDrawerSideSheetFeedback = await page.evaluate(() => {
    const drawer = document.querySelector("#source-form");
    const drawerBox = drawer?.getBoundingClientRect();
    const formBody = drawer?.querySelector(".form-body");
    const style = drawer ? getComputedStyle(drawer) : null;
    return {
      bodyHasOpen: document.body.classList.contains("has-form-drawer-open") ? 1 : 0,
      activeDrawer: document.body.dataset.activeFormDrawer || "",
      role: drawer?.getAttribute("role") || "",
      ariaModal: drawer?.getAttribute("aria-modal") || "",
      toggleExpanded: drawer?.querySelector("[data-form-drawer-toggle]")?.getAttribute("aria-expanded") || "",
      position: style?.position || "",
      zIndex: Number.parseInt(style?.zIndex || "0", 10) || 0,
      width: Math.round(drawerBox?.width || 0),
      height: Math.round(drawerBox?.height || 0),
      viewportWidth: window.innerWidth,
      viewportHeight: window.innerHeight,
      formBodyOverflowY: formBody ? getComputedStyle(formBody).overflowY : "",
      openDrawerCount: document.querySelectorAll("[data-form-drawer]:not(.is-collapsed)").length,
      pageOverflow: Math.max(0, document.documentElement.scrollWidth - window.innerWidth),
    };
  });
  workspacePrimaryActionFocus.access = await page.evaluate(() => document.activeElement?.getAttribute("name") || "");
  workspacePrimaryActionHighlight.access = await page.locator("#source-form.is-target-highlighted").count();
  workspacePrimaryActionStatus.access = (await page.locator("#status").textContent())?.trim() || "";
  workspacePrimaryActionStatusTone.access = await page.locator("#status").getAttribute("data-status-tone");
  activeRailTargets.access = await page
    .locator('#view-rail .view-rail-button.is-current[aria-current="true"]')
    .evaluateAll((elements) => elements.map((element) => element.getAttribute("data-view-rail-target") || ""));
  await page.waitForTimeout(1900);
  await page.locator('#source-form button[type="submit"]').click();
  await expect(page.locator("#source-form.is-target-highlighted")).toHaveCount(1, { timeout: 1000 });
  const requiredFieldFeedback = await page.evaluate(() => ({
    status: document.querySelector("#status")?.textContent?.trim() || "",
    tone: document.querySelector("#status")?.dataset.statusTone || "",
    activeName: document.activeElement?.getAttribute("name") || "",
    highlighted: document.querySelector("#source-form")?.classList.contains("is-target-highlighted") ? 1 : 0,
    inlineCount: document.querySelectorAll("#source-form [data-field-feedback]").length,
    inlineText: document.querySelector("#source-form [data-field-feedback]")?.textContent?.trim() || "",
    ariaInvalid: document.querySelector('#source-form input[name="name"]')?.getAttribute("aria-invalid") || "",
    describedBy: document.querySelector('#source-form input[name="name"]')?.getAttribute("aria-describedby") || "",
  }));
  await page.locator('#source-form input[name="name"]').fill("QA inline source");
  const requiredFieldFeedbackCleared = await page.evaluate(() => ({
    inlineCount: document.querySelectorAll("#source-form [data-field-feedback]").length,
    ariaInvalid: document.querySelector('#source-form input[name="name"]')?.getAttribute("aria-invalid") || "",
    invalidClass: document.querySelector('#source-form input[name="name"]')?.classList.contains("is-field-invalid") ? 1 : 0,
  }));
  formDrawerDraftFeedback = await page.evaluate(() => {
    const drawer = document.querySelector("#source-form");
    const draft = drawer?.querySelector("[data-form-draft]");
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const headerBox = drawer?.querySelector(".panel-header")?.getBoundingClientRect();
    const draftBox = draft?.getBoundingClientRect();
    return {
      dirty: drawer?.classList.contains("is-dirty") ? 1 : 0,
      draftHidden: draft?.hidden ?? true,
      text: draft?.textContent?.trim() || "",
      overflow: headerBox && draftBox && draftBox.width > 0 && draftBox.height > 0 && outside(draftBox, headerBox) ? 1 : 0,
    };
  });
  await page.keyboard.press("Escape");
  await expect(page.locator("#source-form.is-collapsed")).toHaveCount(1, { timeout: 1000 });
  formDrawerEscapeFeedback = await page.evaluate(() => {
    const drawer = document.querySelector("#source-form");
    return {
      bodyHasOpen: document.body.classList.contains("has-form-drawer-open") ? 1 : 0,
      activeDrawer: document.body.dataset.activeFormDrawer || "",
      role: drawer?.getAttribute("role") || "",
      ariaModal: drawer?.getAttribute("aria-modal") || "",
      toggleExpanded: drawer?.querySelector("[data-form-drawer-toggle]")?.getAttribute("aria-expanded") || "",
      dirty: drawer?.classList.contains("is-dirty") ? 1 : 0,
      value: document.querySelector('#source-form input[name="name"]')?.value || "",
      status: document.querySelector("#status")?.textContent?.trim() || "",
      tone: document.querySelector("#status")?.dataset.statusTone || "",
      pageOverflow: Math.max(0, document.documentElement.scrollWidth - window.innerWidth),
    };
  });
  await page.locator('#source-form [data-form-drawer-toggle]').click();
  await expect(page.locator("#source-form:not(.is-collapsed)")).toBeVisible({ timeout: 1000 });
  await page.locator('#source-form [data-form-drawer-cancel]').click();
  await expect(page.locator("#source-form.is-collapsed")).toHaveCount(1, { timeout: 1000 });
  formDrawerCancelFeedback = await page.evaluate(() => ({
    status: document.querySelector("#status")?.textContent?.trim() || "",
    tone: document.querySelector("#status")?.dataset.statusTone || "",
    value: document.querySelector('#source-form input[name="name"]')?.value || "",
    inlineCount: document.querySelectorAll("#source-form [data-field-feedback]").length,
    ariaInvalid: document.querySelector('#source-form input[name="name"]')?.getAttribute("aria-invalid") || "",
    dirty: document.querySelector("#source-form")?.classList.contains("is-dirty") ? 1 : 0,
    draftHidden: document.querySelector("#source-form [data-form-draft]")?.hidden ?? true,
  }));
  await page.evaluate(() => {
    const originalFetch = window.fetch.bind(window);
    window.__sourceSubmitSeen = false;
    window.__releaseSourceSubmit = null;
    window.fetch = (input, init = {}) => {
      const url = typeof input === "string" ? input : input?.url || "";
      const method = String(init?.method || "GET").toUpperCase();
      if (method === "POST" && String(url).endsWith("/api/sources")) {
        window.__sourceSubmitSeen = true;
        window.__releaseSourceSubmit = () => {
          window.fetch = originalFetch;
        };
        return new Promise((resolve) => {
          window.__releaseSourceSubmit = () => {
            window.fetch = originalFetch;
            resolve(
              new Response(
                JSON.stringify({
                  id: 999001,
                  name: "QA pending source",
                  type: "manual",
                  url: "",
                  display_prefix: "[QA pending source]",
                  default_tags: "",
                  refresh_interval_minutes: 0,
                  status: "active",
                }),
                {
                  status: 200,
                  headers: { "content-type": "application/json" },
                },
              ),
            );
          };
        });
      }
      return originalFetch(input, init);
    };
  });
  await page.locator("#view-primary-action").click();
  await expect(page.locator("#source-form:not(.is-collapsed)")).toBeVisible({ timeout: 5000 });
  await page.locator('#source-form input[name="name"]').fill("QA pending source");
  await page.locator('#source-form button[type="submit"]').click();
  await expect.poll(() => page.evaluate(() => Boolean(window.__sourceSubmitSeen)), { timeout: 1000 }).toBe(true);
  await page.waitForTimeout(100);
  formDrawerSubmitFeedback = await page.evaluate(() => {
    const drawer = document.querySelector("#source-form");
    const submitButton = drawer?.querySelector('button[type="submit"]');
    const cancelButton = drawer?.querySelector("[data-form-drawer-cancel]");
    const nameInput = drawer?.querySelector('input[name="name"]');
    const label = submitButton?.querySelector(".button-label")?.textContent?.trim() || "";
    const symbol = submitButton?.querySelector(".button-symbol")?.textContent?.trim() || "";
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const buttonBox = submitButton?.getBoundingClientRect();
    const buttonOverflow = buttonBox
      ? Array.from(submitButton.querySelectorAll(".button-symbol, .button-label")).reduce((total, element) => {
          const box = element.getBoundingClientRect();
          return total + (box.width > 0 && box.height > 0 && outside(box, buttonBox) ? 1 : 0);
        }, 0)
      : 1;
    return {
      submitting: drawer?.classList.contains("is-submitting") ? 1 : 0,
      ariaBusy: drawer?.getAttribute("aria-busy") || "",
      submitDisabled: submitButton?.disabled ? 1 : 0,
      cancelDisabled: cancelButton?.disabled ? 1 : 0,
      inputDisabled: nameInput?.disabled ? 1 : 0,
      label,
      symbol,
      status: document.querySelector("#status")?.textContent?.trim() || "",
      tone: document.querySelector("#status")?.dataset.statusTone || "",
      overflow: buttonOverflow,
    };
  });
  await page.evaluate(() => window.__releaseSourceSubmit?.());
  await expect(page.locator("#status")).toHaveText("已连接", { timeout: 15000 });
  await expect(page.locator('#source-form button[type="submit"]')).toBeEnabled({ timeout: 15000 });
  formDrawerSubmitRestored = await page.evaluate(() => {
    const drawer = document.querySelector("#source-form");
    const submitButton = drawer?.querySelector('button[type="submit"]');
    const cancelButton = drawer?.querySelector("[data-form-drawer-cancel]");
    const nameInput = drawer?.querySelector('input[name="name"]');
    return {
      collapsed: drawer?.classList.contains("is-collapsed") ? 1 : 0,
      submitting: drawer?.classList.contains("is-submitting") ? 1 : 0,
      ariaBusy: drawer?.getAttribute("aria-busy") || "",
      submitDisabled: submitButton?.disabled ? 1 : 0,
      cancelDisabled: cancelButton?.disabled ? 1 : 0,
      inputDisabled: nameInput?.disabled ? 1 : 0,
      label: submitButton?.querySelector(".button-label")?.textContent?.trim() || "",
      symbol: submitButton?.querySelector(".button-symbol")?.textContent?.trim() || "",
    };
  });
  const refreshOverviewFixture = await page.evaluate(async () => {
    const response = await fetch("/api/overview");
    return response.json();
  });
  await page.evaluate((overviewFixture) => {
    const originalFetch = window.fetch.bind(window);
    window.__refreshProbeSeen = false;
    window.__releaseRefreshProbe = null;
    window.fetch = (input, init = {}) => {
      const url = typeof input === "string" ? input : input?.url || "";
      const method = String(init?.method || input?.method || "GET").toUpperCase();
      if (method === "GET" && String(url).endsWith("/api/overview")) {
        window.__refreshProbeSeen = true;
        return new Promise((resolve) => {
          window.__releaseRefreshProbe = () => {
            window.fetch = originalFetch;
            resolve(
              new Response(JSON.stringify(overviewFixture), {
                status: 200,
                headers: { "content-type": "application/json" },
              }),
            );
          };
        });
      }
      return originalFetch(input, init);
    };
  }, refreshOverviewFixture);
  await page.locator("#refresh").click();
  await expect.poll(() => page.evaluate(() => Boolean(window.__refreshProbeSeen)), { timeout: 1000 }).toBe(true);
  await page.waitForTimeout(100);
  refreshButtonPendingFeedback = await page.evaluate(() => {
    const button = document.querySelector("#refresh");
    const appView = document.querySelector("#app-view");
    const parent = button?.getBoundingClientRect();
    const outside = (child) =>
      parent &&
      (child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1);
    const overflow = parent
      ? Array.from(button.querySelectorAll(".button-symbol, .button-label"))
          .filter((element) => element.offsetParent !== null)
          .reduce((total, element) => {
            const box = element.getBoundingClientRect();
            return total + (box.width > 0 && box.height > 0 && outside(box) ? 1 : 0);
          }, 0)
      : 1;
    return {
      disabled: button?.disabled ? 1 : 0,
      pending: button?.classList.contains("is-pending") ? 1 : 0,
      label: button?.querySelector(".button-label")?.textContent?.trim() || "",
      symbol: button?.querySelector(".button-symbol")?.textContent?.trim() || "",
      status: document.querySelector("#status")?.textContent?.trim() || "",
      tone: document.querySelector("#status")?.dataset.statusTone || "",
      ariaBusy: appView?.getAttribute("aria-busy") || "",
      overflow,
    };
  });
  await page.evaluate(() => window.__releaseRefreshProbe?.());
  await expect(page.locator("#status")).toHaveText("已连接", { timeout: 15000 });
  await expect(page.locator("#refresh")).toBeEnabled({ timeout: 15000 });
  refreshButtonPendingRestored = await page.evaluate(() => {
    const button = document.querySelector("#refresh");
    const appView = document.querySelector("#app-view");
    return {
      disabled: button?.disabled ? 1 : 0,
      pending: button?.classList.contains("is-pending") ? 1 : 0,
      label: button?.querySelector(".button-label")?.textContent?.trim() || "",
      symbol: button?.querySelector(".button-symbol")?.textContent?.trim() || "",
      status: document.querySelector("#status")?.textContent?.trim() || "",
      tone: document.querySelector("#status")?.dataset.statusTone || "",
      ariaBusy: appView?.getAttribute("aria-busy") || "",
    };
  });
  await switchView("identity");
  await page.locator("#view-primary-action").click();
  await expect(page.locator("#token-form:not(.is-collapsed)")).toBeVisible({ timeout: 5000 });
  await expect(page.locator("#token-form.is-target-highlighted")).toHaveCount(1, { timeout: 1000 });
  workspacePrimaryActionFocus.identity = await page.evaluate(() => document.activeElement?.getAttribute("name") || "");
  workspacePrimaryActionHighlight.identity = await page.locator("#token-form.is-target-highlighted").count();
  workspacePrimaryActionStatus.identity = (await page.locator("#status").textContent())?.trim() || "";
  workspacePrimaryActionStatusTone.identity = await page.locator("#status").getAttribute("data-status-tone");
  activeRailTargets.identity = await page
    .locator('#view-rail .view-rail-button.is-current[aria-current="true"]')
    .evaluateAll((elements) => elements.map((element) => element.getAttribute("data-view-rail-target") || ""));
  const workspacePrimaryActionOpensTokenForm = true;
  await switchView("overview");
  await expect(page.locator("#metrics .metric")).toHaveCount(6, { timeout: 5000 });
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
  const overviewInsightPanelCount = await page.locator("#overview-insights [data-overview-insight]").count();
  const overviewInsightSymbolCount = await page.locator("#overview-insights .overview-insight-symbol").count();
  const overviewInsightTitleCount = await page.locator("#overview-insights .overview-insight-title").count();
  const overviewHealthPillCount = await page.locator("#overview-insights .overview-health-pill").count();
  const overviewRegionRowCount = await page.locator("#overview-insights .overview-distribution-row").count();
  const overviewSyncRowCount = await page.locator("#overview-insights .overview-sync-row").count();
  const overviewDeliveryRowCount = await page.locator("#overview-insights .overview-delivery-row").count();
  const overviewInsightOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#overview-insights [data-overview-insight]")).reduce((total, panel) => {
      const panelBox = panel.getBoundingClientRect();
      const elements = Array.from(
        panel.querySelectorAll(
          ".overview-insight-heading, .overview-insight-symbol, .overview-insight-title, .overview-insight-meta, .overview-health-pill, .overview-distribution-row, .overview-distribution-label, .overview-distribution-value, .overview-sync-row, .overview-sync-name, .overview-sync-meta, .overview-delivery-row",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, panelBox)) {
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
  const overviewReadinessOverflowDetails = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const details = [];
    Array.from(document.querySelectorAll("#overview-readiness .readiness-item")).forEach((item, itemIndex) => {
      const itemBox = item.getBoundingClientRect();
      const elements = Array.from(
        item.querySelectorAll(
          ".readiness-symbol-wrap, .readiness-index, .readiness-symbol, .readiness-dot, .readiness-body, .readiness-title-row, .readiness-state, strong, small",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, itemBox)) {
          details.push({
            itemIndex,
            selector: element.className || element.tagName,
            text: (element.textContent || "").trim().replace(/\s+/g, " ").slice(0, 80),
            item: {
              left: Math.round(itemBox.left),
              right: Math.round(itemBox.right),
              top: Math.round(itemBox.top),
              bottom: Math.round(itemBox.bottom),
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
    });
    return details;
  });
  const overviewGuideProgressCount = await page.locator("#overview-readiness [data-guide-progress]").count();
  const overviewGuideProgressValue = overviewGuideProgressCount
    ? (await page.locator("#overview-readiness [data-guide-progress]").first().textContent())?.trim() || ""
    : "";
  const overviewGuideCurrentCount = await page.locator("#overview-readiness [data-guide-current]").count();
  const overviewGuideCurrentActionSymbolCount = await page.locator("#overview-readiness [data-guide-current] .button-symbol").count();
  const overviewGuideActionStripCount = await page.locator("#overview-readiness [data-guide-action-strip]").count();
  const overviewGuideActionButtonCount = await page.locator("#overview-readiness [data-guide-action]").count();
  const overviewGuideActionSymbolCount = await page.locator("#overview-readiness .guide-action-symbol").count();
  const overviewGuideActionLabels = await page
    .locator("#overview-readiness [data-guide-action] strong")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const overviewGuideActionTargets = await page
    .locator("#overview-readiness [data-guide-action]")
    .evaluateAll((elements) => elements.map((element) => element.getAttribute("data-overview-jump-target") || ""));
  const overviewGuideHintCount = await page.locator("#overview-readiness .readiness-hint").count();
  const overviewGuideOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const panel = document.querySelector("#overview-readiness");
    if (!panel) return 1;
    const panelBox = panel.getBoundingClientRect();
    return Array.from(
      panel.querySelectorAll(
        "[data-guide-progress], [data-guide-current], [data-guide-action-strip], [data-guide-action], .guide-progress-bar, .guide-current-copy, .guide-current-action, .guide-action-symbol, .guide-action-copy, .readiness-hint",
      ),
    )
      .filter((element) => element.offsetParent !== null)
      .reduce((total, element) => {
        const box = element.getBoundingClientRect();
        return box.width > 0 && box.height > 0 && outside(box, panelBox) ? total + 1 : total;
      }, 0);
  });
  const overviewGuideOverflowDetails = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const panel = document.querySelector("#overview-readiness");
    if (!panel) return [{ selector: "panel", text: "missing" }];
    const panelBox = panel.getBoundingClientRect();
    const details = [];
    Array.from(
      panel.querySelectorAll(
        "[data-guide-progress], [data-guide-current], [data-guide-action-strip], [data-guide-action], .guide-progress-bar, .guide-current-copy, .guide-current-action, .guide-action-symbol, .guide-action-copy, .readiness-hint",
      ),
    )
      .filter((element) => element.offsetParent !== null)
      .forEach((element) => {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, panelBox)) {
          details.push({
            selector: element.className || element.tagName,
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
      });
    return details;
  });
  let overviewGuideTokenActionNavigates = false;
  let overviewGuidePublishActionNavigates = false;
  const overviewHeroCount = await page.locator("#overview-hero").count();
  const overviewHeroTitle = overviewHeroCount
    ? (await page.locator("#overview-hero .overview-hero-title strong").first().textContent())?.trim() || ""
    : "";
  const overviewHeroActionCount = await page.locator("#overview-hero [data-overview-hero-action]").count();
  const overviewHeroActionSymbolCount = await page.locator("#overview-hero [data-overview-hero-action] .button-symbol").count();
  const overviewHeroStatCount = await page.locator("#overview-hero .overview-hero-stat").count();
  const overviewHeroStatSymbolCount = await page.locator("#overview-hero .overview-hero-stat-symbol").count();
  const overviewHeroOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    const hero = document.querySelector("#overview-hero");
    if (!hero) return 1;
    const heroBox = hero.getBoundingClientRect();
    return Array.from(
      hero.querySelectorAll(
        ".overview-hero-title, .overview-hero-title strong, .overview-hero-kicker, .overview-hero-copy, .overview-hero-stat, .overview-hero-stat-symbol, .overview-hero-stat-copy, [data-overview-hero-action]",
      ),
    )
      .filter((element) => element.offsetParent !== null)
      .reduce((total, element) => {
        const box = element.getBoundingClientRect();
        return box.width > 0 && box.height > 0 && outside(box, heroBox) ? total + 1 : total;
      }, 0);
  });
  let overviewHeroActionNavigates = false;
  const heroAction = page.locator("#overview-hero [data-overview-hero-action]").first();
  if ((await heroAction.count()) > 0) {
    await heroAction.click();
    await page.waitForFunction(() => document.querySelector("#status")?.textContent?.includes("已定位："));
    overviewHeroActionNavigates = true;
    await switchView("overview");
  }
  const tokenGuideAction = page.locator('#overview-readiness [data-guide-action][data-overview-jump-target="tokens"]').first();
  if ((await tokenGuideAction.count()) > 0) {
    await tokenGuideAction.click();
    await page.waitForFunction(() => document.querySelector('[data-dashboard-view="identity"]')?.hidden === false);
    await page.waitForFunction(() => document.querySelector("#status")?.textContent?.includes("已定位：Token"));
    overviewGuideTokenActionNavigates = true;
    await switchView("overview");
  }
  const publishGuideAction = page.locator('#overview-readiness [data-guide-action][data-overview-jump-target="config-publish"]').first();
  if ((await publishGuideAction.count()) > 0) {
    await publishGuideAction.click();
    await page.waitForFunction(() => document.querySelector('[data-dashboard-view="ops"]')?.hidden === false);
    await page.waitForFunction(() => document.querySelector("#status")?.textContent?.includes("已定位：发布配置"));
    overviewGuidePublishActionNavigates = true;
    await switchView("overview");
  }
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
  const trafficWorkbenchCount = await page.locator('[data-dashboard-view="traffic"] [data-traffic-workbench]').count();
  await collectWorkbenchStyles(["[data-traffic-workbench]"]);
  const trafficWorkbenchStatCount = await page.locator('[data-dashboard-view="traffic"] [data-traffic-workbench-stat]').count();
  const trafficWorkbenchSignalCardCount = await page
    .locator('[data-dashboard-view="traffic"] [data-traffic-workbench-signal-card]')
    .count();
  const trafficWorkbenchSymbols = await page
    .locator('[data-dashboard-view="traffic"] [data-traffic-workbench-symbol]')
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const trafficWorkbenchOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll('[data-dashboard-view="traffic"] [data-traffic-workbench]')).reduce(
      (total, workbench) => {
        const workbenchBox = workbench.getBoundingClientRect();
        const elements = Array.from(
          workbench.querySelectorAll(
            ".traffic-workbench-heading, .traffic-workbench-copy, .traffic-workbench-title, .traffic-workbench-kicker, .traffic-workbench-stat, .traffic-workbench-stat-symbol, .traffic-workbench-stat-value, .traffic-workbench-stat-detail, .traffic-workbench-signal-card, .traffic-workbench-signal-copy, .traffic-workbench-signal-name, .traffic-workbench-signal-meta, .traffic-workbench-signal-hint",
          ),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const box = element.getBoundingClientRect();
          const parent =
            element.classList.contains("traffic-workbench-heading") ||
            element.classList.contains("traffic-workbench-stat") ||
            element.classList.contains("traffic-workbench-signal-card")
              ? workbenchBox
              : (
                  element.closest(".traffic-workbench-heading") ||
                  element.closest(".traffic-workbench-stat") ||
                  element.closest(".traffic-workbench-signal-card") ||
                  workbench
                ).getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && (outside(box, parent) || outside(box, workbenchBox))) {
            total += 1;
          }
        }
        return total;
      },
      0,
    );
  });
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
  const trafficEmptyStateActionCount = await page.locator('[data-dashboard-view="traffic"] [data-empty-state-action]').count();
  const trafficEmptyStateActionSymbolCount = await page
    .locator('[data-dashboard-view="traffic"] [data-empty-state-action] .button-symbol')
    .count();
  const trafficEmptyStateOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll('[data-dashboard-view="traffic"] [data-empty-state]')).reduce((total, empty) => {
      const parentBox = empty.getBoundingClientRect();
      const elements = Array.from(
        empty.querySelectorAll("[data-empty-state-symbol], .empty-state-copy, .empty-state-title, .empty-state-hint, [data-empty-state-action], [data-empty-state-action] .button-symbol, [data-empty-state-action] .button-label"),
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
  let trafficEmptyStateActionNavigates = false;
  if (trafficOutboundEmptyStateCount > 0) {
    await page.locator("#traffic-outbounds [data-empty-state-action]").first().click();
    await expect(page.locator('[data-dashboard-view="ops"]')).toBeVisible({ timeout: 5000 });
    await expect(page.locator("#ops-actions")).toBeVisible({ timeout: 5000 });
    trafficEmptyStateActionNavigates = true;
    await switchView("traffic");
  }
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
  const tokenResultSubscriptionProbeCount = await page.locator("#token-result button[data-token-action='probe-subscription']").count();
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
      resultCard.querySelectorAll(".token-result-heading, .token-result-symbol, .token-result-copy, .token-result-meta, .token-subscription-item, .token-subscription-heading, .token-subscription-kind, .token-subscription-meta, .token-subscription-profile, .token-subscription-origin, .token-subscription-probe-result, code, button, a"),
    ).reduce((total, element) => {
      if (element.offsetParent === null) return total;
      const box = element.getBoundingClientRect();
      if (box.width > 0 && box.height > 0 && outside(box, resultBox)) {
        return total + 1;
      }
      return total;
    }, 0);
  });
  await page.keyboard.press("Escape");
  await expect(page.locator("#token-form.is-collapsed")).toHaveCount(1, { timeout: 1000 });
  const tokenCardCount = await page.locator("#tokens .token-card").count();
  const tokenRowCount = tokenCardCount || (await page.locator("#tokens tbody tr").count());
  const tokenWorkbenchCount = await page.locator("#tokens [data-token-workbench]").count();
  await collectWorkbenchStyles(["[data-token-workbench]"]);
  const tokenWorkbenchStatCount = await page.locator("#tokens [data-token-workbench-stat]").count();
  const tokenWorkbenchFormatCount = await page.locator("#tokens [data-token-format-card]").count();
  const tokenWorkbenchFormatSymbols = await page
    .locator("#tokens [data-token-format-symbol]")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const tokenWorkbenchActionStripCount = await page.locator("#tokens [data-token-workbench-action-strip]").count();
  const tokenWorkbenchActionCount = await page.locator("#tokens [data-token-workbench-action]").count();
  const tokenWorkbenchActionSymbols = await page
    .locator("#tokens [data-token-workbench-action] .button-symbol")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  let tokenWorkbenchActionNavigates = false;
  if (tokenWorkbenchActionCount > 0) {
    await page.locator("#tokens [data-token-workbench-action='focus-subscriptions']").first().click();
    await expect(page.locator("#tokens .token-card-subscriptions.is-target-highlighted").first()).toHaveCount(1, { timeout: 1000 });
    await expect(page.locator("#status")).toContainText("已定位：订阅地址", { timeout: 5000 });
    tokenWorkbenchActionNavigates = true;
  }
  const tokenWorkbenchOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#tokens [data-token-workbench]")).reduce((total, workbench) => {
      const workbenchBox = workbench.getBoundingClientRect();
      const elements = Array.from(
        workbench.querySelectorAll(
          ".token-workbench-heading, .token-workbench-symbol, .token-workbench-title, .token-workbench-command-strip, [data-token-workbench-action], .token-workbench-stat, .token-workbench-format-card, .token-workbench-format-symbol, .token-workbench-format-title, .token-workbench-format-meta, .token-workbench-format-hint",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, workbenchBox)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
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
  const tokenSubscriptionCopyCount = await page
    .locator("#tokens .token-subscription-item button[data-token-action='copy-subscription']")
    .count();
  const tokenSubscriptionOpenCount = await page.locator("#tokens a[data-token-subscription-link]").count();
  const tokenSubscriptionProbeCount = await page.locator("#tokens button[data-token-action='probe-subscription']").count();
  const tokenSubscriptionActionSymbolCount = await page.locator("#tokens .token-subscription-actions .button-symbol").count();
  const tokenRotateSubscriptionCount = await page.locator("#tokens button[data-token-action='rotate-subscription']").count();
  let tokenCopyFeedbackVisible = false;
  let tokenCopySymbolRestored = false;
  let tokenProbeFeedbackVisible = false;
  let tokenCardActionFeedback = null;
  let tokenCardActionRestored = null;
  let confirmDialogVisible = false;
  let confirmDialogTitle = "";
  let confirmDialogButtonCount = 0;
  let confirmDialogCancelled = false;
  if (tokenSubscriptionCopyCount > 0) {
    const firstCopyButton = page.locator("#tokens .token-subscription-item button[data-token-action='copy-subscription']").first();
    await firstCopyButton.click();
    await expect(firstCopyButton).toHaveText("已复制", { timeout: 5000 });
    await expect(page.locator("#status")).toContainText("订阅地址已复制", { timeout: 5000 });
    tokenCopyFeedbackVisible = true;
    await page.waitForTimeout(1700);
    await expect(firstCopyButton.locator(".button-symbol")).toHaveText("⧉", { timeout: 5000 });
    await expect(firstCopyButton.locator(".button-label")).toHaveText("复制", { timeout: 5000 });
    tokenCopySymbolRestored = true;
  }
  if (tokenSubscriptionProbeCount > 0) {
    const firstProbeButton = page.locator("#tokens button[data-token-action='probe-subscription']").first();
    await firstProbeButton.click();
    await expect(firstProbeButton).toHaveText(/可访问/, { timeout: 5000 });
    await expect(page.locator("#status")).toContainText("订阅可访问", { timeout: 5000 });
    await expect(page.locator("#tokens [data-token-subscription-probed-at]").first()).toBeVisible({ timeout: 5000 });
    tokenProbeFeedbackVisible = true;
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
  if (tokenCardCount > 0) {
    await page.evaluate(() => {
      const originalFetch = window.fetch.bind(window);
      window.__tokenExtendSeen = false;
      window.__releaseTokenExtend = null;
      window.fetch = (input, init = {}) => {
        const url = typeof input === "string" ? input : input?.url || "";
        const method = String(init?.method || "GET").toUpperCase();
        if (method === "POST" && /\/api\/tokens\/\d+\/extend$/.test(String(url))) {
          window.__tokenExtendSeen = true;
          return new Promise((resolve) => {
            window.__releaseTokenExtend = () => {
              window.fetch = originalFetch;
              resolve(
                new Response(JSON.stringify({ ok: true }), {
                  status: 200,
                  headers: { "content-type": "application/json" },
                }),
              );
            };
          });
        }
        return originalFetch(input, init);
      };
    });
    await page.locator("#tokens button[data-token-action='extend']").first().click();
    await expect.poll(() => page.evaluate(() => Boolean(window.__tokenExtendSeen)), { timeout: 1000 }).toBe(true);
    await page.waitForTimeout(100);
    tokenCardActionFeedback = await page.evaluate(() => {
      const card = document.querySelector("#tokens .token-card");
      const extendButton = card?.querySelector("button[data-token-action='extend']");
      const quotaButton = card?.querySelector("button[data-token-action='quota']");
      const revokeButton = card?.querySelector("button[data-token-action='revoke'], button[data-token-action='restore']");
      const extendInput = card?.querySelector("[data-token-extend-days]");
      const quotaInput = card?.querySelector("[data-token-quota-mib]");
      const label = extendButton?.querySelector(".button-label")?.textContent?.trim() || "";
      const symbol = extendButton?.querySelector(".button-symbol")?.textContent?.trim() || "";
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      const buttonBox = extendButton?.getBoundingClientRect();
      const overflow = buttonBox
        ? Array.from(extendButton.querySelectorAll(".button-symbol, .button-label")).reduce((total, element) => {
            const box = element.getBoundingClientRect();
            return total + (box.width > 0 && box.height > 0 && outside(box, buttonBox) ? 1 : 0);
          }, 0)
        : 1;
      return {
        pending: card?.classList.contains("is-action-pending") ? 1 : 0,
        ariaBusy: card?.getAttribute("aria-busy") || "",
        extendDisabled: extendButton?.disabled ? 1 : 0,
        quotaDisabled: quotaButton?.disabled ? 1 : 0,
        revokeDisabled: revokeButton?.disabled ? 1 : 0,
        extendInputDisabled: extendInput?.disabled ? 1 : 0,
        quotaInputDisabled: quotaInput?.disabled ? 1 : 0,
        label,
        symbol,
        status: document.querySelector("#status")?.textContent?.trim() || "",
        tone: document.querySelector("#status")?.dataset.statusTone || "",
        overflow,
      };
    });
    await page.evaluate(() => window.__releaseTokenExtend?.());
    await expect(page.locator("#status")).toHaveText("已连接", { timeout: 15000 });
    await expect(page.locator("#tokens button[data-token-action='extend']").first()).toBeEnabled({ timeout: 15000 });
    tokenCardActionRestored = await page.evaluate(() => {
      const card = document.querySelector("#tokens .token-card");
      const extendButton = card?.querySelector("button[data-token-action='extend']");
      const extendInput = card?.querySelector("[data-token-extend-days]");
      return {
        pending: card?.classList.contains("is-action-pending") ? 1 : 0,
        ariaBusy: card?.getAttribute("aria-busy") || "",
        extendDisabled: extendButton?.disabled ? 1 : 0,
        extendInputDisabled: extendInput?.disabled ? 1 : 0,
        label: extendButton?.querySelector(".button-label")?.textContent?.trim() || "",
        symbol: extendButton?.querySelector(".button-symbol")?.textContent?.trim() || "",
      };
    });
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
          ".token-card-heading, .token-card-summary, [data-token-summary-chip], .token-card-field, .token-card-section-heading, .token-card-section-symbol, .token-card-section-title, .token-card-section-meta, .token-card-meter, .token-card-subscriptions, .token-card-control, .token-subscription-item, .token-subscription-heading, .token-subscription-kind, .token-subscription-meta, .token-subscription-profile, .token-subscription-origin, .token-subscription-probe-result, .token-card-actions",
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

  await switchView("identity");
  const identityWorkbenchCount = await page.locator("#identity-workbench [data-identity-workbench]").count();
  await collectWorkbenchStyles(["[data-identity-workbench]"]);
  const identityWorkbenchStatCount = await page.locator("#identity-workbench [data-identity-workbench-stat]").count();
  const identityWorkbenchStageCardCount = await page.locator("#identity-workbench [data-identity-workbench-stage-card]").count();
  const identityWorkbenchSymbols = await page
    .locator("#identity-workbench [data-identity-workbench-symbol]")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const identityWorkbenchOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#identity-workbench [data-identity-workbench]")).reduce((total, workbench) => {
      const workbenchBox = workbench.getBoundingClientRect();
      const elements = Array.from(
        workbench.querySelectorAll(
          ".identity-workbench-symbol, .identity-workbench-copy, .identity-workbench-kicker, .identity-workbench-title, .identity-workbench-stat, .identity-workbench-stat-symbol, .identity-workbench-stat-copy, .identity-workbench-stat-label, .identity-workbench-stat-value, .identity-workbench-stat-detail, .identity-workbench-stage-card, .identity-workbench-stage-symbol, .identity-workbench-stage-copy, .identity-workbench-stage-name, .identity-workbench-stage-meta, .identity-workbench-stage-hint",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const parent = element.closest(".identity-workbench-stat")
          ? element.closest(".identity-workbench-stat").getBoundingClientRect()
          : element.closest(".identity-workbench-stage-card")
            ? element.closest(".identity-workbench-stage-card").getBoundingClientRect()
            : workbenchBox;
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && (outside(box, parent) || outside(box, workbenchBox))) {
          total += 1;
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
  const sourceWorkbenchCount = await page.locator("#sources [data-source-workbench]").count();
  await collectWorkbenchStyles(["[data-source-workbench]"]);
  const sourceWorkbenchStatCount = await page.locator("#sources [data-source-workbench-stat]").count();
  const sourceWorkbenchTypeCardCount = await page.locator("#sources [data-source-workbench-type-card]").count();
  const sourceWorkbenchSymbols = await page
    .locator("#sources [data-source-workbench-symbol]")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const sourceWorkbenchOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#sources [data-source-workbench]")).reduce((total, workbench) => {
      const workbenchBox = workbench.getBoundingClientRect();
      const elements = Array.from(
        workbench.querySelectorAll(".source-workbench-heading, .source-workbench-copy, .source-workbench-title, .source-workbench-kicker, .source-workbench-stat, .source-workbench-stat-symbol, .source-workbench-stat-value, .source-workbench-stat-detail, .source-workbench-type-card, .source-workbench-type-copy, .source-workbench-type-name, .source-workbench-type-meta, .source-workbench-type-hint"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const parent = element.classList.contains("source-workbench-heading") ||
          element.classList.contains("source-workbench-stat") ||
          element.classList.contains("source-workbench-type-card")
          ? workbenchBox
          : (element.closest(".source-workbench-heading") ||
              element.closest(".source-workbench-stat") ||
              element.closest(".source-workbench-type-card") ||
              workbench).getBoundingClientRect();
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, parent)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
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
    await page.evaluate(() => {
      const originalFetch = window.fetch.bind(window);
      window.__sourceSavePatchSeen = false;
      window.__releaseSourceSavePatch = null;
      window.fetch = (input, init = {}) => {
        const url = typeof input === "string" ? input : input?.url || "";
        const method = String(init?.method || "GET").toUpperCase();
        if (method === "PATCH" && url.includes("/api/sources/")) {
          window.__sourceSavePatchSeen = true;
          const id = Number.parseInt(String(url).split("/").pop() || "0", 10);
          const payload = JSON.parse(init?.body || "{}");
          return new Promise((resolve) => {
            window.__releaseSourceSavePatch = () => {
              window.fetch = originalFetch;
              resolve(
                new Response(
                  JSON.stringify({
                    id,
                    ...payload,
                    last_sync_at: null,
                    last_error: "",
                    status: "active",
                  }),
                  {
                    status: 200,
                    headers: { "content-type": "application/json" },
                  },
                ),
              );
            };
          });
        }
        return originalFetch(input, init);
      };
    });
    await page.locator("#sources button[data-source-action='save']").first().click();
    await expect.poll(() => page.evaluate(() => Boolean(window.__sourceSavePatchSeen)), { timeout: 1000 }).toBe(true);
    await page.waitForTimeout(100);
    sourceInlineSaveFeedback = await page.evaluate(() => {
      const card = document.querySelector("#sources .source-card-editing");
      const saveButton = card?.querySelector("button[data-source-action='save']");
      const cancelButton = card?.querySelector("button[data-source-action='cancel']");
      const nameInput = card?.querySelector('[data-source-field="name"]');
      const typeSelect = card?.querySelector('[data-source-field="type"]');
      const label = saveButton?.querySelector(".button-label")?.textContent?.trim() || "";
      const symbol = saveButton?.querySelector(".button-symbol")?.textContent?.trim() || "";
      const outside = (child, parent) =>
        child.left < parent.left - 1 ||
        child.right > parent.right + 1 ||
        child.top < parent.top - 1 ||
        child.bottom > parent.bottom + 1;
      const buttonBox = saveButton?.getBoundingClientRect();
      const overflow = buttonBox
        ? Array.from(saveButton.querySelectorAll(".button-symbol, .button-label")).reduce((total, element) => {
            const box = element.getBoundingClientRect();
            return total + (box.width > 0 && box.height > 0 && outside(box, buttonBox) ? 1 : 0);
          }, 0)
        : 1;
      return {
        pending: card?.classList.contains("is-action-pending") ? 1 : 0,
        ariaBusy: card?.getAttribute("aria-busy") || "",
        saveDisabled: saveButton?.disabled ? 1 : 0,
        cancelDisabled: cancelButton?.disabled ? 1 : 0,
        inputDisabled: nameInput?.disabled ? 1 : 0,
        selectDisabled: typeSelect?.disabled ? 1 : 0,
        label,
        symbol,
        status: document.querySelector("#status")?.textContent?.trim() || "",
        tone: document.querySelector("#status")?.dataset.statusTone || "",
        overflow,
      };
    });
    await page.evaluate(() => window.__releaseSourceSavePatch?.());
    await expect(page.locator("#status")).toHaveText("已连接", { timeout: 15000 });
    await expect(page.locator("#sources button[data-source-action='edit']").first()).toBeEnabled({ timeout: 15000 });
    sourceInlineSaveRestored = await page.evaluate(() => {
      const card = document.querySelector("#sources .source-card");
      const editButton = card?.querySelector("button[data-source-action='edit']");
      return {
        pending: card?.classList.contains("is-action-pending") ? 1 : 0,
        ariaBusy: card?.getAttribute("aria-busy") || "",
        editDisabled: editButton?.disabled ? 1 : 0,
        saveCount: card?.querySelectorAll("button[data-source-action='save']").length || 0,
        label: editButton?.querySelector(".button-label")?.textContent?.trim() || "",
        symbol: editButton?.querySelector(".button-symbol")?.textContent?.trim() || "",
      };
    });
  }

  await switchView("nodes");
  const virtualNodeCardCount = await page.locator("#virtual-nodes .virtual-node-card").count();
  const virtualNodeWorkbenchCount = await page.locator("#virtual-nodes [data-virtual-node-workbench]").count();
  await collectWorkbenchStyles(["[data-virtual-node-workbench]"]);
  const virtualNodeWorkbenchStatCount = await page.locator("#virtual-nodes [data-virtual-node-workbench-stat]").count();
  const virtualNodeWorkbenchStrategyCardCount = await page
    .locator("#virtual-nodes [data-virtual-node-workbench-strategy-card]")
    .count();
  const virtualNodeWorkbenchSymbols = await page
    .locator("#virtual-nodes [data-virtual-node-workbench-symbol]")
    .evaluateAll((elements) => elements.map((element) => (element.textContent || "").trim()).filter(Boolean));
  const virtualNodeWorkbenchOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#virtual-nodes [data-virtual-node-workbench]")).reduce(
      (total, workbench) => {
        const workbenchBox = workbench.getBoundingClientRect();
        const elements = Array.from(
          workbench.querySelectorAll(
            ".virtual-node-workbench-heading, .virtual-node-workbench-copy, .virtual-node-workbench-title, .virtual-node-workbench-kicker, .virtual-node-workbench-stat, .virtual-node-workbench-stat-symbol, .virtual-node-workbench-stat-value, .virtual-node-workbench-stat-detail, .virtual-node-workbench-strategy-card, .virtual-node-workbench-strategy-copy, .virtual-node-workbench-strategy-name, .virtual-node-workbench-strategy-meta, .virtual-node-workbench-strategy-hint",
          ),
        ).filter((element) => element.offsetParent !== null);
        for (const element of elements) {
          const parent =
            element.classList.contains("virtual-node-workbench-heading") ||
            element.classList.contains("virtual-node-workbench-stat") ||
            element.classList.contains("virtual-node-workbench-strategy-card")
              ? workbenchBox
              : (
                  element.closest(".virtual-node-workbench-heading") ||
                  element.closest(".virtual-node-workbench-stat") ||
                  element.closest(".virtual-node-workbench-strategy-card") ||
                  workbench
                ).getBoundingClientRect();
          const box = element.getBoundingClientRect();
          if (box.width > 0 && box.height > 0 && outside(box, parent)) {
            total += 1;
          }
        }
        return total;
      },
      0,
    );
  });
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
  const policyWorkbenchCount = await page.locator("#policies [data-policy-workbench]").count();
  await collectWorkbenchStyles(["[data-policy-workbench]"]);
  const policyWorkbenchStatCount = await page.locator("#policies [data-policy-workbench-stat]").count();
  const policyWorkbenchScopeCardCount = await page.locator("#policies [data-policy-workbench-scope-card]").count();
  const policyWorkbenchSymbols = await page
    .locator("#policies [data-policy-workbench-symbol]")
    .evaluateAll((elements) => elements.map((element) => (element.textContent || "").trim()).filter(Boolean));
  const policyWorkbenchOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#policies [data-policy-workbench]")).reduce((total, workbench) => {
      const workbenchBox = workbench.getBoundingClientRect();
      const elements = Array.from(
        workbench.querySelectorAll(
          ".policy-workbench-heading, .policy-workbench-copy, .policy-workbench-title, .policy-workbench-kicker, .policy-workbench-stat, .policy-workbench-stat-symbol, .policy-workbench-stat-value, .policy-workbench-stat-detail, .policy-workbench-scope-card, .policy-workbench-scope-copy, .policy-workbench-scope-name, .policy-workbench-scope-meta, .policy-workbench-scope-hint",
        ),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const parent =
          element.classList.contains("policy-workbench-heading") ||
          element.classList.contains("policy-workbench-stat") ||
          element.classList.contains("policy-workbench-scope-card")
            ? workbenchBox
            : (
                element.closest(".policy-workbench-heading") ||
                element.closest(".policy-workbench-stat") ||
                element.closest(".policy-workbench-scope-card") ||
                workbench
              ).getBoundingClientRect();
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, parent)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
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
  const nodeWorkbenchCount = await page.locator("#nodes [data-node-workbench]").count();
  await collectWorkbenchStyles(["[data-node-workbench]"]);
  const nodeWorkbenchStatCount = await page.locator("#nodes [data-node-workbench-stat]").count();
  const nodeWorkbenchRegionCardCount = await page.locator("#nodes [data-node-workbench-region-card]").count();
  const nodeWorkbenchSymbols = await page
    .locator("#nodes [data-node-workbench-symbol]")
    .evaluateAll((elements) => elements.map((element) => element.textContent.trim()));
  const nodeWorkbenchOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("#nodes [data-node-workbench]")).reduce((total, workbench) => {
      const workbenchBox = workbench.getBoundingClientRect();
      const elements = Array.from(
        workbench.querySelectorAll(".node-workbench-heading, .node-workbench-copy, .node-workbench-kicker, .node-workbench-title, .node-workbench-stat, .node-workbench-stat-symbol, .node-workbench-stat-value, .node-workbench-stat-detail, .node-workbench-region-card, .node-workbench-region-heading, .node-workbench-region-copy, .node-workbench-region-name, .node-workbench-region-meta, .node-workbench-region-bar"),
      ).filter((element) => element.offsetParent !== null);
      for (const element of elements) {
        const parent = element.classList.contains("node-workbench-heading") ||
          element.classList.contains("node-workbench-stat") ||
          element.classList.contains("node-workbench-region-card")
          ? workbenchBox
          : (element.closest(".node-workbench-heading") ||
              element.closest(".node-workbench-stat") ||
              element.closest(".node-workbench-region-card") ||
              workbench).getBoundingClientRect();
        const box = element.getBoundingClientRect();
        if (box.width > 0 && box.height > 0 && outside(box, parent)) {
          total += 1;
        }
      }
      return total;
    }, 0);
  });
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
  let nodeDetailOverlapCount = 0;
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
    nodeDetailOverlapCount = await page.evaluate(() => {
      const isVisible = (element) => {
        if (!element) return false;
        const style = getComputedStyle(element);
        const box = element.getBoundingClientRect();
        return style.display !== "none" && style.visibility !== "hidden" && box.width > 0 && box.height > 0;
      };
      const drawer = document.querySelector("#nodes .node-detail-drawer");
      if (!isVisible(drawer)) return 0;
      const drawerBox = drawer.getBoundingClientRect();
      return Array.from(document.querySelectorAll("#nodes .node-card"))
        .filter((card) => isVisible(card))
        .reduce((total, card) => {
          const cardBox = card.getBoundingClientRect();
          const overlapX = Math.max(0, Math.min(drawerBox.right, cardBox.right) - Math.max(drawerBox.left, cardBox.left));
          const overlapY = Math.max(0, Math.min(drawerBox.bottom, cardBox.bottom) - Math.max(drawerBox.top, cardBox.top));
          return total + (overlapX > 1 && overlapY > 1 ? 1 : 0);
        }, 0);
    });
  }

  await switchView("ops");
  const opsWorkbenchCount = await page.locator("[data-ops-workbench]").count();
  await collectWorkbenchStyles(["[data-ops-workbench]"]);
  const opsWorkbenchStatCount = await page.locator("[data-ops-workbench-stat]").count();
  const opsWorkbenchStepCardCount = await page.locator("[data-ops-workbench-step-card]").count();
  const opsWorkbenchSymbols = await page.locator("[data-ops-workbench-symbol]").evaluateAll((nodes) =>
    nodes.map((node) => node.textContent?.trim() || "").filter(Boolean),
  );
  const opsWorkbenchOverflowCount = await page.evaluate(() => {
    const outside = (child, parent) =>
      child.left < parent.left - 1 ||
      child.right > parent.right + 1 ||
      child.top < parent.top - 1 ||
      child.bottom > parent.bottom + 1;
    return Array.from(document.querySelectorAll("[data-ops-workbench], [data-ops-workbench-step-card]")).reduce((total, card) => {
      const cardBox = card.getBoundingClientRect();
      const elements = Array.from(
        card.querySelectorAll(
          "[data-ops-workbench-symbol], [data-ops-workbench-stat], [data-ops-workbench-step-card], .ops-workbench-title, .ops-workbench-copy, .ops-workbench-step-title, .ops-workbench-step-meta, .ops-workbench-next-action, code, strong, small",
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
  const opsActionCardCount = await page.locator("#ops-actions .ops-action-card").count();
  const opsActionSymbolCount = await page.locator("#ops-actions .ops-action-symbol").count();
  const opsActionChipCount = await page.locator("#ops-actions [data-ops-action-chip]").count();
  const opsActionButtonSymbolCount = await page.locator("#ops-actions button .button-symbol").count();
  let opsActionPendingFeedback = null;
  let opsActionPendingRestored = null;
  await page.evaluate(() => {
    window.__opsConfigCheckRelease = null;
    window.__opsConfigCheckSeen = false;
    const originalFetch = window.fetch.bind(window);
    window.fetch = (input, init) => {
      const url = typeof input === "string" ? input : input?.url || "";
      const method = String(init?.method || "GET").toUpperCase();
      if (url.includes("/api/sing-box/config/check") && method === "POST") {
        window.__opsConfigCheckSeen = true;
        return new Promise((resolve) => {
          window.__opsConfigCheckRelease = () => {
            resolve(
              new Response(
                JSON.stringify({
                  valid: true,
                  config_hash: "qa-ops-pending",
                  inbound_count: 1,
                  outbound_count: 2,
                  upstream_outbound_count: 2,
                  user_count: 1,
                }),
                { status: 200, headers: { "Content-Type": "application/json" } },
              ),
            );
          };
        });
      }
      return originalFetch(input, init);
    };
  });
  await page.locator("#config-check").click();
  await expect.poll(async () => page.evaluate(() => window.__opsConfigCheckSeen === true), { timeout: 3000 }).toBe(true);
  opsActionPendingFeedback = await page.evaluate(() => {
    const card = document.querySelector("#config-check")?.closest(".ops-action-card");
    const grid = document.querySelector("#ops-actions");
    const label = document.querySelector("#config-check .button-label")?.textContent?.trim() || "";
    const symbol = document.querySelector("#config-check .button-symbol")?.textContent?.trim() || "";
    const status = document.querySelector("#status")?.textContent?.trim() || "";
    const tone = document.querySelector("#status")?.dataset.statusTone || "";
    const disabled = (selector) => (grid ? Array.from(grid.querySelectorAll(selector)).filter((element) => element.disabled).length : 0);
    const overflow = card
      ? Array.from(card.querySelectorAll("button, .ops-action-symbol, [data-ops-action-chip], strong")).filter((element) => {
          const parent = card.getBoundingClientRect();
          const box = element.getBoundingClientRect();
          return box.width > 0 && box.height > 0 && (box.left < parent.left - 1 || box.right > parent.right + 1);
        }).length
      : 0;
    return {
      pending: document.querySelectorAll("#ops-actions .ops-action-card.is-action-pending").length,
      ariaBusy: card?.getAttribute("aria-busy") || "",
      gridBusy: grid?.getAttribute("aria-busy") || "",
      checkDisabled: disabled("#config-check"),
      publishDisabled: disabled("#config-publish"),
      rollbackDisabled: disabled("#config-rollback"),
      restartDisabled: disabled("#config-restart"),
      label,
      symbol,
      status,
      tone,
      overflow,
    };
  });
  await page.evaluate(() => window.__opsConfigCheckRelease?.());
  await expect(page.locator("#config-check-result .ops-result-card")).toBeVisible({ timeout: 5000 });
  opsActionPendingRestored = await page.evaluate(() => {
    const card = document.querySelector("#config-check")?.closest(".ops-action-card");
    const grid = document.querySelector("#ops-actions");
    const label = document.querySelector("#config-check .button-label")?.textContent?.trim() || "";
    const symbol = document.querySelector("#config-check .button-symbol")?.textContent?.trim() || "";
    const disabled = (selector) => (grid ? Array.from(grid.querySelectorAll(selector)).filter((element) => element.disabled).length : 0);
    return {
      pending: document.querySelectorAll("#ops-actions .ops-action-card.is-action-pending").length,
      ariaBusy: card?.getAttribute("aria-busy") || "",
      gridBusy: grid?.getAttribute("aria-busy") || "",
      checkDisabled: disabled("#config-check"),
      publishDisabled: disabled("#config-publish"),
      rollbackDisabled: disabled("#config-rollback"),
      restartDisabled: disabled("#config-restart"),
      label,
      symbol,
    };
  });
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
  await page.locator("[data-mobile-more-toggle]").click();
  const mobileMoreMenuVisible = await page.locator("#mobile-more-menu").isVisible();
  const mobileMoreMenuOverflow = await page.locator("#mobile-more-menu").evaluate((menu) =>
    Math.max(0, menu.scrollWidth - menu.clientWidth),
  );
  await page.locator('#mobile-more-menu [data-view-nav="policies"]').click();
  await expect(page.locator("#mobile-more-menu")).toBeHidden({ timeout: 3000 });
  const mobileMoreMenuHidesAfterSelection = await page.locator("#mobile-more-menu").isHidden();
  await switchView("policies");
  const mobilePoliciesVisible = await page.locator('[data-dashboard-view="policies"]').isVisible();
  const mobilePoliciesOverflow = await pageHorizontalOverflow();
  await switchView("nodes");
  const mobileNodesVisible = await page.locator('[data-dashboard-view="nodes"]').isVisible();
  const mobileNodesOverflow = await pageHorizontalOverflow();
  await switchView("ops");
  const mobileOpsVisible = await page.locator('[data-dashboard-view="ops"]').isVisible();
  const mobileOpsOverflow = await pageHorizontalOverflow();
  const mobileMoreToggleActiveForOps = await page.locator("[data-mobile-more-toggle]").evaluate((button) =>
    button.classList.contains("is-active") && button.getAttribute("aria-expanded") === "false",
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await switchView("nodes");
  populatedNavBadgeCount = await page
    .locator("[data-view-count]")
    .evaluateAll((elements) => elements.filter((element) => element.textContent.trim().length > 0 && element.textContent.trim() !== "--").length);
  navBadgeValues = await page
    .locator(".dashboard-sidebar [data-view-count]")
    .evaluateAll((elements) => Object.fromEntries(elements.map((element) => [element.dataset.viewCount, element.textContent.trim()])));
  const unifiedWorkbenchStyles = {
    count: new Set(workbenchStyleEntries.map((entry) => entry.selector)).size,
    entries: workbenchStyleEntries,
    mismatches: workbenchStyleEntries.filter(
      (entry) =>
        entry.backgroundColor !== "rgb(255, 255, 255)" ||
        entry.backgroundImage !== "none" ||
        !entry.borderRadius.startsWith("8px") ||
        entry.borderColor === "rgba(0, 0, 0, 0)" ||
        entry.boxShadow === "none",
    ),
  };

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
  state.mobileDockDirectNavCount = mobileDockDirectNavCount;
  state.mobileDockSymbolCount = mobileDockSymbolCount;
  state.mobileDockBadgeCount = mobileDockBadgeCount;
  state.mobileMoreToggleCount = mobileMoreToggleCount;
  state.mobileMoreNavCount = mobileMoreNavCount;
  state.mobileMoreBadgeCount = mobileMoreBadgeCount;
  state.workspaceCommandCenterCount = workspaceCommandCenterCount;
  state.workspaceCommandCenterStyle = workspaceCommandCenterStyle;
  state.workspaceCommandCenterOverflow = workspaceCommandCenterOverflow;
  state.workspaceCommandRailScrollOverflow = workspaceCommandRailScrollOverflow;
  state.authenticatedShellLayout = authenticatedShellLayout;
  state.populatedNavBadgeCount = populatedNavBadgeCount;
  state.navBadgeValues = navBadgeValues;
  state.dashboardViewCount = dashboardViewCount;
  state.formDrawerCount = formDrawerCount;
  state.collapsedFormDrawerCount = collapsedFormDrawerCount;
  state.formDrawerSymbolCount = formDrawerSymbolCount;
  state.formDrawerToggleSymbolCount = formDrawerToggleSymbolCount;
  state.formDrawerCancelCount = formDrawerCancelCount;
  state.formDrawerCancelSymbolCount = formDrawerCancelSymbolCount;
  state.formDrawerDraftCount = formDrawerDraftCount;
  state.formDrawerCancelFeedback = formDrawerCancelFeedback;
  state.formDrawerDraftFeedback = formDrawerDraftFeedback;
  state.formDrawerEscapeFeedback = formDrawerEscapeFeedback;
  state.formDrawerSideSheetFeedback = formDrawerSideSheetFeedback;
  state.formDrawerSubmitFeedback = formDrawerSubmitFeedback;
  state.formDrawerSubmitRestored = formDrawerSubmitRestored;
  state.formSubmitButtonSymbolCount = formSubmitButtonSymbolCount;
  state.loginButtonSymbolCount = loginButtonSymbolCount;
  state.loginButtonOverflowCount = loginButtonOverflowCount;
  state.logoutButtonPendingFeedback = logoutButtonPendingFeedback;
  state.logoutButtonCompletedFeedback = logoutButtonCompletedFeedback;
  state.refreshButtonSymbolCount = refreshButtonSymbolCount;
  state.refreshButtonOverflowCount = refreshButtonOverflowCount;
  state.refreshButtonPendingFeedback = refreshButtonPendingFeedback;
  state.refreshButtonPendingRestored = refreshButtonPendingRestored;
  state.logoutButtonSymbolCount = logoutButtonSymbolCount;
  state.topbarButtonOverflowCount = topbarButtonOverflowCount;
  state.formDrawerSymbols = formDrawerSymbols;
  state.formDrawerHeaderOverflow = formDrawerHeaderOverflow;
  state.formSubmitOverflow = formSubmitOverflow;
  state.formDrawerNarrowCount = formDrawerNarrowCount;
  state.panelCountCount = panelCountCount;
  state.populatedPanelCountCount = populatedPanelCountCount;
  state.panelSymbolCount = panelSymbolCount;
  state.panelSymbols = panelSymbols;
  state.panelTitleOverflow = panelTitleOverflow;
  state.contentPanelCount = contentPanelCount;
  state.contentPanelMetaCount = contentPanelMetaCount;
  state.contentPanelStyles = contentPanelStyles;
  state.contentPanelHeaderOverflow = contentPanelHeaderOverflow;
  state.viewContextChipCounts = viewContextChipCounts;
  state.viewContextOverflow = viewContextOverflow;
  state.viewRailButtonCounts = viewRailButtonCounts;
  state.viewRailLabels = viewRailLabels;
  state.viewRailSymbolCounts = viewRailSymbolCounts;
  state.viewRailSymbols = viewRailSymbols;
  state.viewRailBadgeCounts = viewRailBadgeCounts;
  state.viewRailOverflow = viewRailOverflow;
  state.viewHeaderSymbols = viewHeaderSymbols;
  state.workspaceInsightCounts = workspaceInsightCounts;
  state.workspaceInsightTitles = workspaceInsightTitles;
  state.workspaceInsightActionCounts = workspaceInsightActionCounts;
  state.workspaceInsightSymbolCounts = workspaceInsightSymbolCounts;
  state.workspaceInsightActionSymbolCounts = workspaceInsightActionSymbolCounts;
  state.workspaceInsightOverflow = workspaceInsightOverflow;
  state.workspaceInsightActionNavigates = workspaceInsightActionNavigates;
  state.workspacePrimaryActionCounts = workspacePrimaryActionCounts;
  state.workspacePrimaryActionLabels = workspacePrimaryActionLabels;
  state.workspacePrimaryActionSymbolCounts = workspacePrimaryActionSymbolCounts;
  state.workspacePrimaryActionOverflow = workspacePrimaryActionOverflow;
  state.workspacePrimaryActionFocus = workspacePrimaryActionFocus;
  state.workspacePrimaryActionHighlight = workspacePrimaryActionHighlight;
  state.workspacePrimaryActionStatus = workspacePrimaryActionStatus;
  state.workspacePrimaryActionStatusTone = workspacePrimaryActionStatusTone;
  state.activeRailTargets = activeRailTargets;
  state.workspacePrimaryActionOpensTokenForm = workspacePrimaryActionOpensTokenForm;
  state.requiredFieldFeedback = requiredFieldFeedback;
  state.requiredFieldFeedbackCleared = requiredFieldFeedbackCleared;
  state.moduleRailOpensTokenForm = moduleRailOpensTokenForm;
  state.overviewMetricCount = overviewMetricCount;
  state.overviewMetricSymbolCount = overviewMetricSymbolCount;
  state.overviewMetricValueCount = overviewMetricValueCount;
  state.overviewMetricOverflowCount = overviewMetricOverflowCount;
  state.overviewInsightPanelCount = overviewInsightPanelCount;
  state.overviewInsightSymbolCount = overviewInsightSymbolCount;
  state.overviewInsightTitleCount = overviewInsightTitleCount;
  state.overviewHealthPillCount = overviewHealthPillCount;
  state.overviewRegionRowCount = overviewRegionRowCount;
  state.overviewSyncRowCount = overviewSyncRowCount;
  state.overviewDeliveryRowCount = overviewDeliveryRowCount;
  state.overviewInsightOverflowCount = overviewInsightOverflowCount;
  state.overviewHeroCount = overviewHeroCount;
  state.overviewHeroTitle = overviewHeroTitle;
  state.overviewHeroActionCount = overviewHeroActionCount;
  state.overviewHeroActionSymbolCount = overviewHeroActionSymbolCount;
  state.overviewHeroStatCount = overviewHeroStatCount;
  state.overviewHeroStatSymbolCount = overviewHeroStatSymbolCount;
  state.overviewHeroOverflowCount = overviewHeroOverflowCount;
  state.overviewHeroActionNavigates = overviewHeroActionNavigates;
  state.overviewReadinessCount = overviewReadinessCount;
  state.overviewReadinessIndexCount = overviewReadinessIndexCount;
  state.overviewReadinessSymbolCount = overviewReadinessSymbolCount;
  state.overviewReadinessStateCount = overviewReadinessStateCount;
  state.overviewReadinessStates = overviewReadinessStates;
  state.overviewReadinessSymbols = overviewReadinessSymbols;
  state.overviewReadinessOverflowCount = overviewReadinessOverflowCount;
  state.overviewReadinessOverflowDetails = overviewReadinessOverflowDetails;
  state.overviewGuideProgressCount = overviewGuideProgressCount;
  state.overviewGuideProgressValue = overviewGuideProgressValue;
  state.overviewGuideCurrentCount = overviewGuideCurrentCount;
  state.overviewGuideCurrentActionSymbolCount = overviewGuideCurrentActionSymbolCount;
  state.overviewGuideActionStripCount = overviewGuideActionStripCount;
  state.overviewGuideActionButtonCount = overviewGuideActionButtonCount;
  state.overviewGuideActionSymbolCount = overviewGuideActionSymbolCount;
  state.overviewGuideActionLabels = overviewGuideActionLabels;
  state.overviewGuideActionTargets = overviewGuideActionTargets;
  state.overviewGuideTokenActionNavigates = overviewGuideTokenActionNavigates;
  state.overviewGuidePublishActionNavigates = overviewGuidePublishActionNavigates;
  state.overviewGuideHintCount = overviewGuideHintCount;
  state.overviewGuideOverflowCount = overviewGuideOverflowCount;
  state.overviewGuideOverflowDetails = overviewGuideOverflowDetails;
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
  state.mobileMoreMenuVisible = mobileMoreMenuVisible;
  state.mobileMoreMenuOverflow = mobileMoreMenuOverflow;
  state.mobileMoreMenuHidesAfterSelection = mobileMoreMenuHidesAfterSelection;
  state.mobilePoliciesVisible = mobilePoliciesVisible;
  state.mobilePoliciesOverflow = mobilePoliciesOverflow;
  state.mobileNodesVisible = mobileNodesVisible;
  state.mobileNodesOverflow = mobileNodesOverflow;
  state.mobileOpsVisible = mobileOpsVisible;
  state.mobileOpsOverflow = mobileOpsOverflow;
  state.mobileMoreToggleActiveForOps = mobileMoreToggleActiveForOps;
  state.teamCardCount = teamCardCount;
  state.identityWorkbenchCount = identityWorkbenchCount;
  state.identityWorkbenchStatCount = identityWorkbenchStatCount;
  state.identityWorkbenchStageCardCount = identityWorkbenchStageCardCount;
  state.identityWorkbenchSymbols = identityWorkbenchSymbols;
  state.identityWorkbenchOverflowCount = identityWorkbenchOverflowCount;
  state.teamEditCount = teamEditCount;
  state.teamEditFieldsVisible = teamEditFieldsVisible;
  state.userCardCount = userCardCount;
  state.userEditCount = userEditCount;
  state.userEditFieldsVisible = userEditFieldsVisible;
  state.identitySummaryChipCount = identitySummaryChipCount;
  state.identityActionButtonSymbolCount = identityActionButtonSymbolCount;
  state.identityVisualOverflowCount = identityVisualOverflowCount;
  state.sourceCardCount = sourceCardCount;
  state.sourceWorkbenchCount = sourceWorkbenchCount;
  state.sourceWorkbenchStatCount = sourceWorkbenchStatCount;
  state.sourceWorkbenchTypeCardCount = sourceWorkbenchTypeCardCount;
  state.sourceWorkbenchSymbols = sourceWorkbenchSymbols;
  state.sourceWorkbenchOverflowCount = sourceWorkbenchOverflowCount;
  state.sourceSummaryChipCount = sourceSummaryChipCount;
  state.sourceActionButtonSymbolCount = sourceActionButtonSymbolCount;
  state.sourceVisualOverflowCount = sourceVisualOverflowCount;
  state.sourceEditCount = sourceEditCount;
  state.sourceEditFieldsVisible = sourceEditFieldsVisible;
  state.sourceInlineSaveFeedback = sourceInlineSaveFeedback;
  state.sourceInlineSaveRestored = sourceInlineSaveRestored;
  state.virtualNodeCardCount = virtualNodeCardCount;
  state.virtualNodeWorkbenchCount = virtualNodeWorkbenchCount;
  state.virtualNodeWorkbenchStatCount = virtualNodeWorkbenchStatCount;
  state.virtualNodeWorkbenchStrategyCardCount = virtualNodeWorkbenchStrategyCardCount;
  state.virtualNodeWorkbenchSymbols = virtualNodeWorkbenchSymbols;
  state.virtualNodeWorkbenchOverflowCount = virtualNodeWorkbenchOverflowCount;
  state.virtualNodeSummaryChipCount = virtualNodeSummaryChipCount;
  state.virtualNodeActionButtonSymbolCount = virtualNodeActionButtonSymbolCount;
  state.virtualNodeVisualOverflowCount = virtualNodeVisualOverflowCount;
  state.virtualNodeEditCount = virtualNodeEditCount;
  state.virtualNodeEditFieldsVisible = virtualNodeEditFieldsVisible;
  state.policyCardCount = policyCardCount;
  state.policyWorkbenchCount = policyWorkbenchCount;
  state.policyWorkbenchStatCount = policyWorkbenchStatCount;
  state.policyWorkbenchScopeCardCount = policyWorkbenchScopeCardCount;
  state.policyWorkbenchSymbols = policyWorkbenchSymbols;
  state.policyWorkbenchOverflowCount = policyWorkbenchOverflowCount;
  state.policySummaryChipCount = policySummaryChipCount;
  state.policyActionButtonSymbolCount = policyActionButtonSymbolCount;
  state.policyVisualOverflowCount = policyVisualOverflowCount;
  state.policyEditCount = policyEditCount;
  state.policyEditFieldsVisible = policyEditFieldsVisible;
  state.nodeSearchVisible = nodeSearchVisible;
  state.nodeFilterValue = nodeFilterValue;
  state.nodeFilteredRegionCount = nodeFilteredRegionCount;
  state.nodeSearchClears = nodeSearchClears;
  state.nodeWorkbenchCount = nodeWorkbenchCount;
  state.nodeWorkbenchStatCount = nodeWorkbenchStatCount;
  state.nodeWorkbenchRegionCardCount = nodeWorkbenchRegionCardCount;
  state.nodeWorkbenchSymbols = nodeWorkbenchSymbols;
  state.nodeWorkbenchOverflowCount = nodeWorkbenchOverflowCount;
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
  state.nodeDetailOverlapCount = nodeDetailOverlapCount;
  state.nodeEditSaveSymbolCount = nodeEditSaveSymbolCount;
  state.nodeEditFieldCount = nodeEditFieldCount;
  state.nodeEditInputNames = nodeEditInputNames;
  state.nodeEditFormOverflowCount = nodeEditFormOverflowCount;
  state.east8TimeSamples = east8TimeSamples;
  state.trafficWorkbenchCount = trafficWorkbenchCount;
  state.trafficWorkbenchStatCount = trafficWorkbenchStatCount;
  state.trafficWorkbenchSignalCardCount = trafficWorkbenchSignalCardCount;
  state.trafficWorkbenchSymbols = trafficWorkbenchSymbols;
  state.trafficWorkbenchOverflowCount = trafficWorkbenchOverflowCount;
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
  state.trafficEmptyStateActionCount = trafficEmptyStateActionCount;
  state.trafficEmptyStateActionSymbolCount = trafficEmptyStateActionSymbolCount;
  state.trafficEmptyStateActionNavigates = trafficEmptyStateActionNavigates;
  state.trafficEmptyStateOverflowCount = trafficEmptyStateOverflowCount;
  state.trafficVisualOverflowCount = trafficVisualOverflowCount;
  state.tokenRowCount = tokenRowCount;
  state.tokenCardCount = tokenCardCount;
  state.tokenWorkbenchCount = tokenWorkbenchCount;
  state.tokenWorkbenchStatCount = tokenWorkbenchStatCount;
  state.tokenWorkbenchFormatCount = tokenWorkbenchFormatCount;
  state.tokenWorkbenchFormatSymbols = tokenWorkbenchFormatSymbols;
  state.tokenWorkbenchActionStripCount = tokenWorkbenchActionStripCount;
  state.tokenWorkbenchActionCount = tokenWorkbenchActionCount;
  state.tokenWorkbenchActionSymbols = tokenWorkbenchActionSymbols;
  state.tokenWorkbenchActionNavigates = tokenWorkbenchActionNavigates;
  state.tokenWorkbenchOverflowCount = tokenWorkbenchOverflowCount;
  state.tokenResultSubscriptionItemCount = tokenResultSubscriptionItemCount;
  state.tokenResultSubscriptionKindCount = tokenResultSubscriptionKindCount;
  state.tokenResultSubscriptionProfileCount = tokenResultSubscriptionProfileCount;
  state.tokenResultSubscriptionProfiles = tokenResultSubscriptionProfiles;
  state.tokenResultSubscriptionOriginCount = tokenResultSubscriptionOriginCount;
  state.tokenResultSubscriptionOrigins = tokenResultSubscriptionOrigins;
  state.tokenResultSubscriptionOpenCount = tokenResultSubscriptionOpenCount;
  state.tokenResultSubscriptionProbeCount = tokenResultSubscriptionProbeCount;
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
  state.tokenSubscriptionProbeCount = tokenSubscriptionProbeCount;
  state.tokenSubscriptionActionSymbolCount = tokenSubscriptionActionSymbolCount;
  state.tokenCopyFeedbackVisible = tokenCopyFeedbackVisible;
  state.tokenCopySymbolRestored = tokenCopySymbolRestored;
  state.tokenProbeFeedbackVisible = tokenProbeFeedbackVisible;
  state.tokenCardActionFeedback = tokenCardActionFeedback;
  state.tokenCardActionRestored = tokenCardActionRestored;
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
  state.opsWorkbenchCount = opsWorkbenchCount;
  state.opsWorkbenchStatCount = opsWorkbenchStatCount;
  state.opsWorkbenchStepCardCount = opsWorkbenchStepCardCount;
  state.opsWorkbenchSymbols = opsWorkbenchSymbols;
  state.opsWorkbenchOverflowCount = opsWorkbenchOverflowCount;
  state.opsActionPendingFeedback = opsActionPendingFeedback;
  state.opsActionPendingRestored = opsActionPendingRestored;
  state.opsResultVisible = opsResultVisible;
  state.opsResultFieldCount = opsResultFieldCount;
  state.opsResultSymbolCount = opsResultSymbolCount;
  state.opsResultChipCount = opsResultChipCount;
  state.opsVisualOverflowCount = opsVisualOverflowCount;
  state.unifiedWorkbenchStyles = unifiedWorkbenchStyles;
  state.calmOpsStylesheetCount = calmOpsStylesheetCount;
  state.calmOpsResponseStatus = calmOpsResponseStatus;
  const overflowingCommandCenter = Object.entries(workspaceCommandCenterOverflow).find(([, overflow]) => overflow > 0);
  const scrollingCommandRail = Object.entries(workspaceCommandRailScrollOverflow).find(([, overflow]) => overflow > 0);
  const overflowingContentPanel = Object.entries(contentPanelHeaderOverflow).find(([, overflow]) => overflow > 0);
  const contentPanelStyleMismatches = contentPanelStyles.filter(
    (entry) =>
      entry.backgroundColor !== "rgb(255, 255, 255)" ||
      entry.backgroundImage !== "none" ||
      !entry.borderRadius.startsWith("8px") ||
      entry.borderColor === "rgba(0, 0, 0, 0)" ||
      entry.boxShadow === "none" ||
      !entry.headerBackgroundColor.includes("248, 250, 252"),
  );
  if (
    workspaceCommandCenterCount !== 1 ||
    !workspaceCommandCenterStyle ||
    !authenticatedShellLayout ||
    authenticatedShellLayout.bodyClass !== 1 ||
    authenticatedShellLayout.topbarPosition !== "fixed" ||
    authenticatedShellLayout.topbarBrandVisible !== 0 ||
    authenticatedShellLayout.topbarLeft < 200 ||
    authenticatedShellLayout.topbarRightGap > 1 ||
    authenticatedShellLayout.sidebarPosition !== "sticky" ||
    authenticatedShellLayout.sidebarTop > 1 ||
    authenticatedShellLayout.sidebarHeight < authenticatedShellLayout.viewportHeight - 2 ||
    authenticatedShellLayout.workspaceTop > 1 ||
    authenticatedShellLayout.commandCenterTop <= authenticatedShellLayout.topbarHeight ||
    authenticatedShellLayout.overlapsCommandCenter !== 0 ||
    workspaceCommandCenterStyle.display !== "grid" ||
    !workspaceCommandCenterStyle.backgroundColor.includes("255, 255, 255") ||
    workspaceCommandCenterStyle.backgroundImage !== "none" ||
    !workspaceCommandCenterStyle.borderRadius.startsWith("8px") ||
    workspaceCommandCenterStyle.boxShadow === "none" ||
    overflowingCommandCenter ||
    scrollingCommandRail
  ) {
    throw new Error(`workspace command center is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (contentPanelCount !== 9 || contentPanelMetaCount !== 9 || overflowingContentPanel || contentPanelStyleMismatches.length > 0) {
    throw new Error(`content panels are incomplete, inconsistent, or overflowing: ${JSON.stringify(state)}`);
  }
  if (unifiedWorkbenchStyles.count < 8 || unifiedWorkbenchStyles.mismatches.length > 0) {
    throw new Error(`module workbenches do not share the v2 surface style: ${JSON.stringify(state)}`);
  }
  if (
    identityWorkbenchCount !== 1 ||
    identityWorkbenchStatCount !== 4 ||
    identityWorkbenchStageCardCount !== 3 ||
    !["身", "团", "员", "钥", "订"].every((symbol) => identityWorkbenchSymbols.includes(symbol)) ||
    identityWorkbenchOverflowCount > 0
  ) {
    throw new Error(`identity distribution workbench is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  const expectedTokenWorkbenchSymbols = ["通", "米", "箱"];
  const expectedTokenWorkbenchActionSymbols = ["订", "米", "↗"];
  if (
    tokenRowCount > 0 &&
    (tokenWorkbenchCount !== 1 ||
      tokenWorkbenchStatCount < 3 ||
      tokenWorkbenchFormatCount !== 3 ||
      tokenWorkbenchActionStripCount !== 1 ||
      tokenWorkbenchActionCount !== 3 ||
      expectedTokenWorkbenchSymbols.some((symbol, index) => tokenWorkbenchFormatSymbols[index] !== symbol) ||
      expectedTokenWorkbenchActionSymbols.some((symbol, index) => tokenWorkbenchActionSymbols[index] !== symbol) ||
      !tokenWorkbenchActionNavigates ||
      tokenWorkbenchOverflowCount > 0)
  ) {
    throw new Error(`token subscription workbench is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
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
    tokenResultSubscriptionProbeCount !== 3 ||
    tokenResultSubscriptionActionSymbolCount !== 9 ||
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
      tokenSubscriptionProbeCount !== tokenSubscriptionCopyCount ||
      tokenSubscriptionActionSymbolCount !== tokenSubscriptionCopyCount * 3)
  ) {
    throw new Error(`token subscription cards are incomplete: ${JSON.stringify(state)}`);
  }
  if (tokenSubscriptionCopyCount > 0 && !tokenCopyFeedbackVisible) {
    throw new Error(`token subscription copy feedback missing: ${JSON.stringify(state)}`);
  }
  if (tokenSubscriptionCopyCount > 0 && !tokenCopySymbolRestored) {
    throw new Error(`token subscription copy symbol restore missing: ${JSON.stringify(state)}`);
  }
  if (tokenSubscriptionCopyCount > 0 && !tokenProbeFeedbackVisible) {
    throw new Error(`token subscription probe feedback missing: ${JSON.stringify(state)}`);
  }
  if (
    tokenCardCount > 0 &&
    (!tokenCardActionFeedback ||
      tokenCardActionFeedback.pending !== 1 ||
      tokenCardActionFeedback.ariaBusy !== "true" ||
      tokenCardActionFeedback.extendDisabled !== 1 ||
      tokenCardActionFeedback.quotaDisabled !== 1 ||
      tokenCardActionFeedback.revokeDisabled !== 1 ||
      tokenCardActionFeedback.extendInputDisabled !== 1 ||
      tokenCardActionFeedback.quotaInputDisabled !== 1 ||
      tokenCardActionFeedback.label !== "续期中" ||
      tokenCardActionFeedback.symbol !== "…" ||
      tokenCardActionFeedback.status !== "续期 Token 中" ||
      tokenCardActionFeedback.tone !== "loading" ||
      tokenCardActionFeedback.overflow > 0 ||
      !tokenCardActionRestored ||
      tokenCardActionRestored.pending !== 0 ||
      tokenCardActionRestored.ariaBusy !== "" ||
      tokenCardActionRestored.extendDisabled !== 0 ||
      tokenCardActionRestored.extendInputDisabled !== 0 ||
      tokenCardActionRestored.label !== "续期" ||
      tokenCardActionRestored.symbol !== "+")
  ) {
    throw new Error(`token card action pending feedback missing: ${JSON.stringify(state)}`);
  }
  if (tokenVisualOverflowCount > 0 || tokenVisualOverlapCount > 0 || tokenActionsScrollOverflowCount > 0) {
    throw new Error(`token controls visually overflow or overlap: ${JSON.stringify(state)}`);
  }
  if (
    trafficWorkbenchCount !== 1 ||
    trafficWorkbenchStatCount !== 4 ||
    trafficWorkbenchSignalCardCount !== 3 ||
    !["量", "钥", "时", "出", "额", "日"].every((symbol) => trafficWorkbenchSymbols.includes(symbol)) ||
    trafficWorkbenchOverflowCount > 0 ||
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
    trafficEmptyStateCount !== trafficEmptyStateActionCount ||
    trafficEmptyStateCount !== trafficEmptyStateActionSymbolCount ||
    (trafficOutboundEmptyStateCount > 0 && !trafficEmptyStateActionNavigates) ||
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
      sourceVisualOverflowCount > 0 ||
      !sourceInlineSaveFeedback ||
      sourceInlineSaveFeedback.pending !== 1 ||
      sourceInlineSaveFeedback.ariaBusy !== "true" ||
      sourceInlineSaveFeedback.saveDisabled !== 1 ||
      sourceInlineSaveFeedback.cancelDisabled !== 1 ||
      sourceInlineSaveFeedback.inputDisabled !== 1 ||
      sourceInlineSaveFeedback.selectDisabled !== 1 ||
      sourceInlineSaveFeedback.label !== "保存中" ||
      sourceInlineSaveFeedback.symbol !== "…" ||
      sourceInlineSaveFeedback.status !== "保存来源中" ||
      sourceInlineSaveFeedback.tone !== "loading" ||
      sourceInlineSaveFeedback.overflow > 0 ||
      !sourceInlineSaveRestored ||
      sourceInlineSaveRestored.pending !== 0 ||
      sourceInlineSaveRestored.ariaBusy !== "" ||
      sourceInlineSaveRestored.editDisabled !== 0 ||
      sourceInlineSaveRestored.saveCount !== 0 ||
      sourceInlineSaveRestored.label !== "编辑" ||
      sourceInlineSaveRestored.symbol !== "✎")
  ) {
    throw new Error(`source cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    sourceEditCount > 0 &&
    (sourceWorkbenchCount !== 1 ||
      sourceWorkbenchStatCount < 4 ||
      sourceWorkbenchTypeCardCount < 1 ||
      sourceWorkbenchOverflowCount > 0 ||
      !["源", "健", "刷", "异"].every((symbol) => sourceWorkbenchSymbols.includes(symbol)))
  ) {
    throw new Error(`source workbench is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    virtualNodeEditCount > 0 &&
    (virtualNodeWorkbenchCount !== 1 ||
      virtualNodeWorkbenchStatCount < 4 ||
      virtualNodeWorkbenchStrategyCardCount < 1 ||
      virtualNodeWorkbenchOverflowCount > 0 ||
      !["网", "启", "端", "筛"].every((symbol) => virtualNodeWorkbenchSymbols.includes(symbol)) ||
      virtualNodeCardCount !== virtualNodeEditCount ||
      virtualNodeSummaryChipCount !== virtualNodeCardCount * 4 ||
      virtualNodeActionButtonSymbolCount !== virtualNodeCardCount ||
      virtualNodeVisualOverflowCount > 0)
  ) {
    throw new Error(`virtual node cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    policyEditCount > 0 &&
    (policyWorkbenchCount !== 1 ||
      policyWorkbenchStatCount < 4 ||
      policyWorkbenchScopeCardCount < 1 ||
      policyWorkbenchOverflowCount > 0 ||
      !["策", "启", "限", "网"].every((symbol) => policyWorkbenchSymbols.includes(symbol)) ||
      policyCardCount !== policyEditCount ||
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
    (nodeWorkbenchCount !== 1 ||
      nodeWorkbenchStatCount < 4 ||
      nodeWorkbenchRegionCardCount < Math.min(nodeRegionCount, 4) ||
      nodeWorkbenchOverflowCount > 0 ||
      !["点", "活", "区", "协"].every((symbol) => nodeWorkbenchSymbols.includes(symbol)))
  ) {
    throw new Error(`node pool workbench is incomplete or overflowing: ${JSON.stringify(state)}`);
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
  if (nodeDetailVisible && nodeDetailOverlapCount > 0) {
    throw new Error(`node detail should not cover node cards on desktop: ${JSON.stringify(state)}`);
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
    opsWorkbenchCount !== 1 ||
    opsWorkbenchStatCount !== 4 ||
    opsWorkbenchStepCardCount < 3 ||
    !["运", "闭", "配", "入", "上", "源", "点", "网", "钥", "策"].every((symbol) => opsWorkbenchSymbols.includes(symbol)) ||
    opsWorkbenchOverflowCount > 0
  ) {
    throw new Error(`ops delivery workbench is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    opsActionCardCount > 0 &&
    (!opsActionPendingFeedback ||
      opsActionPendingFeedback.pending !== 1 ||
      opsActionPendingFeedback.ariaBusy !== "true" ||
      opsActionPendingFeedback.gridBusy !== "true" ||
      opsActionPendingFeedback.checkDisabled !== 1 ||
      opsActionPendingFeedback.publishDisabled !== 1 ||
      opsActionPendingFeedback.rollbackDisabled !== 1 ||
      opsActionPendingFeedback.restartDisabled !== 1 ||
      opsActionPendingFeedback.label !== "检查配置中" ||
      opsActionPendingFeedback.symbol !== "…" ||
      opsActionPendingFeedback.status !== "检查配置中" ||
      opsActionPendingFeedback.tone !== "loading" ||
      opsActionPendingFeedback.overflow > 0 ||
      !opsActionPendingRestored ||
      opsActionPendingRestored.pending !== 0 ||
      opsActionPendingRestored.ariaBusy !== "" ||
      opsActionPendingRestored.gridBusy !== "" ||
      opsActionPendingRestored.checkDisabled !== 0 ||
      opsActionPendingRestored.publishDisabled !== 0 ||
      opsActionPendingRestored.rollbackDisabled !== 0 ||
      opsActionPendingRestored.restartDisabled !== 0 ||
      opsActionPendingRestored.label !== "检查配置" ||
      opsActionPendingRestored.symbol !== "检")
  ) {
    throw new Error(`ops action pending feedback missing: ${JSON.stringify(state)}`);
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
    !mobileMoreMenuVisible ||
    !mobileMoreMenuHidesAfterSelection ||
    !mobileMoreToggleActiveForOps ||
    !mobilePoliciesVisible ||
    !mobileNodesVisible ||
    !mobileOpsVisible ||
    mobileDockOverflow > 2 ||
    mobileMoreMenuOverflow > 2 ||
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
    mobileDockCount !== 5 ||
    mobileDockDirectNavCount !== primaryMobileViews.length ||
    mobileDockSymbolCount !== 5 ||
    mobileDockBadgeCount !== 5 ||
    mobileMoreToggleCount !== 1 ||
    mobileMoreNavCount !== overflowMobileViews.length ||
    mobileMoreBadgeCount !== overflowMobileViews.length ||
    populatedNavBadgeCount !== viewNames.length * 2 ||
    dashboardViewCount !== viewNames.length
  ) {
    throw new Error(`dashboard navigation is incomplete: ${JSON.stringify(state)}`);
  }
  const expectedFormDrawerSymbols = ["源", "导", "网", "团", "员", "钥", "策"];
  const overflowingFormDrawerHeader = Object.entries(formDrawerHeaderOverflow).find(([, overflow]) => overflow > 0);
  const overflowingFormSubmit = Object.entries(formSubmitOverflow).find(([, overflow]) => overflow > 0);
  const narrowFormDrawer = Object.entries(formDrawerNarrowCount).find(([, count]) => count > 0);
  if (
    formDrawerCount !== 7 ||
    collapsedFormDrawerCount !== 7 ||
    formDrawerSymbolCount !== 7 ||
    formDrawerToggleSymbolCount !== 7 ||
    formDrawerCancelCount !== 7 ||
    formDrawerCancelSymbolCount !== 7 ||
    formDrawerDraftCount !== 7 ||
    formSubmitButtonSymbolCount !== 7 ||
    loginButtonSymbolCount !== 1 ||
    loginButtonOverflowCount > 0 ||
    refreshButtonSymbolCount !== 1 ||
    logoutButtonSymbolCount !== 1 ||
    refreshButtonOverflowCount > 0 ||
    topbarButtonOverflowCount > 0 ||
    expectedFormDrawerSymbols.some((symbol, index) => formDrawerSymbols[index] !== symbol) ||
    overflowingFormDrawerHeader ||
    overflowingFormSubmit ||
    narrowFormDrawer
  ) {
    throw new Error(`form drawers should be collapsed and visually stable by default: ${JSON.stringify(state)}`);
  }
  if (
    formDrawerSideSheetFeedback.bodyHasOpen !== 1 ||
    formDrawerSideSheetFeedback.activeDrawer !== "source-form" ||
    formDrawerSideSheetFeedback.role !== "dialog" ||
    formDrawerSideSheetFeedback.ariaModal !== "true" ||
    formDrawerSideSheetFeedback.toggleExpanded !== "true" ||
    formDrawerSideSheetFeedback.position !== "fixed" ||
    formDrawerSideSheetFeedback.zIndex < 90 ||
    formDrawerSideSheetFeedback.openDrawerCount !== 1 ||
    formDrawerSideSheetFeedback.formBodyOverflowY !== "auto" ||
    formDrawerSideSheetFeedback.width <= 0 ||
    formDrawerSideSheetFeedback.width > formDrawerSideSheetFeedback.viewportWidth ||
    formDrawerSideSheetFeedback.pageOverflow > 2
  ) {
    throw new Error(`form drawer should open as a stable side sheet: ${JSON.stringify(state)}`);
  }
  if (
    formDrawerDraftFeedback.dirty !== 1 ||
    formDrawerDraftFeedback.draftHidden !== false ||
    formDrawerDraftFeedback.text !== "草稿未提交" ||
    formDrawerDraftFeedback.overflow > 0
  ) {
    throw new Error(`form drawer should show a stable unsaved draft marker after input: ${JSON.stringify(state)}`);
  }
  if (
    formDrawerEscapeFeedback.bodyHasOpen !== 0 ||
    formDrawerEscapeFeedback.activeDrawer ||
    formDrawerEscapeFeedback.role ||
    formDrawerEscapeFeedback.ariaModal ||
    formDrawerEscapeFeedback.toggleExpanded !== "false" ||
    formDrawerEscapeFeedback.dirty !== 1 ||
    formDrawerEscapeFeedback.value !== "QA inline source" ||
    !formDrawerEscapeFeedback.status?.includes("已收起：添加来源") ||
    formDrawerEscapeFeedback.tone !== "info" ||
    formDrawerEscapeFeedback.pageOverflow > 2
  ) {
    throw new Error(`form drawer escape should collapse the side sheet without losing draft: ${JSON.stringify(state)}`);
  }
  if (
    !formDrawerCancelFeedback.status?.includes("已取消：添加来源") ||
    formDrawerCancelFeedback.tone !== "info" ||
    formDrawerCancelFeedback.value !== "" ||
    formDrawerCancelFeedback.inlineCount !== 0 ||
    formDrawerCancelFeedback.ariaInvalid !== "" ||
    formDrawerCancelFeedback.dirty !== 0 ||
    formDrawerCancelFeedback.draftHidden !== true
  ) {
    throw new Error(`form drawer cancel should clear draft input and report feedback: ${JSON.stringify(state)}`);
  }
  if (
    formDrawerSubmitFeedback.submitting !== 1 ||
    formDrawerSubmitFeedback.ariaBusy !== "true" ||
    formDrawerSubmitFeedback.submitDisabled !== 1 ||
    formDrawerSubmitFeedback.cancelDisabled !== 1 ||
    formDrawerSubmitFeedback.inputDisabled !== 1 ||
    formDrawerSubmitFeedback.label !== "添加中" ||
    formDrawerSubmitFeedback.symbol !== "…" ||
    !formDrawerSubmitFeedback.status?.includes("正在提交：添加来源") ||
    formDrawerSubmitFeedback.tone !== "loading" ||
    formDrawerSubmitFeedback.overflow > 0
  ) {
    throw new Error(`form drawer submit should show a stable pending state: ${JSON.stringify(state)}`);
  }
  if (
    formDrawerSubmitRestored.collapsed !== 1 ||
    formDrawerSubmitRestored.submitting !== 0 ||
    formDrawerSubmitRestored.ariaBusy ||
    formDrawerSubmitRestored.submitDisabled !== 0 ||
    formDrawerSubmitRestored.cancelDisabled !== 0 ||
    formDrawerSubmitRestored.inputDisabled !== 0 ||
    formDrawerSubmitRestored.label !== "添加" ||
    formDrawerSubmitRestored.symbol !== "+"
  ) {
    throw new Error(`form drawer submit should restore controls after completion: ${JSON.stringify(state)}`);
  }
  if (
    logoutButtonPendingFeedback.disabled !== 1 ||
    logoutButtonPendingFeedback.pending !== 1 ||
    logoutButtonPendingFeedback.label !== "退出中" ||
    logoutButtonPendingFeedback.symbol !== "…" ||
    logoutButtonPendingFeedback.status !== "退出中" ||
    logoutButtonPendingFeedback.tone !== "loading" ||
    logoutButtonPendingFeedback.ariaBusy !== "true" ||
    logoutButtonPendingFeedback.overflow > 0 ||
    logoutButtonCompletedFeedback.loginVisible !== true ||
    logoutButtonCompletedFeedback.appHidden !== 1 ||
    logoutButtonCompletedFeedback.logoutHidden !== 1 ||
    logoutButtonCompletedFeedback.logoutDisabled !== 0 ||
    logoutButtonCompletedFeedback.status !== "未登录" ||
    logoutButtonCompletedFeedback.tone !== "warning" ||
    logoutButtonCompletedFeedback.ariaBusy
  ) {
    throw new Error(`logout should show pending feedback and return to login cleanly: ${JSON.stringify(state)}`);
  }
  if (
    refreshButtonPendingFeedback.disabled !== 1 ||
    refreshButtonPendingFeedback.pending !== 1 ||
    refreshButtonPendingFeedback.label !== "刷新中" ||
    refreshButtonPendingFeedback.symbol !== "…" ||
    refreshButtonPendingFeedback.status !== "刷新中" ||
    refreshButtonPendingFeedback.tone !== "loading" ||
    refreshButtonPendingFeedback.ariaBusy !== "true" ||
    refreshButtonPendingFeedback.overflow > 0 ||
    refreshButtonPendingRestored.disabled !== 0 ||
    refreshButtonPendingRestored.pending !== 0 ||
    refreshButtonPendingRestored.label !== "刷新" ||
    refreshButtonPendingRestored.symbol !== "↻" ||
    refreshButtonPendingRestored.status !== "已连接" ||
    refreshButtonPendingRestored.tone !== "success" ||
    refreshButtonPendingRestored.ariaBusy
  ) {
    throw new Error(`refresh should show and restore a stable pending state: ${JSON.stringify(state)}`);
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
  const missingInsightView = Object.entries(workspaceInsightCounts).find(([, count]) => count !== 1);
  const emptyInsightTitleView = Object.entries(workspaceInsightTitles).find(([, title]) => title.length === 0);
  const missingInsightActionView = Object.entries(workspaceInsightActionCounts).find(([, count]) => count !== 1);
  const missingInsightSymbolView = Object.entries(workspaceInsightSymbolCounts).find(([, count]) => count !== 1);
  const missingInsightActionSymbolView = Object.entries(workspaceInsightActionSymbolCounts).find(([, count]) => count !== 1);
  const overflowingInsightView = Object.entries(workspaceInsightOverflow).find(([, overflow]) => overflow > 0);
  if (
    missingInsightView ||
    emptyInsightTitleView ||
    missingInsightActionView ||
    missingInsightSymbolView ||
    missingInsightActionSymbolView ||
    overflowingInsightView ||
    !workspaceInsightActionNavigates
  ) {
    throw new Error(`dashboard workspace insight is incomplete: ${JSON.stringify(state)}`);
  }
  const missingPrimaryActionView = Object.entries(workspacePrimaryActionCounts).find(([, count]) => count !== 1);
  const missingPrimaryActionSymbolView = Object.entries(workspacePrimaryActionSymbolCounts).find(([, count]) => count !== 1);
  const overflowingPrimaryActionView = Object.entries(workspacePrimaryActionOverflow).find(([, overflow]) => overflow > 0);
  const missingPrimaryActionFocus = workspacePrimaryActionFocus.access !== "name" || workspacePrimaryActionFocus.identity !== "user_id";
  const missingPrimaryActionHighlight = workspacePrimaryActionHighlight.access !== 1 || workspacePrimaryActionHighlight.identity !== 1;
  const missingPrimaryActionStatus =
    !workspacePrimaryActionStatus.access.includes("已定位：添加来源") ||
    workspacePrimaryActionStatusTone.access !== "info" ||
    !workspacePrimaryActionStatus.identity.includes("已定位：创建 Token") ||
    workspacePrimaryActionStatusTone.identity !== "info";
  const missingRailTarget =
    !activeRailTargets.access?.includes("source-form") || !activeRailTargets.identity?.includes("token-form");
  const expectedPrimaryActionLabels = {
    access: "添加来源",
    nodes: "创建网关",
    identity: "签发 Token",
    policies: "创建策略",
    traffic: "查看运维",
    ops: "检查收口",
  };
  const primaryActionLabelMismatch = Object.entries(workspacePrimaryActionLabels).find(([view, label]) => {
    if (view === "overview") return !["继续初始化", "去运维发布"].includes(label);
    return expectedPrimaryActionLabels[view] && label !== expectedPrimaryActionLabels[view];
  });
  if (
    missingPrimaryActionView ||
    missingPrimaryActionSymbolView ||
    overflowingPrimaryActionView ||
    missingPrimaryActionFocus ||
    missingPrimaryActionHighlight ||
    missingPrimaryActionStatus ||
    missingRailTarget ||
    primaryActionLabelMismatch ||
    !workspacePrimaryActionOpensTokenForm
  ) {
    throw new Error(`dashboard workspace primary action is incomplete: ${JSON.stringify(state)}`);
  }
  if (
    !requiredFieldFeedback.status.includes("请先补齐：名称") ||
    requiredFieldFeedback.tone !== "warning" ||
    requiredFieldFeedback.activeName !== "name" ||
    requiredFieldFeedback.highlighted !== 1 ||
    requiredFieldFeedback.inlineCount !== 1 ||
    !requiredFieldFeedback.inlineText.includes("请填写名称") ||
    requiredFieldFeedback.ariaInvalid !== "true" ||
    !requiredFieldFeedback.describedBy.includes("source-form-name-feedback") ||
    requiredFieldFeedbackCleared.inlineCount !== 0 ||
    requiredFieldFeedbackCleared.ariaInvalid ||
    requiredFieldFeedbackCleared.invalidClass !== 0
  ) {
    throw new Error(`form required field feedback is missing: ${JSON.stringify(state)}`);
  }
  const sparseRailView = Object.entries(viewRailButtonCounts).find(([, count]) => count < 1);
  if (sparseRailView || !moduleRailOpensTokenForm) {
    throw new Error(`dashboard module rail is incomplete: ${JSON.stringify(state)}`);
  }
  const expectedRailSymbols = {
    overview: ["概", "向", "步"],
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
  const expectedRailLabels = {
    overview: ["运行概览", "初始化", "下一步"],
  };
  const railLabelMismatch = Object.entries(expectedRailLabels).find(([view, expected]) => {
    const actual = viewRailLabels[view] || [];
    return expected.some((label, index) => actual[index] !== label);
  });
  if (railLabelMismatch) {
    throw new Error(`dashboard module rail labels should match the human setup guide: ${JSON.stringify(state)}`);
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
    overviewHeroCount !== 1 ||
    !overviewHeroTitle ||
    overviewHeroActionCount !== 1 ||
    overviewHeroActionSymbolCount !== 1 ||
    overviewHeroStatCount !== 3 ||
    overviewHeroStatSymbolCount !== 3 ||
    overviewHeroOverflowCount > 0 ||
    !overviewHeroActionNavigates ||
    overviewMetricCount !== 6 ||
    overviewMetricSymbolCount !== 6 ||
    overviewMetricValueCount !== 6 ||
    overviewMetricOverflowCount > 0
  ) {
    throw new Error(`overview metric cards are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (
    overviewInsightPanelCount !== 3 ||
    overviewInsightSymbolCount !== 3 ||
    overviewInsightTitleCount !== 3 ||
    overviewHealthPillCount < 3 ||
    overviewRegionRowCount < 1 ||
    overviewSyncRowCount < 1 ||
    overviewDeliveryRowCount < 4 ||
    overviewInsightOverflowCount > 0
  ) {
    throw new Error(`overview insight panels are incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  const expectedOverviewReadinessSymbols = ["源", "点", "网", "身", "订", "策", "发"];
  const expectedOverviewGuideActionLabels = ["去复制订阅", "检查收口", "发布配置"];
  const expectedOverviewGuideActionTargets = ["tokens", "delivery-readiness", "config-publish"];
  if (
    overviewReadinessCount !== 7 ||
    overviewReadinessIndexCount !== 7 ||
    overviewReadinessSymbolCount !== 7 ||
    overviewReadinessStateCount !== 7 ||
    overviewReadinessStates.some((state) => !["就绪", "待补"].includes(state)) ||
    expectedOverviewReadinessSymbols.some((symbol, index) => overviewReadinessSymbols[index] !== symbol) ||
    overviewReadinessOverflowCount > 0 ||
    overviewGuideProgressCount !== 1 ||
    !/^\d+\/7$/.test(overviewGuideProgressValue) ||
    overviewGuideCurrentCount !== 1 ||
    overviewGuideCurrentActionSymbolCount !== 1 ||
    overviewGuideActionStripCount !== 1 ||
    overviewGuideActionButtonCount !== 3 ||
    overviewGuideActionSymbolCount !== 3 ||
    expectedOverviewGuideActionLabels.some((label, index) => overviewGuideActionLabels[index] !== label) ||
    expectedOverviewGuideActionTargets.some((target, index) => overviewGuideActionTargets[index] !== target) ||
    !overviewGuideTokenActionNavigates ||
    !overviewGuidePublishActionNavigates ||
    overviewGuideHintCount !== 7 ||
    overviewGuideOverflowCount > 0 ||
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
    calmOpsResponseStatus !== 200 ||
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
