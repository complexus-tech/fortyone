/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { useTimelinePagination } from "./use-timeline-pagination";

type Options = Parameters<typeof useTimelinePagination>[0];
const options = (overrides: Partial<Options> = {}): Options => ({
  enabled: true,
  hasOlder: true,
  isFetching: false,
  firstItemId: "message-100",
  positionReady: true,
  onLoadOlder: jest.fn().mockResolvedValue(undefined),
  ...overrides,
});
afterEach(() => {
  cleanup();
  jest.restoreAllMocks();
});

const mockAnimationFrames = () => {
  const frames = new Map<number, FrameRequestCallback>();
  let nextFrame = 0;
  jest.spyOn(window, "requestAnimationFrame").mockImplementation((callback) => {
    frames.set(++nextFrame, callback);
    return nextFrame;
  });
  jest.spyOn(window, "cancelAnimationFrame").mockImplementation((frame) => {
    frames.delete(frame);
  });
  return () => {
    const pending = [...frames.values()];
    frames.clear();
    act(() => {
      for (const callback of pending) callback(0);
    });
  };
};

describe("automatic older history pagination", () => {
  it("waits for initial positioning before accepting the transient top state", async () => {
    const props = options({ positionReady: false });
    const { result, rerender } = renderHook(useTimelinePagination, {
      initialProps: props,
    });
    act(() => {
      result.current.onNearTopChange(true);
    });
    await act(async () => undefined);
    expect(props.onLoadOlder).not.toHaveBeenCalled();

    act(() => {
      result.current.onNearTopChange(false);
    });
    rerender({ ...props, positionReady: true });
    await act(async () => undefined);
    expect(props.onLoadOlder).not.toHaveBeenCalled();

    act(() => {
      result.current.onNearTopChange(true);
    });
    await waitFor(() => {
      expect(props.onLoadOlder).toHaveBeenCalledTimes(1);
    });
  });

  it("fills a short viewport once positioning is ready", async () => {
    const props = options({ positionReady: false });
    const { result, rerender } = renderHook(useTimelinePagination, {
      initialProps: props,
    });
    act(() => {
      result.current.onNearTopChange(true);
    });
    rerender({ ...props, positionReady: true });
    await waitFor(() => {
      expect(props.onLoadOlder).toHaveBeenCalledTimes(1);
    });
    rerender({ ...props, positionReady: true, firstItemId: "message-50" });
    await waitFor(() => {
      expect(props.onLoadOlder).toHaveBeenCalledTimes(2);
    });
    rerender({
      ...props,
      positionReady: true,
      firstItemId: "message-1",
      hasOlder: false,
    });
    await act(async () => undefined);
    expect(props.onLoadOlder).toHaveBeenCalledTimes(2);
  });

  it("serializes boundary requests even before query fetching state is updated", async () => {
    let finish: () => void = () => undefined;
    const firstRequest = new Promise<void>((resolve) => {
      finish = resolve;
    });
    const load = jest
      .fn()
      .mockReturnValueOnce(firstRequest)
      .mockResolvedValue(undefined);
    const props = options({ onLoadOlder: load });
    const { result, rerender } = renderHook(useTimelinePagination, {
      initialProps: props,
    });
    act(() => {
      result.current.onNearTopChange(true);
    });
    await waitFor(() => {
      expect(load).toHaveBeenCalledTimes(1);
    });
    act(() => {
      result.current.onNearTopChange(false);
      result.current.onNearTopChange(true);
    });
    rerender({ ...props, firstItemId: "message-50" });
    await act(async () => undefined);
    expect(load).toHaveBeenCalledTimes(1);

    await act(async () => {
      finish();
    });
    await waitFor(() => {
      expect(load).toHaveBeenCalledTimes(2);
    });
  });

  it("does not cancel or compete with an existing query refresh", async () => {
    const props = options({ isFetching: true });
    const { result, rerender } = renderHook(useTimelinePagination, {
      initialProps: props,
    });
    act(() => {
      result.current.onNearTopChange(true);
    });
    await act(async () => undefined);
    expect(props.onLoadOlder).not.toHaveBeenCalled();
    rerender({ ...props, isFetching: false });
    await waitFor(() => {
      expect(props.onLoadOlder).toHaveBeenCalledTimes(1);
    });
  });

  it("leaves failed boundaries for explicit retry instead of retrying on each scroll", async () => {
    const failure = new Error("Connection interrupted");
    const load = jest.fn().mockRejectedValue(failure);
    const props = options({ onLoadOlder: load });
    const { result, rerender } = renderHook(useTimelinePagination, {
      initialProps: props,
    });
    act(() => {
      result.current.onNearTopChange(true);
    });
    await waitFor(() => {
      expect(load).toHaveBeenCalledTimes(1);
    });
    rerender({ ...props, error: failure });
    act(() => {
      result.current.onNearTopChange(false);
    });
    act(() => {
      result.current.onNearTopChange(true);
    });
    rerender({ ...props, error: null });
    await act(async () => undefined);
    expect(load).toHaveBeenCalledTimes(1);

    load.mockResolvedValue(undefined);
    rerender({ ...props, firstItemId: "message-50", error: failure });
    await act(async () => undefined);
    expect(load).toHaveBeenCalledTimes(1);
    rerender({ ...props, firstItemId: "message-50", error: null });
    await waitFor(() => {
      expect(load).toHaveBeenCalledTimes(2);
    });
  });

  it("does not load while inactive or after leaving the top threshold", async () => {
    const props = options({ enabled: false });
    const { result, rerender } = renderHook(useTimelinePagination, {
      initialProps: props,
    });
    act(() => {
      result.current.onNearTopChange(true);
    });
    await act(async () => undefined);
    expect(props.onLoadOlder).not.toHaveBeenCalled();
    act(() => {
      result.current.onNearTopChange(false);
    });
    rerender({ ...props, enabled: true });
    await act(async () => undefined);
    expect(props.onLoadOlder).not.toHaveBeenCalled();
  });

  it("rechecks the actual viewport after prepending restores its scroll anchor", async () => {
    const flushFrame = mockAnimationFrames();
    let atTop = true;
    const props = options({ isNearTop: () => atTop });
    const { result, rerender } = renderHook(useTimelinePagination, {
      initialProps: props,
    });
    act(() => {
      result.current.onNearTopChange(true);
    });
    expect(props.onLoadOlder).not.toHaveBeenCalled();
    flushFrame();
    await act(async () => undefined);
    expect(props.onLoadOlder).toHaveBeenCalledTimes(1);

    // Virtuoso's prior top event can remain true until after the prepend render.
    rerender({ ...props, firstItemId: "message-50" });
    atTop = false;
    flushFrame();
    await act(async () => undefined);
    expect(props.onLoadOlder).toHaveBeenCalledTimes(1);

    act(() => {
      result.current.onNearTopChange(false);
    });
    atTop = true;
    act(() => {
      result.current.onNearTopChange(true);
    });
    flushFrame();
    await act(async () => undefined);
    expect(props.onLoadOlder).toHaveBeenCalledTimes(2);
  });

  it("cancels the scheduled load when paging becomes inactive or the list unmounts", async () => {
    const flushFrame = mockAnimationFrames();
    const props = options();
    const { result, rerender, unmount } = renderHook(useTimelinePagination, {
      initialProps: props,
    });
    act(() => {
      result.current.onNearTopChange(true);
    });
    rerender({ ...props, enabled: false });
    flushFrame();
    await act(async () => undefined);
    expect(props.onLoadOlder).not.toHaveBeenCalled();

    rerender(props);
    unmount();
    flushFrame();
    await act(async () => undefined);
    expect(props.onLoadOlder).not.toHaveBeenCalled();
  });
});
