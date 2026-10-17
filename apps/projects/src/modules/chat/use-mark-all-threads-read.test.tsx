/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { markAllThreadsRead } from "./api";
import { chatKeys } from "./constants";
import { reconcileChatQueries } from "./realtime";
import { useMarkAllThreadsRead } from "./use-mark-all-threads-read";

let mockWorkspaceSlug = "workspace";
let mockUserId = "viewer";
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspaceSlug }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({
    data: mockUserId ? { user: { id: mockUserId } } : null,
  }),
}));
jest.mock("./api", () => ({ markAllThreadsRead: jest.fn() }));
jest.mock("./realtime", () => ({ reconcileChatQueries: jest.fn() }));

let client: QueryClient;
const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);

beforeEach(() => {
  mockWorkspaceSlug = "workspace";
  mockUserId = "viewer";
  client = new QueryClient({
    defaultOptions: { mutations: { retry: 3 }, queries: { retry: false } },
  });
  jest.mocked(reconcileChatQueries).mockResolvedValue(undefined);
});

afterEach(() => {
  cleanup();
  client.clear();
  jest.resetAllMocks();
});

describe("marking all threads as read", () => {
  it("shares one request and pending state across conversation remounts until reconciliation finishes", async () => {
    let completeRequest!: () => void;
    let completeReconciliation!: () => void;
    jest.mocked(markAllThreadsRead).mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          completeRequest = resolve;
        }),
    );
    jest.mocked(reconcileChatQueries).mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          completeReconciliation = resolve;
        }),
    );
    const first = renderHook(useMarkAllThreadsRead, { wrapper });
    let request!: Promise<void>;
    act(() => {
      request = first.result.current.mutateAsync();
      expect(first.result.current.mutateAsync()).toBe(request);
    });
    await waitFor(() => {
      expect(markAllThreadsRead).toHaveBeenCalledTimes(1);
    });
    first.unmount();
    const remounted = renderHook(useMarkAllThreadsRead, { wrapper });
    expect(remounted.result.current.isPending).toBe(true);
    act(() => {
      expect(remounted.result.current.mutateAsync()).toBe(request);
    });
    await act(async () => {
      completeRequest();
    });
    expect(remounted.result.current.isPending).toBe(true);
    expect(markAllThreadsRead).toHaveBeenCalledTimes(1);
    act(() => {
      expect(remounted.result.current.mutateAsync()).toBe(request);
    });
    await act(async () => {
      completeReconciliation();
      await request;
    });
    await waitFor(() => {
      expect(remounted.result.current.isPending).toBe(false);
    });
    expect(reconcileChatQueries).toHaveBeenCalledTimes(1);
  });

  it("keeps pending requests separate for each workspace and account", async () => {
    const finish: (() => void)[] = [];
    jest.mocked(markAllThreadsRead).mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          finish.push(resolve);
        }),
    );
    const first = renderHook(useMarkAllThreadsRead, { wrapper });
    let firstRequest!: Promise<void>;
    act(() => {
      firstRequest = first.result.current.mutateAsync();
    });
    await waitFor(() => {
      expect(markAllThreadsRead).toHaveBeenCalledTimes(1);
    });
    first.unmount();

    mockWorkspaceSlug = "other-workspace";
    const otherWorkspace = renderHook(useMarkAllThreadsRead, { wrapper });
    expect(otherWorkspace.result.current.isPending).toBe(false);
    let workspaceRequest!: Promise<void>;
    act(() => {
      workspaceRequest = otherWorkspace.result.current.mutateAsync();
    });
    await waitFor(() => {
      expect(markAllThreadsRead).toHaveBeenCalledTimes(2);
    });
    otherWorkspace.unmount();

    mockWorkspaceSlug = "workspace";
    mockUserId = "other-viewer";
    const otherAccount = renderHook(useMarkAllThreadsRead, { wrapper });
    expect(otherAccount.result.current.isPending).toBe(false);
    let accountRequest!: Promise<void>;
    act(() => {
      accountRequest = otherAccount.result.current.mutateAsync();
    });
    await waitFor(() => {
      expect(markAllThreadsRead).toHaveBeenCalledTimes(3);
    });
    expect(new Set([firstRequest, workspaceRequest, accountRequest]).size).toBe(
      3,
    );
    await act(async () => {
      finish.forEach((resolve) => {
        resolve();
      });
      await Promise.all([firstRequest, workspaceRequest, accountRequest]);
    });
    expect(
      jest
        .mocked(reconcileChatQueries)
        .mock.calls.map(([, workspace]) => workspace),
    ).toEqual(["workspace", "other-workspace", "workspace"]);
  });

  it("preserves cached messages and newer unread replies until reconciliation", async () => {
    let complete: () => void = () => undefined;
    jest.mocked(markAllThreadsRead).mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          complete = resolve;
        }),
    );
    const unreadKey = [...chatKeys.unread("workspace"), "viewer"];
    const unread = {
      totalUnread: 8,
      unreadThreads: 2,
      conversations: [
        { conversationId: "room", unreadRootCount: 5, unreadReplyCount: 3 },
      ],
    };
    client.setQueryData(unreadKey, unread);
    const { result } = renderHook(useMarkAllThreadsRead, { wrapper });
    let request: Promise<void>;
    act(() => {
      request = result.current.mutateAsync();
    });
    await waitFor(() => {
      expect(result.current.isPending).toBe(true);
    });
    expect(client.getQueryData(unreadKey)).toEqual(unread);
    const updated = { ...unread, totalUnread: 9, unreadThreads: 3 };
    client.setQueryData(unreadKey, updated);

    await act(async () => {
      complete();
      await request;
    });
    expect(client.getQueryData(unreadKey)).toEqual(updated);
    expect(reconcileChatQueries).toHaveBeenCalledWith(client, "workspace");
  });

  it("reconciles the original workspace when navigation changes during the request", async () => {
    let complete: () => void = () => undefined;
    jest.mocked(markAllThreadsRead).mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          complete = resolve;
        }),
    );
    const { result, rerender } = renderHook(useMarkAllThreadsRead, { wrapper });
    let request: Promise<void>;
    act(() => {
      request = result.current.mutateAsync();
    });
    await waitFor(() => {
      expect(markAllThreadsRead).toHaveBeenCalledWith({
        workspaceSlug: "workspace",
      });
    });
    mockWorkspaceSlug = "other-workspace";
    rerender();

    await act(async () => {
      complete();
      await request;
    });
    expect(reconcileChatQueries).toHaveBeenCalledWith(client, "workspace");
  });

  it("surfaces failures without retrying a new read snapshot or clearing the cache", async () => {
    const failure = new Error("Offline");
    jest.mocked(markAllThreadsRead).mockRejectedValue(failure);
    const unreadKey = chatKeys.unread("workspace");
    client.setQueryData(unreadKey, { unreadThreads: 3 });
    const { result } = renderHook(useMarkAllThreadsRead, { wrapper });

    await act(async () => {
      await expect(result.current.mutateAsync()).rejects.toBe(failure);
    });
    expect(markAllThreadsRead).toHaveBeenCalledTimes(1);
    expect(reconcileChatQueries).not.toHaveBeenCalled();
    expect(client.getQueryData(unreadKey)).toEqual({ unreadThreads: 3 });
    jest.mocked(markAllThreadsRead).mockResolvedValue(undefined);
    await act(async () => {
      await result.current.mutateAsync();
    });
    expect(markAllThreadsRead).toHaveBeenCalledTimes(2);
  });

  it("rejects actions before an authenticated scope is available", async () => {
    mockUserId = "";
    const { result } = renderHook(useMarkAllThreadsRead, { wrapper });
    await expect(result.current.mutateAsync()).rejects.toThrow("Sign in");
    expect(markAllThreadsRead).not.toHaveBeenCalled();
  });
});
