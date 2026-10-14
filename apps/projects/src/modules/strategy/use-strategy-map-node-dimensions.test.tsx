/* global afterEach, beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, cleanup, renderHook } from "@testing-library/react";
import type { StrategyNodeDimensions } from "./strategy-map-layout";
import { useStrategyMapCanvasNodeDimensions } from "./use-strategy-map-node-dimensions";

const originalResizeObserver = globalThis.ResizeObserver;
const originalMutationObserver = globalThis.MutationObserver;
const resizeObserver = {
  disconnect: jest.fn(),
  observe: jest.fn(),
  unobserve: jest.fn(),
};
const mutationObserver = {
  disconnect: jest.fn(),
  observe: jest.fn(),
  takeRecords: jest.fn(),
};
let resizeCallback: ResizeObserverCallback;
let mutationCallback: MutationCallback;
let canvas: HTMLDivElement;

const createNode = (id: string, width = 340, height = 174) => {
  const node = document.createElement("div");
  node.dataset.nodeId = id;
  const readWidth = jest.fn(() => width);
  const readHeight = jest.fn(() => height);
  Object.defineProperties(node, {
    offsetWidth: { configurable: true, get: readWidth },
    offsetHeight: { configurable: true, get: readHeight },
  });
  return { node, readHeight, readWidth };
};

const notifyMutation = (addedNodes: Node[] = [], removedNodes: Node[] = []) => {
  act(() => {
    mutationCallback(
      [{ addedNodes, removedNodes } as unknown as MutationRecord],
      mutationObserver as unknown as MutationObserver,
    );
  });
};

const notifyResize = (
  ...entries: { target: HTMLElement; borderBoxSize?: unknown }[]
) => {
  act(() => {
    resizeCallback(
      // This fixture also exercises the legacy non-array borderBoxSize shape.
      entries as unknown as ResizeObserverEntry[],
      resizeObserver as unknown as ResizeObserver,
    );
  });
};

const renderDimensions = () => {
  const canvasRef = { current: canvas };
  const dimensionsRef = {
    current: {} as Record<string, StrategyNodeDimensions>,
  };
  let renderCount = 0;
  const hook = renderHook(() => {
    renderCount++;
    return useStrategyMapCanvasNodeDimensions({ canvasRef, dimensionsRef });
  });
  return { ...hook, dimensionsRef, getRenderCount: () => renderCount };
};

beforeEach(() => {
  jest.clearAllMocks();
  globalThis.ResizeObserver = jest.fn((callback: ResizeObserverCallback) => {
    resizeCallback = callback;
    return resizeObserver;
  }) as unknown as typeof ResizeObserver;
  globalThis.MutationObserver = jest.fn((callback: MutationCallback) => {
    mutationCallback = callback;
    return mutationObserver;
  }) as unknown as typeof MutationObserver;
  canvas = document.createElement("div");
  document.body.append(canvas);
});

afterEach(() => {
  cleanup();
  canvas.remove();
  globalThis.ResizeObserver = originalResizeObserver;
  globalThis.MutationObserver = originalMutationObserver;
  jest.restoreAllMocks();
});

it("does no canvas scan or card measurement when a descendant changes in a 500-card map", () => {
  const nodes = Array.from({ length: 500 }, (_, index) =>
    createNode(`objective:${index}`),
  );
  canvas.append(...nodes.map(({ node }) => node));
  const scanCanvas = jest.spyOn(canvas, "querySelectorAll");
  const { result, getRenderCount } = renderDimensions();
  const originalDimensions = result.current;
  const originalRenderCount = getRenderCount();
  expect(scanCanvas).toHaveBeenCalledTimes(1);
  expect(Object.keys(result.current)).toHaveLength(500);
  nodes.forEach(({ readHeight, readWidth }) => {
    readHeight.mockClear();
    readWidth.mockClear();
  });

  const label = document.createElement("span");
  nodes[0].node.append(label);
  notifyMutation([label]);
  label.textContent = "Updated title";
  notifyMutation([label.firstChild!]);

  expect(scanCanvas).toHaveBeenCalledTimes(1);
  expect(result.current).toBe(originalDimensions);
  expect(getRenderCount()).toBe(originalRenderCount);
  nodes.forEach(({ readHeight, readWidth }) => {
    expect(readHeight).not.toHaveBeenCalled();
    expect(readWidth).not.toHaveBeenCalled();
  });
});

it("measures only added cards and retains removed card dimensions for connector geometry", () => {
  const existing = createNode("objective:existing");
  canvas.append(existing.node);
  const { result } = renderDimensions();
  existing.readHeight.mockClear();
  existing.readWidth.mockClear();
  const wrapper = document.createElement("div");
  const added = createNode("objective:added", 280, 160);
  wrapper.append(added.node);
  canvas.append(wrapper);
  notifyMutation([wrapper]);

  expect(result.current["objective:added"]).toEqual({
    height: 160,
    width: 280,
  });
  expect(resizeObserver.observe).toHaveBeenLastCalledWith(added.node, {
    box: "border-box",
  });
  expect(added.readHeight).toHaveBeenCalledTimes(1);
  expect(existing.readHeight).not.toHaveBeenCalled();
  expect(existing.readWidth).not.toHaveBeenCalled();
  const beforeRemoval = result.current;

  wrapper.remove();
  notifyMutation([], [wrapper]);
  expect(resizeObserver.unobserve).toHaveBeenCalledWith(added.node);
  expect(result.current).toBe(beforeRemoval);
  expect(result.current["objective:added"]).toEqual({
    height: 160,
    width: 280,
  });
});

it("does not remeasure or resubscribe a card moved within one mutation batch", () => {
  const { node, readHeight, readWidth } = createNode("objective:moved");
  canvas.append(node);
  const { result } = renderDimensions();
  const originalDimensions = result.current;
  readHeight.mockClear();
  readWidth.mockClear();
  resizeObserver.observe.mockClear();
  const wrapper = document.createElement("div");
  canvas.append(wrapper);
  wrapper.append(node);
  notifyMutation([wrapper, node], [node]);

  expect(result.current).toBe(originalDimensions);
  expect(readHeight).not.toHaveBeenCalled();
  expect(readWidth).not.toHaveBeenCalled();
  expect(resizeObserver.observe).not.toHaveBeenCalled();
  expect(resizeObserver.unobserve).not.toHaveBeenCalled();
});

it("retains precise cached geometry when a culled card remounts instead of rounding its endpoints", () => {
  const original = createNode("objective:resident");
  canvas.append(original.node);
  const { result, getRenderCount } = renderDimensions();
  notifyResize({
    target: original.node,
    borderBoxSize: [{ blockSize: 174.25, inlineSize: 340.5 }],
  });
  const preciseDimensions = result.current;
  const renderCount = getRenderCount();
  original.node.remove();
  notifyMutation([], [original.node]);

  const remounted = createNode("objective:resident");
  canvas.append(remounted.node);
  notifyMutation([remounted.node]);
  expect(resizeObserver.observe).toHaveBeenLastCalledWith(remounted.node, {
    box: "border-box",
  });
  expect(remounted.readHeight).not.toHaveBeenCalled();
  expect(remounted.readWidth).not.toHaveBeenCalled();
  expect(result.current).toBe(preciseDimensions);
  expect(getRenderCount()).toBe(renderCount);

  notifyResize({
    target: remounted.node,
    borderBoxSize: [{ blockSize: 174.25, inlineSize: 340.5 }],
  });
  expect(result.current).toBe(preciseDimensions);
  expect(getRenderCount()).toBe(renderCount);
});

it("publishes precise border-box resize dimensions without layout reads and batches updates", () => {
  const first = createNode("objective:first");
  const second = createNode("objective:second");
  canvas.append(first.node, second.node);
  const { result, dimensionsRef, getRenderCount } = renderDimensions();
  const initialRenderCount = getRenderCount();
  first.readHeight.mockClear();
  first.readWidth.mockClear();
  second.readHeight.mockClear();
  second.readWidth.mockClear();

  notifyResize(
    {
      target: first.node,
      borderBoxSize: [{ blockSize: 180.25, inlineSize: 340.5 }],
    },
    {
      target: second.node,
      borderBoxSize: [{ blockSize: 190, inlineSize: 350 }],
    },
  );
  expect(result.current["objective:first"]).toEqual({
    height: 180.25,
    width: 340.5,
  });
  expect(result.current["objective:second"]).toEqual({
    height: 190,
    width: 350,
  });
  expect(dimensionsRef.current).toBe(result.current);
  expect(getRenderCount()).toBe(initialRenderCount + 1);
  [first, second].forEach(({ readHeight, readWidth }) => {
    expect(readHeight).not.toHaveBeenCalled();
    expect(readWidth).not.toHaveBeenCalled();
  });

  const unchanged = result.current;
  notifyResize({
    target: first.node,
    borderBoxSize: [{ blockSize: 180.25, inlineSize: 340.5 }],
  });
  expect(result.current).toBe(unchanged);
  expect(getRenderCount()).toBe(initialRenderCount + 1);
});

it("supports the older WebKit box shape and measures only resized targets when box sizes are absent", () => {
  const first = createNode("objective:first");
  const second = createNode("objective:second");
  canvas.append(first.node, second.node);
  const { result } = renderDimensions();
  first.readHeight.mockClear();
  first.readWidth.mockClear();
  second.readHeight.mockClear();
  second.readWidth.mockClear();

  notifyResize({
    target: first.node,
    borderBoxSize: { blockSize: 181, inlineSize: 341 },
  });
  expect(result.current["objective:first"]).toEqual({
    height: 181,
    width: 341,
  });
  expect(first.readHeight).not.toHaveBeenCalled();
  first.readHeight.mockReturnValue(205);
  first.readWidth.mockReturnValue(350);
  notifyResize({ target: first.node });

  expect(result.current["objective:first"]).toEqual({
    height: 205,
    width: 350,
  });
  expect(first.readHeight).toHaveBeenCalledTimes(1);
  expect(first.readWidth).toHaveBeenCalledTimes(1);
  expect(second.readHeight).not.toHaveBeenCalled();
  expect(second.readWidth).not.toHaveBeenCalled();
});

it("ignores stale resize entries for removed cards and disconnects both observers on unmount", () => {
  const { node, readHeight } = createNode("objective:removed");
  canvas.append(node);
  const { dimensionsRef, unmount } = renderDimensions();
  const originalDimensions = dimensionsRef.current;
  node.remove();
  notifyMutation([], [node]);
  readHeight.mockClear();
  notifyResize({ target: node });
  expect(readHeight).not.toHaveBeenCalled();
  expect(dimensionsRef.current).toBe(originalDimensions);

  unmount();
  expect(resizeObserver.disconnect).toHaveBeenCalledTimes(1);
  expect(mutationObserver.disconnect).toHaveBeenCalledTimes(1);
  canvas.append(node);
  notifyMutation([node]);
  notifyResize({ target: node });
  expect(readHeight).not.toHaveBeenCalled();
  expect(dimensionsRef.current).toBe(originalDimensions);
});
