/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { PropsWithChildren } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/shared/story/cache-keys";
import type {
  CommentState,
  StoryCommentState,
} from "../types/comment-collaboration";
import {
  getCommentState,
  setCommentReaction,
} from "../queries/comment-collaboration";
import {
  useCommentCollaboration,
  useStoryCommentState,
} from "./comment-collaboration";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
}));
jest.mock("../queries/comment-collaboration", () => ({
  getCommentState: jest.fn(),
  setCommentReaction: jest.fn(),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const entry: CommentState = {
  commentId: "comment",
  reactions: [],
  resolvedAt: null,
  resolvedBy: null,
};
const state: StoryCommentState = {
  lastReadAt: null,
  observedAt: "2026-10-16T12:00:00Z",
  unreadCount: 1,
  latestCommentId: "comment",
  comments: [entry],
  threadReads: {},
};
const queryKey = storyKeys.commentState("workspace", "story");

const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
};

const setup = (initial: StoryCommentState = state) => {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: Infinity },
      mutations: { retry: false },
    },
  });
  client.setQueryData(queryKey, initial);
  const wrapper = ({ children }: PropsWithChildren) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  const cached = () => client.getQueryData<StoryCommentState>(queryKey);
  return { client, wrapper, cached };
};

beforeEach(() => {
  jest.resetAllMocks();
});

