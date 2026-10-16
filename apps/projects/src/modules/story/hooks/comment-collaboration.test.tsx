/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { PropsWithChildren } from "react";
import { act, renderHook } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/shared/story/cache-keys";
import type {
  CommentState,
  StoryCommentState,
} from "../types/comment-collaboration";
import {
  markCommentsRead,
  setCommentReaction,
  setCommentResolution,
} from "../queries/comment-collaboration";
import {
  useCommentCollaboration,
  useMarkCommentsRead,
} from "./comment-collaboration";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
}));
jest.mock("../queries/comment-collaboration", () => ({
  markCommentsRead: jest.fn(),
  setCommentReaction: jest.fn(),
  setCommentResolution: jest.fn(),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const entry: CommentState = {
  commentId: "comment",
  reactions: [],
  resolvedAt: null,
  resolvedBy: null,
};
const state: StoryCommentState = {
  lastReadAt: "2026-10-15T12:00:00Z",
  observedAt: "2026-10-16T12:00:00Z",
  unreadCount: 1,
  latestCommentId: "comment",
  comments: [entry],
  threadReads: {},
};
const queryKey = storyKeys.commentState("workspace", "story");
const setup = () => {
  const client = new QueryClient({
    defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
  });
  client.setQueryData(queryKey, state);
  const wrapper = ({ children }: PropsWithChildren) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return { client, wrapper };
};

beforeEach(() => {
  jest.clearAllMocks();
});

describe("comment collaboration mutations", () => {
  it("merges per-thread receipts without consuming another thread or replacing collaboration data", async () => {
    const { client, wrapper } = setup();
    client.setQueryData(queryKey, {
      ...state,
      threadReads: { other: "2026-10-16T12:00:00Z" },
    });
    jest
      .mocked(markCommentsRead)
      .mockResolvedValue({
        ...state,
        threadReads: {
          root: "2026-10-16T11:00:00Z",
          other: "2026-10-16T10:00:00Z",
        },
      });
    const { result } = renderHook(() => useMarkCommentsRead("story"), {
      wrapper,
    });
    await act(async () => {
      await result.current.mutateAsync({
        threadId: "root",
        observedAt: "2026-10-16T11:00:00Z",
      });
    });
    expect(client.getQueryData<StoryCommentState>(queryKey)).toMatchObject({
      lastReadAt: state.lastReadAt,
      threadReads: {
        root: "2026-10-16T11:00:00Z",
        other: "2026-10-16T12:00:00Z",
      },
      comments: state.comments,
    });
    expect(markCommentsRead).toHaveBeenCalledWith(
      "story",
      { threadId: "root", observedAt: "2026-10-16T11:00:00Z" },
      { workspaceSlug: "workspace" },
    );
  });

  it("rolls back only a failed thread receipt", async () => {
    const { client, wrapper } = setup();
    client.setQueryData(queryKey, {
      ...state,
      threadReads: { other: "2026-10-16T12:00:00Z" },
    });
    jest.mocked(markCommentsRead).mockRejectedValue(new Error("offline"));
    const { result } = renderHook(() => useMarkCommentsRead("story"), {
      wrapper,
    });
    await act(async () => {
      await expect(
        result.current.mutateAsync({
          threadId: "root",
          observedAt: "2026-10-16T11:00:00Z",
        }),
      ).rejects.toThrow("offline");
    });
    expect(
      client.getQueryData<StoryCommentState>(queryKey)?.threadReads,
    ).toEqual({ other: "2026-10-16T12:00:00Z" });
  });
  it("merges reaction and resolution responses without overwriting unrelated fields or threads", async () => {
    const { client, wrapper } = setup();
    const other: CommentState = {
      ...entry,
      commentId: "other",
      resolvedAt: "2026-10-16T09:00:00Z",
    };
    client.setQueryData(queryKey, { ...state, comments: [entry, other] });
    jest.mocked(setCommentResolution).mockResolvedValue({
      ...entry,
      resolvedAt: "2026-10-16T12:00:00Z",
      resolvedBy: "me",
    });
    jest.mocked(setCommentReaction).mockResolvedValue({
      ...entry,
      reactions: [{ emoji: "👍", count: 1, reacted: true }],
    });
    const { result } = renderHook(
      () => useCommentCollaboration("story", "comment"),
      { wrapper },
    );
    await act(async () => {
      await result.current.resolution.mutateAsync(true);
    });
    await act(async () => {
      await result.current.reaction.mutateAsync({ emoji: "👍", reacted: true });
    });
    const cached = client.getQueryData<StoryCommentState>(queryKey)!;
    expect(cached.comments[0]).toMatchObject({
      resolvedBy: "me",
      resolvedAt: "2026-10-16T12:00:00Z",
      reactions: [{ emoji: "👍", count: 1, reacted: true }],
    });
    expect(cached.comments[1]).toEqual(other);
    expect(setCommentReaction).toHaveBeenCalledWith("comment", "👍", true, {
      workspaceSlug: "workspace",
    });
  });
  it("keeps saved reactions intact and surfaces a failed request", async () => {
    const { client, wrapper } = setup();
    jest.mocked(setCommentReaction).mockRejectedValue(new Error("offline"));
    const { result } = renderHook(
      () => useCommentCollaboration("story", "comment"),
      { wrapper },
    );
    await act(async () => {
      await expect(
        result.current.reaction.mutateAsync({ emoji: "👀", reacted: true }),
      ).rejects.toThrow("offline");
    });
    expect(client.getQueryData(queryKey)).toEqual(state);
    expect(toast.error).toHaveBeenCalledWith(
      "Could not update reaction. Please try again.",
    );
  });
  it("does not roll back read position or replace newer reaction data with a read response", async () => {
    const { client, wrapper } = setup();
    const current = {
      ...state,
      lastReadAt: "2026-10-16T12:00:00Z",
      comments: [
        { ...entry, reactions: [{ emoji: "✅", count: 2, reacted: true }] },
      ],
    };
    client.setQueryData(queryKey, current);
    jest
      .mocked(markCommentsRead)
      .mockResolvedValue({ ...state, lastReadAt: "2026-10-16T11:00:00Z" });
    const { result } = renderHook(() => useMarkCommentsRead("story"), {
      wrapper,
    });
    await act(async () => {
      await result.current.mutateAsync({ observedAt: "2026-10-16T11:00:00Z" });
    });
    expect(client.getQueryData(queryKey)).toEqual(current);
    jest.mocked(markCommentsRead).mockResolvedValue({
      ...state,
      lastReadAt: "2026-10-16T13:00:00Z",
      unreadCount: 0,
    });
    await act(async () => {
      await result.current.mutateAsync({ observedAt: "2026-10-16T13:00:00Z" });
    });
    expect(client.getQueryData<StoryCommentState>(queryKey)?.comments).toEqual(
      current.comments,
    );
    expect(client.getQueryData<StoryCommentState>(queryKey)?.lastReadAt).toBe(
      "2026-10-16T13:00:00Z",
    );
  });
});
