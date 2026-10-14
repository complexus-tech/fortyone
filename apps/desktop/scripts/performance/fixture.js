const nextFrame = () =>
  new Promise((resolve) => requestAnimationFrame(resolve));
const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const errors = [];
window.addEventListener("error", (e) => errors.push(e.message));
window.addEventListener("unhandledrejection", (e) =>
  errors.push(String(e.reason)),
);
const quantile = (values, p) =>
  [...values].sort((a, b) => a - b)[
    Math.min(values.length - 1, Math.floor(values.length * p))
  ] ?? null;
const summarize = (values) => ({
  n: values.length,
  p50: quantile(values, 0.5),
  p95: quantile(values, 0.95),
  p99: quantile(values, 0.99),
  max: values.length ? Math.max(...values) : null,
  over25: values.filter((v) => v > 25).length,
  over50: values.filter((v) => v > 50).length,
});
const send = (payload) =>
  window.webkit.messageHandlers.result.postMessage(payload);
const board = document.getElementById("board");
const columns = [];
const allCards = [];
for (let col = 0; col < 4; col++) {
  const column = document.createElement("section");
  column.className = "column";
  const heading = document.createElement("h2");
  heading.textContent = ["Backlog", "In progress", "Review", "Done"][col];
  const scroller = document.createElement("div");
  scroller.className = "scroller";
  column.append(heading, scroller);
  board.append(column);
  columns.push(scroller);
  for (let row = 0; row < CONFIG.cards / 4; row++) {
    const card = document.createElement("article");
    card.className = "card rounded-xl";
    card.innerHTML = `<div class="card-title">Improve workspace experience ${col + row * 4}</div><div class="meta"><span class="chip rounded-md">In progress</span><span class="chip rounded-md">Priority</span><button class="chip rounded-md">Assign</button><span class="chip rounded-md">Sprint 24</span></div><div class="reference">FRT-${1040 + row * 4 + col}</div>`;
    scroller.append(card);
    allCards.push(card);
  }
}
const pageTemplate = document.createElement("template");
pageTemplate.innerHTML = board.innerHTML;
let controller;
let hoverTarget = null;
let hoveredCard = null;
const ancestorChain = (node) => {
  const nodes = [];
  for (let n = node; n; n = n.parentElement) nodes.push(n);
  return nodes;
};
function moveHover(target) {
  const previous = ancestorChain(hoverTarget);
  const next = ancestorChain(target);
  for (const node of previous)
    if (!next.includes(node))
      node.dispatchEvent(
        new MouseEvent("mouseleave", { relatedTarget: target }),
      );
  if (hoveredCard) hoveredCard.classList.remove("is-hover");
  hoveredCard = target?.closest(".card");
  if (hoveredCard) hoveredCard.classList.add("is-hover");
  for (const node of next.reverse())
    if (!previous.includes(node))
      node.dispatchEvent(
        new MouseEvent("mouseenter", { relatedTarget: hoverTarget }),
      );
  hoverTarget = target;
}
function operation(step) {
  const column = columns[step % 4];
  if (step % 8 === 0)
    column.scrollTop =
      (column.scrollTop + 230) %
      Math.max(1, column.scrollHeight - column.clientHeight);
  const row = Math.min(
    column.children.length - 1,
    Math.floor(column.scrollTop / 158) + (step % 3),
  );
  const card = column.children[row];
  moveHover(step % 2 ? card.querySelectorAll(".chip")[step % 4] : card);
  if (step % 3 === 0) card.classList.toggle("selected");
  if (step % 5 === 0) card.querySelector(".chip").classList.toggle("updated");
  if (step === 36) board.classList.toggle("condensed");
}
function reset() {
  moveHover(null);
  board.classList.remove("condensed");
  for (const card of allCards) {
    card.classList.remove("selected");
    card.querySelector(".chip").classList.remove("updated");
  }
  for (const column of columns) column.scrollTop = 0;
}
let phaseActive = false;
let frameGaps = [];
let lastFrame = null;
let rafCount = 0;
function frame(time) {
  if (phaseActive && lastFrame !== null) frameGaps.push(time - lastFrame);
  lastFrame = time;
  rafCount++;
  requestAnimationFrame(frame);
}
requestAnimationFrame(frame);
async function phase(name, work) {
  frameGaps = [];
  lastFrame = null;
  let mutations = 0;
  const handlerTimes = [];
  const observer = new MutationObserver(
    (records) => (mutations += records.length),
  );
  observer.observe(board, {
    subtree: true,
    attributes: true,
    attributeFilter: ["style", "data-hyperellipse-host"],
  });
  send({ type: "phase-start", name, id: CONFIG.id });
  await nextFrame();
  phaseActive = true;
  const started = performance.now();
  await work(handlerTimes);
  await nextFrame();
  await nextFrame();
  const duration = performance.now() - started;
  phaseActive = false;
  observer.disconnect();
  const result = {
    name,
    duration,
    frames: summarize(frameGaps),
    handler: summarize(handlerTimes),
    styleMutations: mutations,
    visibility: document.visibilityState,
    hasFocus: document.hasFocus(),
  };
  send({ type: "phase-end", id: CONFIG.id, result });
  return result;
}
(async () => {
  await document.fonts.ready;
  await nextFrame();
  await nextFrame();
  const support = CSS.supports("corner-shape", "squircle");
  const start = performance.now();
  if (CONFIG.mode === "on") {
    document.documentElement.dataset.squircleFallback = "true";
    controller = registerHyperellipse({ pendingRadiusScale: 0.75 / 2.2 });
  }
  await nextFrame();
  await nextFrame();
  const startupMs = performance.now() - start;
  for (let step = 0; step < 16; step++) {
    operation(step);
    await delay(50);
  }
  reset();
  await delay(600);
  await nextFrame();
  const initial = {
    viewport: [innerWidth, innerHeight],
    cards: allCards.length,
    targets: document.querySelectorAll("[class*=rounded]").length,
    painted: document.querySelectorAll("[data-hyperellipse-host]").length,
    supported: support,
    controllerActive: controller?.active ?? false,
    visibility: document.visibilityState,
    hasFocus: document.hasFocus(),
    startupMs,
    rafCount,
  };
  send({ type: "ready", id: CONFIG.id, initial });
  await phase("idle", () => delay(1500));
  await phase("workload", async (times) => {
    for (let step = 0; step < 72; step++) {
      const t = performance.now();
      operation(step);
      times.push(performance.now() - t);
      await delay(60);
    }
  });
  reset();
  await delay(500);
  await phase("page-ticket", async (times) => {
    for (let step = 0; step < 8; step++) {
      let t = performance.now();
      board.replaceChildren(pageTemplate.content.cloneNode(true));
      times.push(performance.now() - t);
      await nextFrame();
      await nextFrame();
      await delay(80);
      const modal = document.createElement("div");
      modal.className = "ticket rounded-xl";
      modal.innerHTML =
        "<h2>Ticket details</h2><p>Representative ticket content</p><div class=meta>" +
        Array.from(
          { length: 12 },
          (_, i) =>
            '<button class="chip rounded-md">Property ' + i + "</button>",
        ).join("") +
        "</div>";
      t = performance.now();
      document.body.append(modal);
      times.push(performance.now() - t);
      await nextFrame();
      await nextFrame();
      await delay(80);
      modal.remove();
    }
  });
  await delay(500);
  await phase("idle-after", () => delay(1000));
  send({ type: "trial-end", id: CONFIG.id, initial, errors });
})().catch((error) =>
  send({ type: "fatal", error: String(error), id: CONFIG.id }),
);