describe("optimistic comment reactions", () => {
  it("shows the first reaction before saving and reconciles the server count", async () => {
    const { wrapper, cached } = setup({ ...state, comments: [] });
    const request = deferred<CommentState>();
    jest.mocked(setCommentReaction).mockReturnValue(request.promise);
    const { result } = renderHook(
      () => useCommentCollaboration("story", "comment"),
      { wrapper },
    );

    act(() => {
      result.current.reaction.mutate({ emoji: "😀", reacted: true });
    });
    await waitFor(() => {
      expect(cached()?.comments[0]?.reactions).toEqual([
        { emoji: "😀", count: 1, reacted: true },
      ]);
    });
    expect(result.current.reaction.isPending).toBe(true);

    act(() => {
      request.resolve({
        ...entry,
        reactions: [{ emoji: "😀", count: 3, reacted: true }],
      });
    });
    await waitFor(() => {
      expect(result.current.reaction.isSuccess).toBe(true);
    });
    expect(cached()?.comments[0].reactions).toEqual([
      { emoji: "😀", count: 3, reacted: true },
    ]);
  });

  it.each([
    {
      count: 2,
      wasReacted: false,
      reacted: true,
      expected: [{ emoji: "👍", count: 3, reacted: true }],
    },
    {
      count: 2,
      wasReacted: true,
      reacted: false,
      expected: [{ emoji: "👍", count: 1, reacted: false }],
    },
    { count: 1, wasReacted: true, reacted: false, expected: [] },
    {
      count: 2,
      wasReacted: true,
      reacted: true,
      expected: [{ emoji: "👍", count: 2, reacted: true }],
    },
    {
      count: 2,
      wasReacted: false,
      reacted: false,
      expected: [{ emoji: "👍", count: 2, reacted: false }],
    },
  ])(
    "updates count $count from reacted=$wasReacted to $reacted immediately and restores failed changes",
    async ({ count, wasReacted, reacted, expected }) => {
      const initial: StoryCommentState = {
        ...state,
        comments: [
          {
            ...entry,
            reactions: [{ emoji: "👍", count, reacted: wasReacted }],
          },
        ],
      };
      const { wrapper, cached } = setup(initial);
      const request = deferred<CommentState>();
      jest.mocked(setCommentReaction).mockReturnValue(request.promise);
      const { result } = renderHook(
        () => useCommentCollaboration("story", "comment"),
        { wrapper },
      );

      act(() => {
        result.current.reaction.mutate({ emoji: "👍", reacted });
      });
      await waitFor(() => {
        expect(setCommentReaction).toHaveBeenCalled();
      });
      expect(cached()?.comments[0].reactions).toEqual(expected);
      act(() => {
        request.reject(new Error("offline"));
      });
      await waitFor(() => {
        expect(result.current.reaction.isError).toBe(true);
      });
      expect(cached()).toEqual(initial);
      expect(toast.error).toHaveBeenCalledTimes(1);
    },
  );

  it("rolls back only the failed emoji and preserves new receipts, resolution and other reactions", async () => {
    const { client, wrapper, cached } = setup();
    const request = deferred<CommentState>();
    jest.mocked(setCommentReaction).mockReturnValue(request.promise);
    const { result } = renderHook(
      () => useCommentCollaboration("story", "comment"),
      { wrapper },
    );
    act(() => {
      result.current.reaction.mutate({ emoji: "😀", reacted: true });
    });
    await waitFor(() => {
      expect(cached()?.comments[0].reactions).toHaveLength(1);
    });

    const other: CommentState = {
      ...entry,
      commentId: "other",
      reactions: [{ emoji: "🎉", count: 4, reacted: true }],
    };
    client.setQueryData<StoryCommentState>(queryKey, (current) => ({
      ...current!,
      lastReadAt: "2026-10-16T13:00:00Z",
      threadReads: { other: "2026-10-16T13:00:00Z" },
      comments: [
        {
          ...current!.comments[0],
          resolvedAt: "2026-10-16T13:00:00Z",
          resolvedBy: "teammate",
          reactions: [
            ...current!.comments[0].reactions,
            { emoji: "👀", count: 2, reacted: true },
          ],
        },
        other,
      ],
    }));
    act(() => {
      request.reject(new Error("offline"));
    });
    await waitFor(() => {
      expect(result.current.reaction.isError).toBe(true);
    });
    expect(cached()).toMatchObject({
      lastReadAt: "2026-10-16T13:00:00Z",
      threadReads: { other: "2026-10-16T13:00:00Z" },
      comments: [
        {
          resolvedAt: "2026-10-16T13:00:00Z",
          resolvedBy: "teammate",
          reactions: [{ emoji: "👀", count: 2, reacted: true }],
        },
        other,
      ],
    });
  });

  it("removes a newly introduced empty state entry after failure", async () => {
    const initial = { ...state, comments: [] };
    const { wrapper, cached } = setup(initial);
    jest.mocked(setCommentReaction).mockRejectedValue(new Error("offline"));
    const { result } = renderHook(
      () => useCommentCollaboration("story", "comment"),
      { wrapper },
    );
    act(() => {
      result.current.reaction.mutate({ emoji: "😀", reacted: true });
    });
    await waitFor(() => {
      expect(result.current.reaction.isError).toBe(true);
    });
    expect(cached()).toEqual(initial);
  });

  it("does not invent a read-state snapshot when the cache has not loaded", async () => {
    const { client, wrapper, cached } = setup();
    client.removeQueries({ queryKey });
    jest.mocked(setCommentReaction).mockResolvedValue({
      ...entry,
      reactions: [{ emoji: "😀", count: 1, reacted: true }],
    });
    const { result } = renderHook(
      () => useCommentCollaboration("story", "comment"),
      { wrapper },
    );
    act(() => {
      result.current.reaction.mutate({ emoji: "😀", reacted: true });
    });
    await waitFor(() => {
      expect(result.current.reaction.isSuccess).toBe(true);
    });
    expect(cached()).toBeUndefined();
  });

  it("cancels an older in-flight read before applying the reaction", async () => {
    const { client, wrapper, cached } = setup();
    const oldRead = deferred<StoryCommentState>();
    const read = client
      .fetchQuery({ queryKey, queryFn: () => oldRead.promise })
      .catch(() => undefined);
    const request = deferred<CommentState>();
    jest.mocked(setCommentReaction).mockReturnValue(request.promise);
    const { result } = renderHook(
      () => useCommentCollaboration("story", "comment"),
      { wrapper },
    );
    act(() => {
      result.current.reaction.mutate({ emoji: "😀", reacted: true });
    });
    await waitFor(() => {
      expect(cached()?.comments[0].reactions).toHaveLength(1);
    });
    await act(async () => {
      oldRead.resolve(state);
      await read;
    });
    expect(cached()?.comments[0].reactions).toEqual([
      { emoji: "😀", count: 1, reacted: true },
    ]);
    act(() => {
      request.resolve({
        ...entry,
        reactions: [{ emoji: "😀", count: 1, reacted: true }],
      });
    });
    await waitFor(() => {
      expect(result.current.reaction.isSuccess).toBe(true);
    });
  });

  it("keeps overlapping reactions visible and refetches after the last request settles", async () => {
    const { client, wrapper, cached } = setup();
    jest.mocked(getCommentState).mockResolvedValue(state);
    const first = deferred<CommentState>();
    const second = deferred<CommentState>();
    jest
      .mocked(setCommentReaction)
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    const { result } = renderHook(
      () => ({
        feed: useStoryCommentState("story"),
        first: useCommentCollaboration("story", "comment"),
        second: useCommentCollaboration("story", "comment"),
      }),
      { wrapper },
    );
    await waitFor(() => {
      expect(result.current.feed.isFetching).toBe(false);
    });
    expect(getCommentState).toHaveBeenCalledTimes(1);

    act(() => {
      result.current.first.reaction.mutate({ emoji: "😀", reacted: true });
    });
    await waitFor(() => {
      expect(cached()?.comments[0].reactions).toHaveLength(1);
    });
    act(() => {
      result.current.second.reaction.mutate({ emoji: "👀", reacted: true });
    });
    await waitFor(() => {
      expect(cached()?.comments[0].reactions).toHaveLength(2);
    });
    await act(async () => {
      await client.invalidateQueries({ queryKey });
      await result.current.feed.refetch();
    });
    expect(getCommentState).toHaveBeenCalledTimes(1);

    act(() => {
      first.resolve({
        ...entry,
        reactions: [{ emoji: "😀", count: 2, reacted: true }],
      });
    });
    await waitFor(() => {
      expect(result.current.first.reaction.isSuccess).toBe(true);
    });
    expect(getCommentState).toHaveBeenCalledTimes(1);
    expect(cached()?.comments[0].reactions).toEqual([
      { emoji: "😀", count: 2, reacted: true },
      { emoji: "👀", count: 1, reacted: true },
    ]);

    const saved: CommentState = {
      ...entry,
      reactions: [
        { emoji: "😀", count: 2, reacted: true },
        { emoji: "👀", count: 3, reacted: true },
      ],
    };
    jest
      .mocked(getCommentState)
      .mockResolvedValue({ ...state, comments: [saved] });
    act(() => {
      second.resolve(saved);
    });
    await waitFor(() => {
      expect(getCommentState).toHaveBeenCalledTimes(2);
    });
    await waitFor(() => {
      expect(cached()?.comments[0]).toEqual(saved);
    });
  });
});
