/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, renderHook } from "@testing-library/react";
import type { Comment } from "@/types/comment";
import { useCommentThread } from "./comment-collaboration";
import { useCommentFeed } from "./use-comment-feed";

jest.mock("./comment-maya", () => ({
  useStoryMayaRuns: () => ({ data: { enabled: false, runs: [] } }),
  useMayaRunRefresh: () => undefined,
}));

const REPLY_ID = "20000000-0000-4000-8000-000000000001";
let mockSearch = new URLSearchParams();
let mockComments: Comment[] = [];
let mockCommentsPending = false;
let mockThread: Comment | undefined;
let mockThreadError = false;
const mockState = {
  lastReadAt: "2026-10-15T00:00:00Z",
  observedAt: "2026-10-16T13:00:00Z",
  comments: [],
  unreadCount: 1,
  latestCommentId: REPLY_ID,
};
jest.mock("next/navigation", () => ({ useSearchParams: () => mockSearch }));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "me" } } }),
}));
jest.mock("./story-comments", () => ({
  useStoryCommentsInfinite: () => ({
    isPending: mockCommentsPending,
    data: { pages: [{ comments: mockComments }] },
  }),
}));
jest.mock("./comment-collaboration", () => ({
  useStoryCommentState: () => ({ data: mockState }),
  useCommentThread: jest.fn((_storyId: string, commentId: string | null) =>
    commentId ? { data: mockThread, isError: mockThreadError } : {},
  ),
}));
jest.mock("./comment-read-tracking", () => ({
  useCommentReadTracking: () => ({
    containerRef: { current: null },
    lastReadAt: mockState.lastReadAt,
    hadUnread: true,
  }),
}));

const root: Comment = {
  id: "root",
  parentId: null,
  storyId: "story",
  userId: "author",
  user: {
    id: "author",
    username: "author",
    fullName: "Author",
    isActive: true,
    isSystem: false,
    avatarUrl: null,
  },
  comment: "<p>Original root</p>",
  subComments: [],
  createdAt: "2026-10-15T12:00:00Z",
  updatedAt: "2026-10-15T12:00:00Z",
};
const reply: Comment = {
  ...root,
  id: REPLY_ID,
  parentId: root.id,
  createdAt: "2026-10-16T12:00:00Z",
  comment: "<p>New reply</p>",
};

beforeEach(() => {
  jest.clearAllMocks();
  mockSearch = new URLSearchParams();
  mockComments = [root];
  mockCommentsPending = false;
  mockThread = { ...root, subComments: [reply] };
  mockThreadError = false;
});

describe("comment feed navigation", () => {
  it("waits for the initial feed before scrolling to a linked reply", () => {
    mockSearch.set("comment", REPLY_ID);
    mockCommentsPending = true;
    mockComments = [];
    const target = document.createElement("div");
    target.id = `comment-${REPLY_ID}`;
    target.tabIndex = -1;
    const scroll = jest.fn();
    target.scrollIntoView = scroll;
    document.body.append(target);
    const { rerender } = renderHook(() => useCommentFeed("story"));
    expect(scroll).not.toHaveBeenCalled();
    mockCommentsPending = false;
    mockComments = [mockThread!];
    rerender();
    expect(scroll).toHaveBeenCalledTimes(1);
    expect(document.activeElement).toBe(target);
    rerender();
    expect(scroll).toHaveBeenCalledTimes(1);
    target.remove();
  });
  it("keeps optimistic nesting stable after the API refreshes flat descendants", () => {
    const grandchild = { ...reply, id: "grandchild", parentId: reply.id };
    mockComments = [
      { ...root, subComments: [{ ...reply, subComments: [grandchild] }] },
    ];
    const { result, rerender } = renderHook(() => useCommentFeed("story"));
    const optimisticTree = result.current.comments;
    mockComments = [{ ...root, subComments: [reply, grandchild] }];
    rerender();
    expect(result.current.comments).toEqual(optimisticTree);
    expect(result.current.comments[0].subComments[0].subComments).toEqual([
      grandchild,
    ]);
    expect(mockComments[0].subComments).toEqual([reply, grandchild]);
  });

  it("normalizes a linked off-page third-level reply before displaying it", () => {
    const grandchildId = "20000000-0000-4000-8000-000000000002";
    const grandchild = { ...reply, id: grandchildId, parentId: reply.id };
    mockSearch.set("comment", grandchildId);
    mockComments = [];
    mockThread = { ...root, subComments: [reply, grandchild] };
    const { result } = renderHook(() => useCommentFeed("story"));
    expect(useCommentThread).toHaveBeenLastCalledWith("story", grandchildId);
    expect(result.current.comments[0].subComments[0].subComments).toEqual([
      grandchild,
    ]);
  });

  it("uses the fetched thread when a new target reply is absent from its cached root", () => {
    mockSearch.set("comment", REPLY_ID);
    const { result } = renderHook(() => useCommentFeed("story"));
    expect(useCommentThread).toHaveBeenLastCalledWith("story", REPLY_ID);
    expect(result.current.comments).toHaveLength(1);
    expect(result.current.comments[0].subComments).toEqual([reply]);
  });
  it("loads the server's latest reply even if its root is outside the current page", () => {
    mockComments = [{ ...root, id: "different-root" }];
    const { result } = renderHook(() => useCommentFeed("story"));
    act(() => {
      result.current.jumpToLatest();
    });
    expect(useCommentThread).toHaveBeenLastCalledWith("story", REPLY_ID);
    expect(result.current.targetId).toBe(REPLY_ID);
    expect(result.current.comments.map((comment) => comment.id)).toEqual([
      "root",
      "different-root",
    ]);
  });
  it("does not display stale linked content after a permission or deletion error", () => {
    mockSearch.set("comment", REPLY_ID);
    mockComments = [];
    mockThreadError = true;
    const { result } = renderHook(() => useCommentFeed("story"));
    expect(result.current.comments).toEqual([]);
  });
  it("uses an already loaded target without fetching a duplicate thread", () => {
    mockSearch.set("comment", REPLY_ID);
    mockComments = [mockThread!];
    const { result } = renderHook(() => useCommentFeed("story"));
    expect(useCommentThread).toHaveBeenLastCalledWith("story", null);
    expect(result.current.comments).toHaveLength(1);
  });
});

