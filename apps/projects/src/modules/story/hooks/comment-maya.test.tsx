/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { commentMutationKey } from "@/shared/story/comment-cache";
import { storyKeys } from "@/shared/story/cache-keys";
import type { CommentMayaRun } from "../types/comment-maya";
import { getStoryMayaRuns } from "../queries/comment-maya";
import { useMayaRunRefresh, useStoryMayaRuns } from "./comment-maya";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("../queries/comment-maya", () => ({
  getStoryMayaRuns: jest.fn(),
  updateCommentMayaRun: jest.fn(),
}));

describe("Maya result cache refresh", () => {
  it("refreshes an already published reply while delivery finishes and again on terminal status without looping", () => {
    const client = new QueryClient();
    const invalidate = jest.spyOn(client, "invalidateQueries");
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const run: CommentMayaRun = {
      id: "run",
      sourceId: "request",
      requesterId: "me",
      status: "running",
      summary: "",
      updatedAt: "2026-10-16T12:00:00Z",
      version: 2,
      replyCommentId: "reply",
    };
    const { rerender } = renderHook(
      ({ runs }) => {
        useMayaRunRefresh("story", runs);
      },
      { wrapper, initialProps: { runs: [run] } },
    );
    expect(invalidate).toHaveBeenCalledWith({
      queryKey: storyKeys.detail("acme", "story"),
      exact: true,
    });
    const first = invalidate.mock.calls.length;
    rerender({ runs: [{ ...run, status: "succeeded" }] });
    expect(invalidate.mock.calls.length).toBeGreaterThan(first);
    const terminal = invalidate.mock.calls.length;
    rerender({ runs: [{ ...run, status: "succeeded" }] });
    expect(invalidate).toHaveBeenCalledTimes(terminal);
    client.clear();
  });
  it("waits for another comment mutation before refreshing a completed result and refreshes each version once", async () => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const invalidate = jest.spyOn(client, "invalidateQueries");
    client.setQueryData(storyKeys.mayaRuns("acme", "story"), {
      enabled: true,
      runs: [],
    });
    client.setQueryData(storyKeys.detail("acme", "story"), { id: "story" });
    client.setQueryData(storyKeys.list("acme", "all"), []);
    let settle!: () => void;
    const mutation = client.getMutationCache().build(client, {
      mutationKey: commentMutationKey("acme"),
      mutationFn: (_variables: { storyId: string }) =>
        new Promise<void>((resolve) => {
          settle = resolve;
        }),
    });
    let pending!: Promise<void>;
    await act(async () => {
      pending = mutation.execute({ storyId: "story" });
    });
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const active: CommentMayaRun = {
      id: "run",
      sourceId: "request",
      requesterId: "me",
      status: "running",
      summary: "",
      updatedAt: "2026-10-16T12:00:00Z",
      version: 1,
    };
    const { rerender } = renderHook(
      ({ runs }) => {
        useMayaRunRefresh("story", runs);
      },
      { wrapper, initialProps: { runs: [active] } },
    );
    const completed: CommentMayaRun = {
      ...active,
      status: "succeeded",
      version: 2,
      replyCommentId: "reply",
    };
    rerender({ runs: [completed] });
    expect(invalidate).not.toHaveBeenCalled();
    await act(async () => {
      settle();
      await pending;
    });
    await waitFor(() => {
      expect(invalidate).toHaveBeenCalledWith({
        queryKey: storyKeys.detail("acme", "story"),
        exact: true,
      });
    });
    expect(invalidate).toHaveBeenCalledWith({
      queryKey: storyKeys.activitiesInfinite("acme", "story"),
      refetchType: "all",
    });
    expect(
      client.getQueryState(storyKeys.mayaRuns("acme", "story"))?.isInvalidated,
    ).toBe(false);
    expect(
      client.getQueryState(storyKeys.detail("acme", "story"))?.isInvalidated,
    ).toBe(true);
    expect(
      client.getQueryState(storyKeys.list("acme", "all"))?.isInvalidated,
    ).toBe(true);
    const count = invalidate.mock.calls.length;
    rerender({ runs: [{ ...completed }] });
    expect(invalidate).toHaveBeenCalledTimes(count);
    client.clear();
  });
});

const getRuns = jest.mocked(getStoryMayaRuns);
afterEach(() => {
  jest.useRealTimers();
});

describe("Maya polling", () => {
  it("lets only the feed poll and slows down after work finishes", async () => {
    jest.useFakeTimers();
    getRuns.mockReset();
    const run: CommentMayaRun = {
      id: "run",
      sourceId: "request",
      requesterId: "me",
      status: "running",
      summary: "",
      updatedAt: "2026-10-16T12:00:00Z",
      version: 1,
    };
    getRuns.mockResolvedValue({ enabled: true, runs: [run] });
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const { unmount } = renderHook(
      () => {
        useStoryMayaRuns("story", { poll: true });
        useStoryMayaRuns("story");
      },
      { wrapper },
    );
    await act(async () => {
      await jest.advanceTimersByTimeAsync(100);
    });
    expect(getRuns).toHaveBeenCalledTimes(1);
    await act(async () => {
      await jest.advanceTimersByTimeAsync(2_000);
    });
    expect(getRuns).toHaveBeenCalledTimes(2);
    const completed = {
      enabled: true,
      runs: [{ ...run, status: "succeeded" as const, version: 2 }],
    };
    getRuns.mockResolvedValue(completed);
    await act(async () => {
      client.setQueryData(storyKeys.mayaRuns("acme", "story"), completed);
      await jest.advanceTimersByTimeAsync(100);
    });
    await act(async () => {
      await jest.advanceTimersByTimeAsync(10_000);
    });
    expect(getRuns).toHaveBeenCalledTimes(2);
    await act(async () => {
      await jest.advanceTimersByTimeAsync(20_000);
    });
    expect(getRuns).toHaveBeenCalledTimes(3);
    unmount();
    client.clear();
  });
});
