import { useState, useLayoutEffect } from "react";
import type { Objective, KeyResult } from "@/modules/objectives/types";
import type { StrategyMap } from "@/shared/strategy-map/types";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrategyMapCanvas } from "@strategy-fixture/canvas";
import { createStrategyMapLayout } from "@strategy-fixture/layout";
import { objectiveKeys } from "@/modules/objectives/constants";
declare const FIXTURE_MODE: string;
declare global {
  interface Window {
    fixtureKeyResults: ReadonlyMap<string, KeyResult[]>;
    webkit?: {
      messageHandlers: { result: { postMessage: (data: unknown) => void } };
    };
  }
}

const objectives = Array.from({ length: 500 }, (_, index) => ({
  id: `objective-${index}`,
  name: `Objective ${index + 1}: improve customer outcomes`,
  keyResultCount: 2,
  stats: { completed: 3, total: 8 },
  description: "Synthetic fixture objective",
})) as unknown as Objective[];
const keyResults = new Map(
  objectives.map(({ id }, i) => [
    id,
    [0, 1].map((j) => ({
      id: `${id}-kr-${j}`,
      objectiveId: id,
      name: `Key result ${i + 1}.${j + 1}`,
      startValue: 0,
      currentValue: 30,
      targetValue: 100,
    })) as unknown as KeyResult[],
  ]),
);
window.fixtureKeyResults = keyResults;
const strategy = {
  ultimateGoal: "Build an excellent product",
  description: "Synthetic isolated 500-objective map",
  pillars: Array.from({ length: 5 }, (_, i) => ({
    id: `pillar-${i}`,
    name: `Pillar ${i + 1}`,
    description: "Synthetic pillar",
    orderIndex: i,
    objectiveIds: objectives.slice(i * 100, (i + 1) * 100).map((o) => o.id),
  })),
};
const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false, gcTime: Infinity } },
});
objectives.forEach(({ id }) =>
  queryClient.setQueryData(
    objectiveKeys.keyResults("motion-fixture", id),
    keyResults.get(id),
  ),
);
const layout = createStrategyMapLayout(
  strategy as StrategyMap,
  objectives,
  keyResults,
  new Set(objectives.map((o) => o.id)),
);
const noAction = () => {};
const report = (type: string, data: Record<string, unknown>) => {
  window.webkit?.messageHandlers.result.postMessage({
    type,
    mode: FIXTURE_MODE,
    focused: document.hasFocus(),
    visible: document.visibilityState,
    ...data,
  });
  console.log(type, JSON.stringify(data));
};
const frames = (count = 1) =>
  new Promise<void>((resolve) => {
    const next = () => (count-- > 0 ? requestAnimationFrame(next) : resolve());
    next();
  });
const wait = (ms: number) => new Promise((r) => setTimeout(r, ms));
let zoomValue = 1;
let commitCount = 0;
const App = () => {
  const [zoom, setZoom] = useState(1);
  useLayoutEffect(() => {
    zoomValue = zoom;
    commitCount++;
  });
  return (
    <QueryClientProvider client={queryClient}>
      <div className="fixture-host">
        <StrategyMapCanvas
          canEdit
          objectives={objectives}
          strategy={strategy}
          showUnaligned
          zoom={zoom}
          onZoomChange={setZoom}
          onAddPillar={noAction}
          onAlign={noAction}
          onDeletePillar={noAction}
          onSelectGoal={noAction}
          onSelectKeyResult={noAction}
          onSelectObjective={noAction}
          onSelectPillar={noAction}
        />
      </div>
      <div className="fixture-status">
        {FIXTURE_MODE} · 500 objectives · {Math.round(zoom * 100)}%
      </div>
    </QueryClientProvider>
  );
};
createRoot(document.getElementById("root")!).render(<App />);
let collecting = false;
let gaps: number[] = [];
let last: number | null = null;
const monitor = (time: number) => {
  if (collecting && last !== null) gaps.push(time - last);
  last = time;
  requestAnimationFrame(monitor);
};
requestAnimationFrame(monitor);
const quantile = (values: number[], fraction: number) =>
  [...values].sort((a, b) => a - b)[
    Math.min(values.length - 1, Math.floor(values.length * fraction))
  ] ?? 0;
const frameStats = () => ({
  frames: gaps.length,
  p50: quantile(gaps, 0.5),
  p95: quantile(gaps, 0.95),
  max: Math.max(...gaps),
  over50: gaps.filter((x) => x > 50).length,
});
const dataNodes = () =>
  Array.from(document.querySelectorAll<HTMLElement>("[data-node-id]"));