describe("comment target highlight", () => {
  const targets: HTMLElement[] = [];
  const mountTarget = (id: string) => {
    const target = document.createElement("div");
    target.id = `comment-${id}`;
    target.tabIndex = -1;
    const scroll = jest.fn();
    target.scrollIntoView = scroll;
    document.body.append(target);
    targets.push(target);
    return { target, scroll };
  };
  const advance = (milliseconds: number) => {
    act(() => {
      jest.advanceTimersByTime(milliseconds);
    });
  };

  beforeEach(() => {
    jest.useFakeTimers();
    mockSearch.set("comment", REPLY_ID);
  });
  afterEach(() => {
    targets.splice(0).forEach((target) => {
      target.remove();
    });
    jest.useRealTimers();
  });

  it("starts its five seconds only after the initial feed and target are revealed", () => {
    mockCommentsPending = true;
    const { result, rerender } = renderHook(() => useCommentFeed("story"));
    advance(6_000);
    expect(result.current.highlightedId).toBeNull();

    mockCommentsPending = false;
    rerender();
    advance(6_000);
    expect(result.current.highlightedId).toBeNull();

    const { target, scroll } = mountTarget(REPLY_ID);
    rerender();
    expect(document.activeElement).toBe(target);
    expect(scroll).toHaveBeenCalledWith({
      block: "center",
      behavior: "instant",
    });
    expect(result.current.highlightedId).toBe(REPLY_ID);
    advance(4_999);
    expect(result.current.highlightedId).toBe(REPLY_ID);
    advance(1);
    expect(result.current.highlightedId).toBeNull();
    expect(result.current.targetId).toBe(REPLY_ID);
    expect(result.current.comments[0].subComments).toEqual([reply]);
  });

  it("does not restart the highlight or focus when feed data refreshes", () => {
    const { scroll } = mountTarget(REPLY_ID);
    const { result, rerender } = renderHook(() => useCommentFeed("story"));
    advance(4_000);
    mockComments = [{ ...root }];
    rerender();
    advance(1_000);
    expect(result.current.highlightedId).toBeNull();
    rerender();
    expect(result.current.highlightedId).toBeNull();
    expect(scroll).toHaveBeenCalledTimes(1);
  });

  it("cancels the prior timer and highlights again when returning from an unavailable target", () => {
    const { scroll } = mountTarget(REPLY_ID);
    const { result, rerender } = renderHook(() => useCommentFeed("story"));
    advance(4_000);
    mockSearch = new URLSearchParams(
      "comment=20000000-0000-4000-8000-000000000002",
    );
    rerender();
    expect(result.current.highlightedId).toBeNull();
    expect(jest.getTimerCount()).toBe(0);

    mockSearch = new URLSearchParams(`comment=${REPLY_ID}`);
    rerender();
    advance(1_000);
    expect(result.current.highlightedId).toBe(REPLY_ID);
    expect(scroll).toHaveBeenCalledTimes(2);
    advance(4_000);
    expect(result.current.highlightedId).toBeNull();
  });

  it("restarts the full highlight period for repeated jumps to the same reply", () => {
    const { scroll } = mountTarget(REPLY_ID);
    const { result } = renderHook(() => useCommentFeed("story"));
    const initialReveal = result.current.revealKey;
    advance(4_000);
    act(() => {
      result.current.jumpToLatest();
    });
    expect(result.current.revealKey).not.toBe(initialReveal);
    expect(scroll).toHaveBeenCalledTimes(2);
    advance(1_000);
    expect(result.current.highlightedId).toBe(REPLY_ID);
    advance(4_000);
    expect(result.current.highlightedId).toBeNull();
    act(() => {
      result.current.jumpToLatest();
    });
    expect(result.current.highlightedId).toBe(REPLY_ID);
  });

  it("cleans up the timeout when the feed unmounts", () => {
    mountTarget(REPLY_ID);
    const { unmount } = renderHook(() => useCommentFeed("story"));
    expect(jest.getTimerCount()).toBe(1);
    unmount();
    expect(jest.getTimerCount()).toBe(0);
  });
});
