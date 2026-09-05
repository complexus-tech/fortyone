/* global beforeEach, describe, expect, it -- Jest globals are provided by the projects test runner. */
import { act, renderHook } from "@testing-library/react";
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
});