// Check logical SVG endpoint continuity independently of frame timing. A source
// anchor may briefly use default dimensions before its first ResizeObserver;
// target dots should always meet the visible top-center of the target card.
const endpointStats = (viewport: HTMLElement) => {
  const viewportBounds = viewport.getBoundingClientRect();
  const dots = new Set(
    Array.from(
      viewport.querySelectorAll("svg circle"),
      (dot) => `${dot.getAttribute("cx")}:${dot.getAttribute("cy")}`,
    ),
  );
  const sourceAnchors = new Set(
    Array.from(viewport.querySelectorAll("svg path"), (path) => {
      const match = path.getAttribute("d")?.match(/^M ([^ ]+) ([^ ]+)/);
      return match ? `${match[1]}:${match[2]}` : "";
    }),
  );
  let missingTargetDots = 0;
  let misalignedSourceAnchors = 0;
  let targetChecks = 0;
  let sourceChecks = 0;
  for (const node of dataNodes()) {
    const bounds = node.getBoundingClientRect();
    const center = (bounds.left + bounds.right) / 2;
    if (center < viewportBounds.left + 8 || center > viewportBounds.right - 8)
      continue;
    const sourceX = node.offsetLeft + node.offsetWidth / 2;
    if (
      node.dataset.nodeId !== "goal" &&
      bounds.top > viewportBounds.top + 8 &&
      bounds.top < viewportBounds.bottom - 8
    ) {
      targetChecks++;
      if (!dots.has(`${sourceX}:${node.offsetTop}`)) missingTargetDots++;
    }
    if (
      !node.dataset.nodeId?.startsWith("key-result:") &&
      bounds.bottom > viewportBounds.top + 8 &&
      bounds.bottom < viewportBounds.bottom - 8
    ) {
      sourceChecks++;
      if (
        !sourceAnchors.has(`${sourceX}:${node.offsetTop + node.offsetHeight}`)
      )
        misalignedSourceAnchors++;
    }
  }
  return {
    missingTargetDots,
    misalignedSourceAnchors,
    targetChecks,
    sourceChecks,
  };
};
async function run() {
  await wait(1000);
  const viewport = document.querySelector<HTMLElement>(".overflow-auto")!;
  if (!viewport) throw Error("Viewport missing");
  const startX = layout.positions["objective:objective-0"].x - 70;
  viewport.scrollLeft = startX;
  viewport.scrollTop = 600;
  await frames(4);
  await wait(200);
  report("ready", {
    userAgent: navigator.userAgent,
    focused: document.hasFocus(),
    visible: document.visibilityState,
    layout: { width: layout.width, height: layout.height },
    viewport: { width: viewport.clientWidth, height: viewport.clientHeight },
  });
  let removedNodes = 0,
    addedNodes = 0,
    pathChanges = 0,
    addedPaths = 0,
    removedPaths = 0;
  const observer = new MutationObserver((records) => {
    if (!collecting) return;
    for (const r of records) {
      if (r.type === "attributes" && r.attributeName === "d") pathChanges++;
      for (const n of r.addedNodes) {
        if (n instanceof Element) {
          addedPaths +=
            Number(n.tagName.toLowerCase() === "path") +
            n.querySelectorAll("path").length;
        }
        if (n instanceof HTMLElement) {
          if (n.dataset.nodeId) addedNodes++;
          addedNodes += n.querySelectorAll("[data-node-id]").length;
        }
      }
      for (const n of r.removedNodes) {
        if (n instanceof Element) {
          removedPaths +=
            Number(n.tagName.toLowerCase() === "path") +
            n.querySelectorAll("path").length;
        }
        if (n instanceof HTMLElement) {
          if (n.dataset.nodeId) removedNodes++;
          removedNodes += n.querySelectorAll("[data-node-id]").length;
        }
      }
    }
  });
  observer.observe(viewport, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: ["d"],
  });
  const samples: Record<string, number>[] = [];
  let missingFrames = 0,
    totalMissing = 0;
  const phaseStart = performance.now();
  gaps = [];
  collecting = true;
  for (let step = 0; step < 100; step++) {
    viewport.scrollTop = 600 + (step < 50 ? step : 99 - step) * 220;
    viewport.scrollLeft = startX + Math.sin(step / 12) * 80;
    await frames(1);
    const actual = new Set(dataNodes().map((n) => n.dataset.nodeId));
    const bounds = {
      left: viewport.scrollLeft / zoomValue,
      top: viewport.scrollTop / zoomValue,
      right: (viewport.scrollLeft + viewport.clientWidth) / zoomValue,
      bottom: (viewport.scrollTop + viewport.clientHeight) / zoomValue,
    };
    const missing = objectives.filter((o) => {
      const p = layout.positions["objective:" + o.id];
      const height = 218 + (Number(o.id.split("-")[1]) % 3) * 19;
      return (
        p.x + 340 > bounds.left + 8 &&
        p.x < bounds.right - 8 &&
        p.y + height > bounds.top + 8 &&
        p.y < bounds.bottom - 8 &&
        !actual.has("objective:" + o.id)
      );
    }).length;
    if (missing) {
      missingFrames++;
      totalMissing += missing;
    }
    samples.push({
      step,
      top: viewport.scrollTop,
      nodes: actual.size,
      paths: viewport.querySelectorAll("svg path").length,
      missing,
      ...endpointStats(viewport),
    });
  }
  collecting = false;
  report("phase", {
    name: "fast-pan",
    elapsed: performance.now() - phaseStart,
    missingFrames,
    totalMissing,
    removedNodes,
    addedNodes,
    pathChanges,
    addedPaths,
    removedPaths,
    ...frameStats(),
    samples,
  });
  viewport.scrollLeft = startX;
  viewport.scrollTop = 2000;
  await frames(4);
  gaps = [];
  collecting = true;
  const zStart = zoomValue;
  const zooms = [zoomValue];
  const worldXs: number[] = [];
  const worldYs: number[] = [];
  const zoomStart = performance.now();
  const box = viewport.getBoundingClientRect();
  const x = 500,
    y = 350;
  const beforeCommits = commitCount;
  for (let step = 0; step < 60; step++) {
    for (let burst = 0; burst < 4; burst++)
      viewport.dispatchEvent(
        new WheelEvent("wheel", {
          bubbles: true,
          cancelable: true,
          ctrlKey: true,
          clientX: box.left + x,
          clientY: box.top + y,
          deltaY: step < 30 ? 1 : -1,
        }),
      );
    await frames(1);
    zooms.push(zoomValue);
    worldXs.push((viewport.scrollLeft + x) / zoomValue);
    worldYs.push((viewport.scrollTop + y) / zoomValue);
  }
  await frames(2);
  collecting = false;
  report("phase", {
    name: "fine-pinch",
    elapsed: performance.now() - zoomStart,
    startZoom: zStart,
    endZoom: zoomValue,
    minZoom: Math.min(...zooms),
    uniqueZooms: new Set(zooms).size,
    anchorRangeX: Math.max(...worldXs) - Math.min(...worldXs),
    anchorRangeY: Math.max(...worldYs) - Math.min(...worldYs),
    commits: commitCount - beforeCommits,
    worldXs,
    worldYs,
    zooms,
    ...frameStats(),
  });
  gaps = [];
  collecting = true;
  const jumpStart = performance.now();
  let jumpMissingFrames = 0,
    jumpMissingTotal = 0;
  removedNodes = 0;
  addedNodes = 0;
  removedPaths = 0;
  addedPaths = 0;
  const jumpSamples: Record<string, number>[] = [];
  for (let step = 0; step < 50; step++) {
    viewport.scrollTop = 1000 + ((step * 620) % 9000);
    viewport.scrollLeft = startX;
    await frames(1);
    const actual = new Set(dataNodes().map((n) => n.dataset.nodeId));
    const top = viewport.scrollTop / zoomValue,
      bottom = (viewport.scrollTop + viewport.clientHeight) / zoomValue,
      left = viewport.scrollLeft / zoomValue,
      right = (viewport.scrollLeft + viewport.clientWidth) / zoomValue;
    const missing = objectives.filter((o) => {
      const p = layout.positions["objective:" + o.id];
      return (
        p.x + 340 > left + 8 &&
        p.x < right - 8 &&
        p.y + 218 > top + 8 &&
        p.y < bottom - 8 &&
        !actual.has("objective:" + o.id)
      );
    }).length;
    if (missing) {
      jumpMissingFrames++;
      jumpMissingTotal += missing;
    }
    jumpSamples.push({
      step,
      missing,
      nodes: actual.size,
      paths: viewport.querySelectorAll("svg path").length,
      ...endpointStats(viewport),
    });
  }
  collecting = false;
  report("phase", {
    name: "rapid-pan-reversal",
    elapsed: performance.now() - jumpStart,
    missingFrames: jumpMissingFrames,
    totalMissing: jumpMissingTotal,
    removedNodes,
    addedNodes,
    removedPaths,
    addedPaths,
    ...frameStats(),
    samples: jumpSamples,
  });
  observer.disconnect();
  report("complete", {
    focused: document.hasFocus(),
    visible: document.visibilityState,
  });
}
run().catch((error) =>
  report("fatal", { error: String(error), stack: error.stack }),
);
