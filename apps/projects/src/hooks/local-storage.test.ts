/* global beforeEach, describe, expect, it -- Jest globals are provided by the projects test runner. */
import { act, renderHook } from "@testing-library/react";
import { createElement, StrictMode } from "react";
import type { PropsWithChildren } from "react";
import { useLocalStorage } from "./local-storage";

describe("useLocalStorage view preferences", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("restores the correct empty-group preference when the layout key changes", () => {
    localStorage.setItem(
      "view:list",
      JSON.stringify({ showEmptyGroups: false }),
    );
    localStorage.setItem(
      "view:kanban",
      JSON.stringify({ showEmptyGroups: true }),
    );
    const { result, rerender, unmount } = renderHook(
      ({ layout }) =>
        useLocalStorage(`view:${layout}`, { showEmptyGroups: true }),
      { initialProps: { layout: "list" } },
    );

    expect(result.current[0].showEmptyGroups).toBe(false);
    rerender({ layout: "kanban" });
    expect(result.current[0].showEmptyGroups).toBe(true);
    act(() => {
      result.current[1]({ showEmptyGroups: false });
    });
    rerender({ layout: "list" });
    act(() => {
      result.current[1]({ showEmptyGroups: true });
    });
    rerender({ layout: "kanban" });
    expect(result.current[0].showEmptyGroups).toBe(false);
    expect(JSON.parse(localStorage.getItem("view:list")!)).toEqual({
      showEmptyGroups: true,
    });
    unmount();

    const restored = renderHook(() =>
      useLocalStorage("view:kanban", { showEmptyGroups: true }),
    );
    expect(restored.result.current[0].showEmptyGroups).toBe(false);
  });

  it("uses the new key's default and supports consecutive functional updates", () => {
    const { result, rerender } = renderHook(
      ({ storageKey }) => useLocalStorage(storageKey, 0),
      { initialProps: { storageKey: "first" } },
    );
    act(() => {
      result.current[1](3);
    });
    rerender({ storageKey: "second" });
    expect(result.current[0]).toBe(0);
    act(() => {
      result.current[1]((value) => value + 1);
      result.current[1]((value) => value + 1);
    });
    expect(result.current[0]).toBe(2);
    expect(localStorage.getItem("first")).toBe("3");
    expect(localStorage.getItem("second")).toBe("2");
  });

  it("persists only the committed result when React replays functional updates", () => {
    const write = jest.spyOn(Storage.prototype, "setItem");
    try {
      const { result, rerender } = renderHook(
        ({ storageKey }) => useLocalStorage(storageKey, 0),
        {
          initialProps: { storageKey: "first" },
          wrapper: ({ children }: PropsWithChildren) =>
            createElement(StrictMode, null, children),
        },
      );
      expect(write).not.toHaveBeenCalled();
      act(() => {
        result.current[1]((value) => value + 1);
        result.current[1]((value) => value + 1);
      });
      expect(result.current[0]).toBe(2);
      expect(write).toHaveBeenCalledTimes(1);
      expect(write).toHaveBeenCalledWith("first", "2");

      rerender({ storageKey: "second" });
      expect(result.current[0]).toBe(0);
      expect(localStorage.getItem("second")).toBeNull();
      expect(write).toHaveBeenCalledTimes(1);
    } finally {
      write.mockRestore();
    }
  });

  it("keeps hydration-safe preferences across object defaults and functional updates", () => {
    localStorage.setItem("view:list", JSON.stringify({ count: 4 }));
    localStorage.setItem("view:kanban", JSON.stringify({ count: 10 }));
    const { result, rerender } = renderHook(
      ({ layout }) =>
        useLocalStorage(
          `view:${layout}`,
          { count: 0 },
          {
            initializeWithValue: false,
          },
        ),
      { initialProps: { layout: "list" } },
    );

    expect(result.current[0]).toEqual({ count: 4 });
    rerender({ layout: "list" });
    expect(result.current[0]).toEqual({ count: 4 });
    rerender({ layout: "kanban" });
    expect(result.current[0]).toEqual({ count: 10 });
    act(() => {
      result.current[1]((value) => ({ count: value.count + 1 }));
      result.current[1]((value) => ({ count: value.count + 1 }));
    });
    expect(result.current[0]).toEqual({ count: 12 });
    expect(localStorage.getItem("view:list")).toBe('{"count":4}');
    expect(localStorage.getItem("view:kanban")).toBe('{"count":12}');
  });
});
