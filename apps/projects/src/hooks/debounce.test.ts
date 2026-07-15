/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, renderHook } from "@testing-library/react";
import { useDebounce, useDebouncedCallback } from "./debounce";

describe("useDebounce", () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it("only invokes the callback with the latest rapidly queued value", () => {
    const callback = jest.fn();
    const { result } = renderHook(() => useDebounce(callback, 1000));

    act(() => {
      result.current("first");
      result.current("second");
      result.current("final");
      jest.advanceTimersByTime(1000);
    });

    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith("final");
  });

  it("flushes a pending value immediately", () => {
    const callback = jest.fn();
    const { result } = renderHook(() => useDebouncedCallback(callback, 1000));

    act(() => {
      result.current.callback("draft");
      result.current.flush();
    });

    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith("draft");

    act(() => {
      jest.advanceTimersByTime(1000);
    });
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it("can flush a pending value when the editor unmounts", () => {
    const callback = jest.fn();
    const { result, unmount } = renderHook(() =>
      useDebouncedCallback(callback, 1000, { flushOnUnmount: true }),
    );

    act(() => {
      result.current.callback("last edit");
      unmount();
    });

    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith("last edit");
  });

  it("uses the newest callback without resetting the pending timer", () => {
    const firstCallback = jest.fn();
    const secondCallback = jest.fn();
    const { result, rerender } = renderHook(
      ({ callback }) => useDebounce(callback, 1000),
      { initialProps: { callback: firstCallback } },
    );

    act(() => {
      result.current("draft");
    });
    rerender({ callback: secondCallback });

    act(() => {
      jest.advanceTimersByTime(1000);
    });

    expect(firstCallback).not.toHaveBeenCalled();
    expect(secondCallback).toHaveBeenCalledWith("draft");
  });
});
