const statusEl = document.querySelector("#status");
const metricsEl = document.querySelector("#metrics");
const sourcesEl = document.querySelector("#sources");
const nodesEl = document.querySelector("#nodes");
const tokensEl = document.querySelector("#tokens");
const refreshEl = document.querySelector("#refresh");

refreshEl.addEventListener("click", load);
load();

async function load() {
  statusEl.textContent = "刷新中";
  try {
    const [overview, sources, nodes, tokens] = await Promise.all([
      getJSON("/api/overview"),
      getJSON("/api/sources"),
      getJSON("/api/nodes"),
      getJSON("/api/tokens"),
    ]);
    renderMetrics(overview);
    renderTable(sourcesEl, sources, ["id", "name", "type", "display_prefix", "status"]);
    renderTable(nodesEl, nodes, ["id", "source_name", "raw_name", "display_name", "protocol", "status"]);
    renderTable(tokensEl, tokens, ["id", "user_id", "token_prefix", "name", "status", "quota_bytes"]);
    statusEl.textContent = "已连接";
  } catch (error) {
    statusEl.textContent = "异常";
    metricsEl.innerHTML = `<div class="empty">${escapeHTML(error.message)}</div>`;
  }
}

async function getJSON(path) {
  const response = await fetch(path);
  if (!response.ok) {
    throw new Error(`${path} ${response.status}`);
  }
  return response.json();
}

function renderMetrics(data) {
  const items = [
    ["团队", data.teams],
    ["用户", data.users],
    ["Token", data.tokens],
    ["来源", data.sources],
    ["节点", data.nodes],
    ["虚拟节点", data.virtual_nodes],
  ];
  metricsEl.innerHTML = items
    .map(([label, value]) => `<div class="metric"><span>${label}</span><strong>${value ?? 0}</strong></div>`)
    .join("");
}

function renderTable(target, rows, columns) {
  if (!rows || rows.length === 0) {
    target.innerHTML = `<div class="empty">暂无数据</div>`;
    return;
  }
  target.innerHTML = `
    <table>
      <thead>
        <tr>${columns.map((column) => `<th>${escapeHTML(column)}</th>`).join("")}</tr>
      </thead>
      <tbody>
        ${rows
          .map(
            (row) =>
              `<tr>${columns
                .map((column) => `<td>${formatCell(row[column])}</td>`)
                .join("")}</tr>`,
          )
          .join("")}
      </tbody>
    </table>
  `;
}

function formatCell(value) {
  if (value === null || value === undefined || value === "") return "";
  if (typeof value === "string" && value.length > 38) {
    return `<code>${escapeHTML(value.slice(0, 38))}…</code>`;
  }
  return escapeHTML(String(value));
}

function escapeHTML(value) {
  return value.replace(/[&<>"']/g, (char) => {
    switch (char) {
      case "&":
        return "&amp;";
      case "<":
        return "&lt;";
      case ">":
        return "&gt;";
      case '"':
        return "&quot;";
      default:
        return "&#039;";
    }
  });
}
