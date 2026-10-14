/* global afterEach, beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, cleanup, renderHook } from "@testing-library/react";
import { useStrategyMapCanvasViewport } from "./use-strategy-map-canvas-viewport";

const originalResizeObserver = globalThis.ResizeObserver;
const resizeObserver = { disconnect: jest.fn(), observe: jest.fn() };
const frames = new Map<number, FrameRequestCallback>();
let nextFrameId = 0;
let element: HTMLDivElement;
let scrollWidth = 5000;
let scrollHeight = 4000;
let roundScrollPositions = false;

const flushFrame = () => {
  act(() => {
    const callbacks = [...frames.values()];
    frames.clear();
    callbacks.forEach((callback) => {
      callback(16);
    });
  });
};

const wheel = (options: WheelEventInit = {}) => {
  const event = new WheelEvent("wheel", {
    bubbles: true,
    cancelable: true,
    clientX: 400,
    clientY: 250,
    ctrlKey: true,
    deltaY: -0.25,
    ...options,
  });
  act(() => {
    element.dispatchEvent(event);
  });
  return event;
};

const scrollTo = (left: number, top: number) => {
  act(() => {
    element.scrollLeft = left;
    element.scrollTop = top;
    element.dispatchEvent(new Event("scroll"));
  });
};

const notifyScroll = () => {
  act(() => {
    element.dispatchEvent(new Event("scroll"));
  });
};

const renderViewport = (initialZoom = 1) => {
  const viewportRef = { current: element };
  const onZoomChange = jest.fn();
  const hook = renderHook(
    ({ resetSignal, zoom }) =>
      useStrategyMapCanvasViewport({
        layoutWidth: 5000,
        onZoomChange,
        resetSignal,
        viewportRef,
        zoom,
      }),
    { initialProps: { resetSignal: 0, zoom: initialZoom } },
  );
  return { ...hook, onZoomChange };
};

beforeEach(() => {
  jest.clearAllMocks();
  frames.clear();
  nextFrameId = 0;
  scrollWidth = 5000;
  scrollHeight = 4000;
  roundScrollPositions = false;
  let scrollLeft = 0;
  let scrollTop = 0;
  element = document.createElement("div");
  Object.defineProperties(element, {
    clientWidth: { configurable: true, value: 1000 },
    clientHeight: { configurable: true, value: 600 },
    scrollWidth: { configurable: true, get: () => scrollWidth },
    scrollHeight: { configurable: true, get: () => scrollHeight },
    scrollLeft: {
      configurable: true,
      get: () => scrollLeft,
      set: (value: number) => {
        const nextValue = roundScrollPositions ? Math.round(value) : value;
        scrollLeft = Math.max(0, Math.min(nextValue, scrollWidth - 1000));
      },
    },
    scrollTop: {
      configurable: true,
      get: () => scrollTop,
      set: (value: number) => {
        const nextValue = roundScrollPositions ? Math.round(value) : value;
        scrollTop = Math.max(0, Math.min(nextValue, scrollHeight - 600));
      },
    },
  });
  jest.spyOn(element, "getBoundingClientRect").mockReturnValue({
    bottom: 650,
    height: 600,
    left: 100,
    right: 1100,
    top: 50,
    width: 1000,
    x: 100,
    y: 50,
    toJSON: () => ({}),
  });
  jest.spyOn(window, "requestAnimationFrame").mockImplementation((callback) => {
    const id = ++nextFrameId;
    frames.set(id, callback);
    return id;
  });
  jest.spyOn(window, "cancelAnimationFrame").mockImplementation((id) => {
    frames.delete(id);
  });
  globalThis.ResizeObserver = jest.fn(
    () => resizeObserver,
  ) as unknown as typeof ResizeObserver;
  document.body.append(element);
});

afterEach(() => {
  cleanup();
  element.remove();
  globalThis.ResizeObserver = originalResizeObserver;
  jest.restoreAllMocks();
});

it.each(["ctrlKey", "metaKey"] as const)(
  "accumulates twenty fine %s wheel events and publishes one continuous zoom per frame",
  (modifier) => {
    const { onZoomChange } = renderViewport();
    for (let index = 0; index < 20; index++) {
      expect(wheel({ ctrlKey: false, [modifier]: true }).defaultPrevented).toBe(
        true,
      );
    }

    expect(onZoomChange).not.toHaveBeenCalled();
    expect(frames.size).toBe(1);
    flushFrame();
    expect(onZoomChange).toHaveBeenCalledTimes(1);
    expect(onZoomChange.mock.calls[0][0]).toBeCloseTo(
      Math.exp(20 * 0.25 * 0.002),
      12,
    );
  },
);

it("keeps the world point under the pointer stable through wheel zoom in and out", () => {
  const { onZoomChange, rerender } = renderViewport();
  scrollTo(1800, 1000);
  flushFrame();
  const worldX = element.scrollLeft + 300;
  const worldY = element.scrollTop + 200;

  wheel({ deltaY: -100 });
  flushFrame();
  const zoomIn = onZoomChange.mock.calls[0][0] as number;
  scrollWidth = 5000 * zoomIn;
  scrollHeight = 4000 * zoomIn;
  rerender({ resetSignal: 0, zoom: zoomIn });
  expect((element.scrollLeft + 300) / zoomIn).toBeCloseTo(worldX, 10);
  expect((element.scrollTop + 200) / zoomIn).toBeCloseTo(worldY, 10);

  wheel({ deltaY: 75 });
  flushFrame();
  const zoomOut = onZoomChange.mock.calls[1][0] as number;
  scrollWidth = 5000 * zoomOut;
  scrollHeight = 4000 * zoomOut;
  rerender({ resetSignal: 0, zoom: zoomOut });
  expect((element.scrollLeft + 300) / zoomOut).toBeCloseTo(worldX, 10);
  expect((element.scrollTop + 200) / zoomOut).toBeCloseTo(worldY, 10);
});

it("uses the pre-resize center when zooming out after the browser clamps native scroll", () => {
  const { rerender } = renderViewport();
  scrollTo(3700, 3000);
  // The scroll listener captures this camera before its deferred render window
  // update. The browser then clamps scroll as React shrinks the spacer.
  scrollWidth = 4000;
  scrollHeight = 3200;
  element.scrollLeft = 3700;
  element.scrollTop = 3000;
  expect(element.scrollLeft).toBe(3000);
  expect(element.scrollTop).toBe(2600);

  rerender({ resetSignal: 0, zoom: 0.8 });
  expect(element.scrollLeft).toBeCloseTo((3700 + 500) * 0.8 - 500, 10);
  expect(element.scrollTop).toBeCloseTo((3000 + 300) * 0.8 - 300, 10);
});

it.each([false, true])(
  "does not accumulate integer scroll rounding drift across sixty fine zoom frames (pan halfway: %s)",
  (panHalfway) => {
    roundScrollPositions = true;
    const { onZoomChange, rerender } = renderViewport();
    scrollTo(1800, 1000);
    flushFrame();
    let zoom = 1;
    let expectedWorldX = element.scrollLeft + 300;
    let expectedWorldY = element.scrollTop + 200;

    for (let frame = 0; frame < 60; frame++) {
      if (panHalfway && frame === 30) {
        const previousLeft = element.scrollLeft;
        const previousTop = element.scrollTop;
        scrollTo(previousLeft + 140, previousTop - 80);
        flushFrame();
        expectedWorldX += (element.scrollLeft - previousLeft) / zoom;
        expectedWorldY += (element.scrollTop - previousTop) / zoom;
      }
      for (let burst = 0; burst < 4; burst++) {
        wheel({ deltaY: frame < 30 ? 1 : -1 });
      }
      flushFrame();
      zoom = onZoomChange.mock.calls[frame][0] as number;
      scrollWidth = 5000 * zoom;
      scrollHeight = 4000 * zoom;
      rerender({ resetSignal: 0, zoom });
      // Native scroll events arrive after imperative zoom positioning too.
      // They must preserve its fractional remainder rather than resetting it.
      notifyScroll();
      flushFrame();

      expect(
        Math.abs((element.scrollLeft + 300) / zoom - expectedWorldX),
      ).toBeLessThan(1 / zoom);
      expect(
        Math.abs((element.scrollTop + 200) / zoom - expectedWorldY),
      ).toBeLessThan(1 / zoom);
    }

    expect(onZoomChange).toHaveBeenCalledTimes(60);
    expect(zoom).toBeCloseTo(1, 10);
  },
);

it("resets from a non-default zoom exactly once and cancels pending wheel input", () => {
  scrollWidth = 7500;
  scrollHeight = 6000;
  const { onZoomChange, rerender } = renderViewport(1.5);
  scrollTo(4000, 2000);
  wheel();
  scrollWidth = 5000;
  scrollHeight = 4000;
  rerender({ resetSignal: 1, zoom: 1 });

  expect(element.scrollLeft).toBe(2000);
  expect(element.scrollTop).toBe(0);
  flushFrame();
  expect(onZoomChange).not.toHaveBeenCalled();
  expect(element.scrollLeft).toBe(2000);
  expect(element.scrollTop).toBe(0);
});

it("retains render window identity through small native scrolls and replenishes before crossing its guard", () => {
  const { result } = renderViewport();
  const originalViewport = result.current;

  for (let index = 1; index <= 20; index++) {
    scrollTo(2000 + index * 5, index * 4);
    flushFrame();
    expect(result.current).toBe(originalViewport);
  }

  scrollTo(3800, 2600);
  flushFrame();
  expect(result.current).not.toBe(originalViewport);
  expect(result.current.left).toBeLessThan(element.scrollLeft);
  expect(result.current.top).toBeLessThan(element.scrollTop);
  expect(result.current.left + result.current.width).toBeGreaterThan(
    element.scrollLeft + element.clientWidth,
  );
  expect(result.current.top + result.current.height).toBeGreaterThan(
    element.scrollTop + element.clientHeight,
  );
});

it("keeps scroll and wheel listeners and the resize observer stable across zoom changes", () => {
  const addEventListener = jest.spyOn(element, "addEventListener");
  const removeEventListener = jest.spyOn(element, "removeEventListener");
  const { rerender } = renderViewport();
  [1.1, 1.2, 0.9, 1].forEach((zoom) => {
    rerender({ resetSignal: 0, zoom });
  });

  expect(addEventListener.mock.calls.map(([type]) => type)).toEqual([
    "scroll",
    "wheel",
  ]);
  expect(removeEventListener).not.toHaveBeenCalled();
  expect(resizeObserver.observe).toHaveBeenCalledTimes(1);
  expect(resizeObserver.disconnect).not.toHaveBeenCalled();
});

it("ignores ordinary wheel scrolling and cancels pending work on unmount", () => {
  const removeEventListener = jest.spyOn(element, "removeEventListener");
  const { onZoomChange, unmount } = renderViewport();
  expect(wheel({ ctrlKey: false }).defaultPrevented).toBe(false);
  expect(frames.size).toBe(0);
  wheel();
  scrollTo(2200, 100);
  expect(frames.size).toBe(2);
  unmount();

  expect(frames.size).toBe(0);
  expect(window.cancelAnimationFrame).toHaveBeenCalledTimes(2);
  expect(resizeObserver.disconnect).toHaveBeenCalledTimes(1);
  expect(removeEventListener.mock.calls.map(([type]) => type)).toEqual([
    "scroll",
    "wheel",
  ]);
  flushFrame();
  expect(onZoomChange).not.toHaveBeenCalled();
});
