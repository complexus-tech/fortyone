/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, renderHook } from "@testing-library/react";
import { useWorkTab } from "./use-work-tab";

let mockUserId = "user-a";
let mockWorkspaceSlug = "workspace-a";
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUserId } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspaceSlug }),
}));
const key = "maya:work-tab:v1:user-a:workspace-a";

beforeEach(() => {
  localStorage.clear();
  mockUserId = "user-a";
  mockWorkspaceSlug = "workspace-a";
});
afterEach(() => {
  jest.restoreAllMocks();
});

describe("remembered Maya work tab", () => {
  it("restores a selected tab after remounting", () => {
    const first = renderHook(() => useWorkTab());
    act(() => {
      first.result.current[1]("created");
    });
    expect(localStorage.getItem(key)).toBe("created");
    first.unmount();
    const next = renderHook(() => useWorkTab());
    expect(next.result.current[0]).toBe("created");
  });
  it("isolates the selection when the user or workspace changes", () => {
    const { result, rerender } = renderHook(() => useWorkTab());
    act(() => {
      result.current[1]("assigned");
    });
    mockWorkspaceSlug = "workspace-b";
    rerender();
    expect(result.current[0]).toBe("all");
    mockWorkspaceSlug = "workspace-a";
    mockUserId = "user-b";
    rerender();
    expect(result.current[0]).toBe("all");
    mockUserId = "user-a";
    rerender();
    expect(result.current[0]).toBe("assigned");
  });
  it("ignores unsupported saved values and responds to other-tab updates", () => {
    localStorage.setItem(key, "unsupported");
    const { result } = renderHook(() => useWorkTab());
    expect(result.current[0]).toBe("all");
    act(() => {
      localStorage.setItem(key, "created");
      window.dispatchEvent(new StorageEvent("storage", { key }));
    });
    expect(result.current[0]).toBe("created");
  });
  it("keeps switching usable when storage is blocked", () => {
    mockUserId = "blocked-storage-user";
    jest.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("Unavailable");
    });
    jest.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("Unavailable");
    });
    const { result } = renderHook(() => useWorkTab());
    act(() => {
      result.current[1]("assigned");
    });
    expect(result.current[0]).toBe("assigned");
  });
});
