/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { Comment } from "@/types/comment";
import { storyKeys } from "@/shared/story/cache-keys";
import { getStoryComments } from "../queries/get-comments";
import {
  getCommentState,
  getCommentThread,
} from "../queries/comment-collaboration";
import { useCommentFeed } from "./use-comment-feed";

const REPLY_ID = "20000000-0000-5000-8000-000000000001";
let mockSearch = new URLSearchParams();

jest.mock("next/navigation", () => ({ useSearchParams: () => mockSearch }));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "me" } } }),
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("../queries/get-comments", () => ({ getStoryComments: jest.fn() }));
jest.mock("../queries/comment-collaboration", () => ({
  getCommentState: jest.fn(),
  getCommentThread: jest.fn(),
}));
jest.mock("./comment-maya", () => ({
  useStoryMayaRuns: () => ({ data: { enabled: false, runs: [] } }),
  useMayaRunRefresh: () => undefined,
}));
jest.mock("./comment-read-tracking", () => ({
  useCommentReadTracking: () => ({
    containerRef: { current: null },
    lastReadAt: null,
  }),
}));

const root: Comment = {
  id: "10000000-0000-4000-8000-000000000001",
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
  comment: "<p>Maya, can you explain this task?</p>",
  subComments: [],
  createdAt: "2026-10-17T12:00:00Z",
  updatedAt: "2026-10-17T12:00:00Z",
};
const reply: Comment = {
  ...root,
  id: REPLY_ID,
  parentId: root.id,
  userId: "maya",
  user: { ...root.user, id: "maya", username: "maya", isSystem: true },
  comment: "<p>Here is the explanation.</p>",
};
const thread = { ...root, subComments: [reply] };
const clients: QueryClient[] = [];

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(getCommentThread).mockReset();
  mockSearch = new URLSearchParams({ comment: REPLY_ID });
  jest.mocked(getStoryComments).mockResolvedValue({
    comments: [root],
    pagination: { page: 1, pageSize: 20, hasMore: false, nextPage: 0 },
  });
  jest.mocked(getCommentState).mockResolvedValue({
    lastReadAt: null,
    observedAt: "2026-10-17T12:00:00Z",
    unreadCount: 1,
    latestCommentId: REPLY_ID,
    comments: [],
  });
});

afterEach(() => {
  clients.splice(0).forEach((client) => {
    client.clear();
  });
});

const renderCachedThread = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  clients.push(client);
  client.setQueryData(
    storyKeys.commentThread("acme", "story", REPLY_ID),
    thread,
  );
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return renderHook(() => useCommentFeed("story"), { wrapper });
};

describe("linked Maya reply refresh", () => {
  it.each([
    new TypeError("Failed to fetch"),
    Object.assign(new Error("Server unavailable"), { status: 503 }),
    Object.assign(new Error("Too many requests"), { status: 429 }),
  ])(
    "retains an already displayed reply when a background refresh fails temporarily: %s",
    async (error) => {
      let rejectRefresh!: (error: Error) => void;
      jest.mocked(getCommentThread).mockReturnValue(
        new Promise<Comment>((_resolve, reject) => {
          rejectRefresh = reject;
        }),
      );
      const { result } = renderCachedThread();
      expect(result.current.comments[0].subComments).toEqual([reply]);
      await waitFor(() => {
        expect(getCommentThread).toHaveBeenCalledWith("story", REPLY_ID, {
          workspaceSlug: "acme",
        });
      });
      await act(async () => {
        rejectRefresh(error);
      });
      await waitFor(() => {
        expect(result.current.threadQuery.isError).toBe(true);
      });
      expect(result.current.comments[0].subComments).toEqual([reply]);
    },
  );

  it.each([401, 403, 404, 410])(
    "hides cached replies when the server confirms they are unavailable (%s)",
    async (status) => {
      let rejectRefresh!: (error: Error) => void;
      jest.mocked(getCommentThread).mockReturnValue(
        new Promise<Comment>((_resolve, reject) => {
          rejectRefresh = reject;
        }),
      );
      const { result } = renderCachedThread();
      expect(result.current.comments[0].subComments).toEqual([reply]);
      await waitFor(() => {
        expect(getCommentThread).toHaveBeenCalled();
      });
      await act(async () => {
        rejectRefresh(Object.assign(new Error("Unavailable"), { status }));
      });
      await waitFor(() => {
        expect(result.current.threadQuery.isError).toBe(true);
      });
      expect(result.current.comments[0].subComments).toEqual([]);
    },
  );

  it("reconciles the retained reply with the server after a successful retry", async () => {
    jest
      .mocked(getCommentThread)
      .mockRejectedValueOnce(new TypeError("Failed to fetch"));
    const { result } = renderCachedThread();
    await waitFor(() => {
      expect(result.current.threadQuery.isError).toBe(true);
    });
    expect(result.current.comments[0].subComments).toEqual([reply]);
    const updatedReply = {
      ...reply,
      comment: "<p>Here is the updated explanation.</p>",
      updatedAt: "2026-10-17T12:01:00Z",
    };
    jest
      .mocked(getCommentThread)
      .mockResolvedValue({ ...root, subComments: [updatedReply] });
    await act(async () => {
      await result.current.threadQuery.refetch();
    });
    await waitFor(() => {
      expect(result.current.threadQuery.isError).toBe(false);
    });
    expect(result.current.comments[0].subComments).toEqual([updatedReply]);
  });

  it("never restores revoked cached content when a subsequent retry fails temporarily", async () => {
    jest
      .mocked(getCommentThread)
      .mockRejectedValueOnce(
        Object.assign(new Error("Forbidden"), { status: 403 }),
      );
    const { result } = renderCachedThread();
    await waitFor(() => {
      expect(result.current.threadQuery.isError).toBe(true);
    });
    expect(result.current.comments[0].subComments).toEqual([]);

    jest
      .mocked(getCommentThread)
      .mockRejectedValueOnce(
        Object.assign(new Error("Service unavailable"), { status: 503 }),
      );
    await act(async () => {
      await result.current.threadQuery.refetch();
    });
    expect(result.current.comments[0].subComments).toEqual([]);

    jest.mocked(getCommentThread).mockResolvedValueOnce(thread);
    await act(async () => {
      await result.current.threadQuery.refetch();
    });
    await waitFor(() => {
      expect(result.current.threadQuery.isError).toBe(false);
    });
    expect(result.current.comments[0].subComments).toEqual([reply]);
  });
});
