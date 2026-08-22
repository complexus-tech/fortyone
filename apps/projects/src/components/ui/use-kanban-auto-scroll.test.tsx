/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import { useDndMonitor } from "@dnd-kit/core";
import { act, cleanup, renderHook } from "@testing-library/react";
import { useKanbanAutoScroll } from "./use-kanban-auto-scroll";

jest.mock("@dnd-kit/core", () => ({
  useDndMonitor: jest.fn(),
}));

const mockedUseDndMonitor = jest.mocked(useDndMonitor);
type MonitorListener = Parameters<typeof useDndMonitor>[0];

let frames: Map<number, FrameRequestCallback>;
let monitor: MonitorListener;
let nextFrameId: number;

const dispatchPointerEvent = (type: string, clientX: number) => {
  act(() => {
    window.dispatchEvent(new MouseEvent(type, { clientX }));
  });
};

const dispatchTouchEvent = (type: string, clientX: number) => {
  const event = new Event(type);
  Object.defineProperty(event, "touches", {
    value: {
      item: () => ({ clientX }),
    },
  });
  act(() => {
    window.dispatchEvent(event);
  });
};

const flushNextFrame = (timestamp: number) => {
  const nextFrame = frames.entries().next().value;
  if (!nextFrame) throw new Error("Expected a scheduled animation frame");

  const [frameId, callback] = nextFrame;
  frames.delete(frameId);
  act(() => {
    callback(timestamp);
  });
};

const createScrollContainer = () => {
  const element = document.createElement("div");
  Object.defineProperties(element, {
    clientWidth: { configurable: true, value: 800 },
    scrollWidth: { configurable: true, value: 1_600 },
  });
  element.scrollLeft = 400;
  jest.spyOn(element, "getBoundingClientRect").mockReturnValue({
    bottom: 600,
    height: 500,
    left: 100,
    right: 900,
    top: 100,
    width: 800,
    x: 100,
    y: 100,
    toJSON: () => ({}),
  });
  return element;
};

const startDrag = (clientX: number) => {
  dispatchPointerEvent("pointerdown", clientX);
  act(() => {
    monitor.onDragStart?.({ active: { id: "story-1" } } as DragStartEvent);
  });
};

describe("useKanbanAutoScroll", () => {
  beforeEach(() => {
    frames = new Map();
    nextFrameId = 1;
    mockedUseDndMonitor.mockImplementation((listener) => {
      monitor = listener;
    });
    Object.defineProperty(window, "requestAnimationFrame", {
      configurable: true,
      value: jest.fn((callback: FrameRequestCallback) => {
        const frameId = nextFrameId;
        nextFrameId += 1;
        frames.set(frameId, callback);
        return frameId;
      }),
    });
    Object.defineProperty(window, "cancelAnimationFrame", {
      configurable: true,
      value: jest.fn((frameId: number) => {
        frames.delete(frameId);
      }),
    });
  });

  afterEach(() => {
    cleanup();
    mockedUseDndMonitor.mockReset();
    jest.restoreAllMocks();
  });

  it("reverses direction on the next frame without a React state update", () => {
    const scrollContainer = createScrollContainer();
    renderHook(() => {
      useKanbanAutoScroll({ current: scrollContainer });
    });

    startDrag(900);
    flushNextFrame(0);
    flushNextFrame(16);
    expect(scrollContainer.scrollLeft).toBeCloseTo(407.68);

    dispatchPointerEvent("pointermove", 100);
    flushNextFrame(32);
    expect(scrollContainer.scrollLeft).toBeCloseTo(400);
  });

  it("caps long frames and parks the loop outside an edge zone", () => {
    const scrollContainer = createScrollContainer();
    renderHook(() => {
      useKanbanAutoScroll({ current: scrollContainer });
    });

    startDrag(900);
    flushNextFrame(0);
    flushNextFrame(1_000);
    expect(scrollContainer.scrollLeft).toBeCloseTo(415.36);

    dispatchPointerEvent("pointermove", 500);
    flushNextFrame(1_016);
    expect(frames.size).toBe(0);

    dispatchPointerEvent("pointermove", 100);
    expect(frames.size).toBe(1);
    flushNextFrame(1_032);
    flushNextFrame(1_048);
    expect(scrollContainer.scrollLeft).toBeLessThan(415.36);
  });

  it("tracks touch input without relying on the drag-start event shape", () => {
    const scrollContainer = createScrollContainer();
    renderHook(() => {
      useKanbanAutoScroll({ current: scrollContainer });
    });

    dispatchTouchEvent("touchstart", 900);
    act(() => {
      monitor.onDragStart?.({ active: { id: "story-1" } } as DragStartEvent);
    });
    flushNextFrame(0);
    flushNextFrame(16);
    expect(scrollContainer.scrollLeft).toBeGreaterThan(400);

    dispatchTouchEvent("touchmove", 100);
    flushNextFrame(32);
    expect(scrollContainer.scrollLeft).toBeCloseTo(400);
  });

  it("cancels the pending frame when a drag ends or the hook unmounts", () => {
    const scrollContainer = createScrollContainer();
    const { unmount } = renderHook(() => {
      useKanbanAutoScroll({ current: scrollContainer });
    });

    startDrag(900);
    expect(frames.size).toBe(1);
    act(() => {
      monitor.onDragEnd?.({} as DragEndEvent);
    });
    expect(frames.size).toBe(0);

    startDrag(900);
    expect(frames.size).toBe(1);
    unmount();
    expect(frames.size).toBe(0);
  });
});
