/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, cleanup, render, screen } from "@testing-library/react";
import { VirtuosoMockContext } from "react-virtuoso";
import { Timeline } from "./timeline";

type Message = { id: string; date: string };
const messages: Message[] = Array.from({ length: 60 }, (_, index) => ({
  id: `message-${index}`,
  date: index < 30 ? "2026-10-15" : "2026-10-16",
}));
const getDate = (message: Message) => message.date;
const measurements = { viewportHeight: 480, itemHeight: 96 };
const originalResizeObserver = Object.getOwnPropertyDescriptor(
  globalThis,
  "ResizeObserver",
);
const originalScrollTo = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  "scrollTo",
);
const timeline = (items: Message[], grouped = true) => (
  <VirtuosoMockContext.Provider value={measurements}>
    <Timeline
      getItemDate={grouped ? getDate : undefined}
      header={<p>Conversation introduction</p>}
      items={items}
      renderItem={(item) => <p>{item.id}</p>}
      startAtEnd={grouped}
    />
  </VirtuosoMockContext.Provider>
);

beforeEach(() => {
  jest.useFakeTimers();
  // Keep Virtuoso's real indexing and initial-position logic. JSDOM supplies no
  // layout, so the context provides item sizes and these values emulate a pane.
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: class {
      observe = jest.fn();
      unobserve = jest.fn();
      disconnect = jest.fn();
    },
  });
  jest.spyOn(HTMLElement.prototype, "offsetHeight", "get").mockReturnValue(480);
  jest.spyOn(HTMLElement.prototype, "clientHeight", "get").mockReturnValue(480);
  jest
    .spyOn(HTMLElement.prototype, "scrollHeight", "get")
    .mockReturnValue(6_000);
  jest.spyOn(HTMLElement.prototype, "getBoundingClientRect").mockReturnValue({
    x: 0,
    y: 0,
    left: 0,
    top: 0,
    right: 800,
    bottom: 480,
    width: 800,
    height: 480,
    toJSON: () => ({}),
  });
  Object.defineProperty(HTMLElement.prototype, "scrollTo", {
    configurable: true,
    value(this: HTMLElement, options: ScrollToOptions) {
      this.scrollTop = options.top ?? this.scrollTop;
      this.dispatchEvent(new Event("scroll"));
    },
  });
});

afterEach(() => {
  cleanup();
  jest.restoreAllMocks();
  if (originalResizeObserver)
    Object.defineProperty(globalThis, "ResizeObserver", originalResizeObserver);
  else Reflect.deleteProperty(globalThis, "ResizeObserver");
  if (originalScrollTo)
    Object.defineProperty(HTMLElement.prototype, "scrollTo", originalScrollTo);
  else Reflect.deleteProperty(HTMLElement.prototype, "scrollTo");
  jest.useRealTimers();
});

const settleLayout = async () => {
  for (let frame = 0; frame < 12; frame++) {
    // eslint-disable-next-line no-await-in-loop -- Each layout frame must commit before advancing Virtuoso's next scheduled measurement.
    await act(async () => {
      jest.advanceTimersByTime(50);
    });
  }
};

describe("chat history initialization", () => {
  it("reports every visible message for alert suppression while keeping one read high-water mark", async () => {
    const onVisibleItem = jest.fn();
    const onVisibleItems = jest.fn();
    render(
      <VirtuosoMockContext.Provider value={measurements}>
        <Timeline
          header={null}
          items={messages.slice(0, 3)}
          onVisibleItem={onVisibleItem}
          onVisibleItems={onVisibleItems}
          renderItem={(item) => <p>{item.id}</p>}
        />
      </VirtuosoMockContext.Provider>,
    );
    await settleLayout();
    expect(onVisibleItems).toHaveBeenLastCalledWith(messages.slice(0, 3));
    expect(onVisibleItem).toHaveBeenLastCalledWith(messages[2]);
  });
  it("keeps the introduction usable until history arrives, then shows the latest messages", async () => {
    const { rerender } = render(timeline([]));
    await settleLayout();
    expect(screen.getByText("Conversation introduction")).toBeVisible();

    rerender(timeline(messages));
    await settleLayout();

    expect(screen.getByText("message-59")).toBeVisible();
    expect(screen.queryByText("message-0")).not.toBeInTheDocument();
  });

  it("keeps an empty thread root readable and renders the first arriving reply", async () => {
    const { rerender } = render(timeline([], false));
    await settleLayout();
    expect(screen.getByText("Conversation introduction")).toBeVisible();

    rerender(timeline(messages.slice(0, 1), false));
    await settleLayout();

    expect(screen.getByText("Conversation introduction")).toBeVisible();
    expect(screen.getByText("message-0")).toBeVisible();
  });
});
