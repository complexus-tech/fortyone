import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { native } from "../native";
import { useDesktop } from "./use-desktop";
import type { WorkspaceData } from "../types";

vi.mock("../native", () => ({
  getBoundAccountId: () => "user",
  bindNativeAccount: vi.fn(),
  native: {
    forAccount: () => native,
    bootstrap: vi.fn(),
    readWorkspace: vi.fn(),
    selectWorkspace: vi.fn(),
    sync: vi.fn(),
    listen: vi.fn(),
    saveOperation: vi.fn(),
    logout: vi.fn(),
  },
  errorMessage: (error: unknown) =>
    error instanceof Error ? error.message : String(error),
}));
vi.mock("../drafts", () => ({
  flushDrafts: vi.fn().mockResolvedValue(undefined),
}));
const snapshot = (workspaceId: string): WorkspaceData => ({
  workspaceId,
  stories: [],
  documents: [],
  comments: [],
  statuses: [],
  teams: [],
  pending: [],
  lastSyncedAt: null,
});

beforeEach(() => {
  vi.clearAllMocks();
  Object.defineProperty(navigator, "onLine", {
    configurable: true,
    value: false,
  });
  vi.mocked(native.bootstrap).mockResolvedValue({
    user: { id: "user", fullName: "Test User", email: "test@example.com" },
    workspaces: [
      { id: "one", name: "One", slug: "one" },
      { id: "two", name: "Two", slug: "two" },
    ],
    activeWorkspaceId: "one",
    needsSignIn: false,
  });
  vi.mocked(native.readWorkspace).mockImplementation(async (id) =>
    snapshot(id),
  );
  vi.mocked(native.listen).mockResolvedValue(() => undefined);
});

describe("offline workspace lifecycle", () => {
  it("does not retry a denied credential store on automatic reconnect but allows explicit synchronization", async () => {
    vi.mocked(native.sync).mockResolvedValue(snapshot("one"));
    const { result } = renderHook(useDesktop);
    await waitFor(() => expect(result.current.starting).toBe(false));
    Object.defineProperty(navigator, "onLine", {
      configurable: true,
      value: true,
    });
    await act(async () => {
      window.dispatchEvent(new Event("online"));
    });
    expect(native.sync).toHaveBeenLastCalledWith("one", false);
    await act(async () => {
      await result.current.sync(true);
    });
    expect(native.sync).toHaveBeenLastCalledWith("one", true);
  });
  it("opens downloaded data without attempting a network sync while offline", async () => {
    const { result } = renderHook(useDesktop);
    await waitFor(() => expect(result.current.starting).toBe(false));
    expect(result.current.data?.workspaceId).toBe("one");
    expect(native.sync).not.toHaveBeenCalled();
  });
  it("refreshes the displayed local snapshot when background warmup emits a state change", async () => {
    const { result } = renderHook(useDesktop);
    await waitFor(() => expect(result.current.starting).toBe(false));
    const downloaded = {
      ...snapshot("one"),
      lastSyncedAt: "2026-10-14T12:00:00Z",
    };
    vi.mocked(native.readWorkspace).mockResolvedValue(downloaded);
    const stateListener = vi
      .mocked(native.listen)
      .mock.calls.find(([name]) => name === "desktop-state-changed")?.[1];
    expect(stateListener).toBeDefined();
    await act(async () => {
      stateListener?.({ event: "desktop-state-changed", id: 1, payload: null });
    });
    await waitFor(() =>
      expect(result.current.data?.lastSyncedAt).toBe(downloaded.lastSyncedAt),
    );
    expect(native.sync).not.toHaveBeenCalled();
  });
  it("indicates switching immediately, suppresses duplicate selections, and retains the old workspace on failure", async () => {
    let rejectSwitch!: (error: Error) => void;
    vi.mocked(native.selectWorkspace).mockImplementation(
      () =>
        new Promise((_, reject) => {
          rejectSwitch = reject;
        }),
    );
    const { result } = renderHook(useDesktop);
    await waitFor(() => expect(result.current.starting).toBe(false));
    let switchPromise!: Promise<boolean>;
    act(() => {
      switchPromise = result.current.selectWorkspace("two");
    });
    expect(result.current.switchingId).toBe("two");
    await act(async () => {
      await result.current.selectWorkspace("two");
    });
    expect(native.selectWorkspace).toHaveBeenCalledTimes(1);
    await act(async () => {
      rejectSwitch(new Error("Workspace unavailable"));
      await switchPromise;
    });
    expect(result.current.data?.workspaceId).toBe("one");
    expect(result.current.switchingId).toBeNull();
    expect(result.current.error).toBe("Workspace unavailable");
  });
  it("removes displayed workspace data when native access validation denies a reload", async () => {
    const { result } = renderHook(useDesktop);
    await waitFor(() => expect(result.current.data?.workspaceId).toBe("one"));
    vi.mocked(native.readWorkspace).mockRejectedValue(
      new Error("Workspace access expired"),
    );
    await act(async () => {
      await result.current.reload();
    });
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBe("Workspace access expired");
  });
  it("clears previous-account data immediately and rejects old renderer callbacks for the same workspace", async () => {
    const { result } = renderHook(useDesktop);
    await waitFor(() => expect(result.current.data?.workspaceId).toBe("one"));
    const staleAccept = result.current.acceptData;
    const staleSave = result.current.saveOperation;
    const nextSession = await native.bootstrap();
    vi.mocked(native.bootstrap).mockResolvedValue({
      ...nextSession,
      user: {
        id: "other-account",
        fullName: "Another User",
        email: "another@example.com",
      },
    });
    let finishRead!: (data: WorkspaceData) => void;
    vi.mocked(native.readWorkspace).mockReturnValue(
      new Promise((resolve) => {
        finishRead = resolve;
      }),
    );
    let reload!: Promise<void>;
    act(() => {
      reload = result.current.reload();
    });
    await waitFor(() =>
      expect(result.current.session?.user?.id).toBe("other-account"),
    );
    expect(result.current.data).toBeNull();
    act(() => staleAccept(snapshot("one")));
    expect(result.current.data).toBeNull();
    await expect(
      staleSave({
        workspaceId: "one",
        resourceId: "story",
        kind: "story.update",
        payload: {},
        baseVersion: "version",
      }),
    ).rejects.toThrow("account changed");
    expect(native.saveOperation).not.toHaveBeenCalled();
    await act(async () => {
      finishRead(snapshot("one"));
      await reload;
    });
    expect(result.current.data?.workspaceId).toBe("one");
  });
  it("ignores a late synchronization response from the previous workspace", async () => {
    let finishSync!: (value: WorkspaceData) => void;
    vi.mocked(native.sync).mockImplementation(
      () =>
        new Promise((resolve) => {
          finishSync = resolve;
        }),
    );
    vi.mocked(native.selectWorkspace).mockResolvedValue(snapshot("two"));
    const { result } = renderHook(useDesktop);
    await waitFor(() => expect(result.current.starting).toBe(false));
    let syncPromise!: Promise<void>;
    act(() => {
      syncPromise = result.current.sync();
    });
    await act(async () => {
      await result.current.selectWorkspace("two");
    });
    await act(async () => {
      finishSync(snapshot("one"));
      await syncPromise;
    });
    expect(result.current.data?.workspaceId).toBe("two");
    expect(result.current.session?.activeWorkspaceId).toBe("two");
  });
});
