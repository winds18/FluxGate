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
  const mobileDockCount = await page.locator(".mobile-dock [data-view-nav]").count();
  const dashboardViewCount = await page.locator("[data-dashboard-view]").count();
  const formDrawerCount = await page.locator("[data-form-drawer]").count();
  const collapsedFormDrawerCount = await page.locator("[data-form-drawer].is-collapsed").count();
  const panelCountCount = await page.locator(".panel-count").count();
  const populatedPanelCountCount = await page
    .locator(".panel-count")
    .evaluateAll((elements) => elements.filter((element) => element.textContent.trim().length > 0).length);
  const viewOverflow = {};
  const viewContextChipCounts = {};
  const viewRailButtonCounts = {};
  for (const view of viewNames) {
    await switchView(view);
    viewOverflow[view] = await pageHorizontalOverflow();
    viewContextChipCounts[view] = await page.locator("#view-context .context-chip").count();
    viewRailButtonCounts[view] = await page.locator("#view-rail .view-rail-button").count();
  }
  await switchView("overview");
  const overviewReadinessCount = await page.locator("#overview-readiness .readiness-item").count();
  const overviewNextStepButtonCount = await page.locator("#overview-next-step [data-overview-jump]").count();

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
      resultCard.querySelectorAll(".token-result-heading, .token-subscription-item, .token-subscription-heading, code, button"),
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
  const tokenSubscriptionItemCount = await page.locator("#tokens .token-subscription-item").count();
  const tokenSubscriptionCopyCount = await page.locator("#tokens button[data-token-action='copy-subscription']").count();
  const tokenRotateSubscriptionCount = await page.locator("#tokens button[data-token-action='rotate-subscription']").count();
  let tokenCopyFeedbackVisible = false;
  if (tokenSubscriptionCopyCount > 0) {
    const firstCopyButton = page.locator("#tokens button[data-token-action='copy-subscription']").first();
    await firstCopyButton.click();
    await expect(firstCopyButton).toHaveText("已复制", { timeout: 5000 });
    await expect(page.locator("#status")).toContainText("订阅地址已复制", { timeout: 5000 });
    tokenCopyFeedbackVisible = true;
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
          ".token-card-heading, .token-card-field, .token-card-subscriptions, .token-subscription-item, .token-subscription-heading, .token-card-actions",
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
  state.mobileDockCount = mobileDockCount;
  state.dashboardViewCount = dashboardViewCount;
  state.formDrawerCount = formDrawerCount;
  state.collapsedFormDrawerCount = collapsedFormDrawerCount;
  state.panelCountCount = panelCountCount;
  state.populatedPanelCountCount = populatedPanelCountCount;
  state.viewContextChipCounts = viewContextChipCounts;
  state.viewRailButtonCounts = viewRailButtonCounts;
  state.moduleRailOpensTokenForm = moduleRailOpensTokenForm;
  state.overviewReadinessCount = overviewReadinessCount;
  state.overviewNextStepButtonCount = overviewNextStepButtonCount;
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
  state.tokenResultVisualOverflowCount = tokenResultVisualOverflowCount;
  state.quotaMeterCount = quotaMeterCount;
  state.quotaUsageVisible = quotaUsageVisible;
  state.tokenExtendInputCount = tokenExtendInputCount;
  state.tokenQuotaInputCount = tokenQuotaInputCount;
  state.tokenActionFieldCount = tokenActionFieldCount;
  state.tokenSubscriptionItemCount = tokenSubscriptionItemCount;
  state.tokenSubscriptionCopyCount = tokenSubscriptionCopyCount;
  state.tokenCopyFeedbackVisible = tokenCopyFeedbackVisible;
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
  if (tokenResultSubscriptionItemCount !== 3 || tokenResultVisualOverflowCount > 0) {
    throw new Error(`token subscription result card is incomplete or overflowing: ${JSON.stringify(state)}`);
  }
  if (tokenRowCount > 0 && tokenRotateSubscriptionCount !== tokenRowCount) {
    throw new Error(`token subscription rotate controls missing: ${JSON.stringify(state)}`);
  }
  if (tokenSubscriptionCopyCount > 0 && tokenSubscriptionItemCount !== tokenSubscriptionCopyCount) {
    throw new Error(`token subscription cards are incomplete: ${JSON.stringify(state)}`);
  }
  if (tokenSubscriptionCopyCount > 0 && !tokenCopyFeedbackVisible) {
    throw new Error(`token subscription copy feedback missing: ${JSON.stringify(state)}`);
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
  if (dashboardNavCount !== viewNames.length || mobileDockCount !== viewNames.length || dashboardViewCount !== viewNames.length) {
    throw new Error(`dashboard navigation is incomplete: ${JSON.stringify(state)}`);
  }
  if (formDrawerCount !== 7 || collapsedFormDrawerCount !== 7) {
    throw new Error(`form drawers should be collapsed by default: ${JSON.stringify(state)}`);
  }
  if (panelCountCount !== 9 || populatedPanelCountCount !== 9) {
    throw new Error(`module panel count badges are incomplete: ${JSON.stringify(state)}`);
  }
  const sparseContextView = Object.entries(viewContextChipCounts).find(([, count]) => count < 3);
  if (sparseContextView) {
    throw new Error(`dashboard view context is incomplete: ${JSON.stringify(state)}`);
  }
  const sparseRailView = Object.entries(viewRailButtonCounts).find(([, count]) => count < 1);
  if (sparseRailView || !moduleRailOpensTokenForm) {
    throw new Error(`dashboard module rail is incomplete: ${JSON.stringify(state)}`);
  }
  if (overviewReadinessCount !== 5 || overviewNextStepButtonCount !== 1) {
    throw new Error(`overview readiness board is incomplete: ${JSON.stringify(state)}`);
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
