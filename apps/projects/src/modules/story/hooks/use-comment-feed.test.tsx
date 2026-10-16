/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
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
  mockThread = { ...root, subComments: [reply] };
  mockThreadError = false;
});

describe("comment feed navigation", () => {
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
