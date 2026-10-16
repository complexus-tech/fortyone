import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import type { InfiniteCommentsData } from "@/shared/story/comment-cache";
import type { ApiResponse, Comment } from "@/types";
import { deleteCommentAction } from "@/lib/actions/comments/delete-comment";
import { updateCommentAction } from "@/lib/actions/comments/update-comment";
import { useDeleteCommentMutation } from "@/lib/hooks/delete-comment-mutation";
import { useUpdateCommentMutation } from "@/lib/hooks/update-comment-mutation";
import { storyKeys } from "@/modules/stories/constants";
import {
  findCachedComment,
  insertCachedComment,
} from "@/shared/story/comment-cache";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { retainDesktopComment } from "@/shell/desktop/outbox-projection";
import { commentDesktopStory } from "@/shell/desktop/story-offline";
import { commentStoryAction } from "../actions/comment-story";
import { useCommentStoryMutation } from "./comment-mutation";

jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
  useTerminology: () => ({ getTermDisplay: () => "story" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({
    data: { user: { id: "alice", name: "Alice", username: "alice" } },
  }),
}));
jest.mock("@/shell/desktop/bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("@/shell/desktop/workspace-scope", () => ({
  useDesktopWorkspaceScope: () => ({
    accountId: "alice",
    workspaceId: "workspace",
    workspaceSlug: "acme",
  }),
}));
jest.mock("@/shell/desktop/story-offline", () => ({
  commentDesktopStory: jest.fn(),
  isDesktopPending: (response: { desktopPending?: boolean } | undefined) =>
    response?.desktopPending === true,
}));
jest.mock("@/shell/desktop/outbox-projection", () => ({
  retainDesktopComment: jest.fn(),
}));
jest.mock("../actions/comment-story", () => ({
  commentStoryAction: jest.fn(),
}));
jest.mock("@/lib/actions/comments/update-comment", () => ({
  updateCommentAction: jest.fn(),
}));
jest.mock("@/lib/actions/comments/delete-comment", () => ({
  deleteCommentAction: jest.fn(),
}));

const STORY_ID = "story";
const QUERY_KEY = storyKeys.commentsInfinite("acme", STORY_ID);
const makeComment = (
  id: string,
  overrides: Partial<Comment> = {},
): Comment => ({
  id,
  storyId: STORY_ID,
  parentId: null,
  userId: "alice",
  user: {
    id: "alice",
    username: "alice",
    fullName: "Alice",
    avatarUrl: null,
    isActive: true,
    isSystem: false,
  },
  comment: `<p>${id}</p>`,
  createdAt: "2026-10-01T12:00:00.000Z",
  updatedAt: "2026-10-01T12:00:00.000Z",
  subComments: [],
  ...overrides,
});
const makeData = (...pages: Comment[][]): InfiniteCommentsData => ({
  pageParams: pages.map((_, index) => index + 1),
  pages: pages.map((comments, index) => ({
    comments,
    pagination: {
      page: index + 1,
      pageSize: 20,
      hasMore: index < pages.length - 1,
      nextPage: index + 2,
    },
  })),
});
const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
};
const setup = () => {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: Infinity },
      mutations: { retry: false },
    },
  });
  const Wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return {
    client,
    wrapper: Wrapper,
    cached: () => client.getQueryData<InfiniteCommentsData>(QUERY_KEY),
  };
};

beforeEach(() => {
  jest.resetAllMocks();
  jest.mocked(getDesktopBridge).mockReturnValue(null);
  let sequence = 0;
  Object.defineProperty(crypto, "randomUUID", {
    configurable: true,
    value: jest.fn(
      () => `00000000-0000-4000-8000-${String(++sequence).padStart(12, "0")}`,
    ),
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

it("inserts a reply on a later page after existing replies and replaces its temporary ID on success", async () => {
  const { client, wrapper, cached } = setup();
  const original = makeData(
    [makeComment("latest")],
    [
      makeComment("parent", {
        subComments: [makeComment("older", { parentId: "parent" })],
      }),
    ],
  );
  client.setQueryData(QUERY_KEY, original);
  const response = deferred<ApiResponse<Comment>>();
  jest.mocked(commentStoryAction).mockReturnValue(response.promise);
  const { result } = renderHook(() => useCommentStoryMutation(), { wrapper });
  act(() => {
    result.current.mutate({
      storyId: STORY_ID,
      payload: {
        comment: "<p>New reply</p>",
        parentId: "parent",
        mentions: [],
      },
    });
  });
  await waitFor(() => {
    expect(commentStoryAction).toHaveBeenCalledTimes(1);
  });
  const replies = cached()!.pages[1].comments[0].subComments;
  expect(replies.map((comment) => comment.comment)).toEqual([
    "<p>older</p>",
    "<p>New reply</p>",
  ]);
  expect(replies[1].id).toMatch(/^pending-comment:/);
  expect(cached()!.pages[0]).toEqual(original.pages[0]);
  act(() => {
    response.resolve({
      data: makeComment("saved", {
        parentId: "parent",
        comment: "<p>New reply</p>",
        createdAt: new Date().toISOString(),
      }),
    });
  });
  await waitFor(() => {
    expect(result.current.isSuccess).toBe(true);
  });
  expect(
    cached()!.pages[1].comments[0].subComments.map((comment) => comment.id),
  ).toEqual(["older", "saved"]);
  client.clear();
});

it("rolls back only a failed create while another comment is still being sent", async () => {
  const { client, wrapper, cached } = setup();
  client.setQueryData(QUERY_KEY, makeData([makeComment("existing")]));
  const first = deferred<ApiResponse<Comment>>();
  const second = deferred<ApiResponse<Comment>>();
  jest
    .mocked(commentStoryAction)
    .mockReturnValueOnce(first.promise)
    .mockReturnValueOnce(second.promise);
  const invalidate = jest.spyOn(client, "invalidateQueries");
  const { result } = renderHook(
    () => ({
      first: useCommentStoryMutation(),
      second: useCommentStoryMutation(),
    }),
    { wrapper },
  );
  act(() => {
    result.current.first.mutate({
      storyId: STORY_ID,
      payload: { comment: "First", mentions: [] },
    });
    result.current.second.mutate({
      storyId: STORY_ID,
      payload: { comment: "Second", mentions: [] },
    });
  });
  await waitFor(() => {
    expect(commentStoryAction).toHaveBeenCalledTimes(2);
  });
  const temporaryIds = cached()!
    .pages[0].comments.filter((comment) =>
      comment.id.startsWith("pending-comment:"),
    )
    .map((comment) => comment.id);
  expect(new Set(temporaryIds).size).toBe(2);
  act(() => {
    first.resolve({ error: { message: "First failed" } });
  });
  await waitFor(() => {
    expect(result.current.first.isError).toBe(true);
  });
  expect(cached()!.pages[0].comments.map((comment) => comment.comment)).toEqual(
    ["Second", "<p>existing</p>"],
  );
  expect(invalidate).not.toHaveBeenCalled();
  act(() => {
    second.resolve({
      data: makeComment("second-saved", {
        comment: "Second",
        createdAt: new Date().toISOString(),
      }),
    });
  });
  await waitFor(() => {
    expect(result.current.second.isSuccess).toBe(true);
  });
  expect(invalidate).toHaveBeenCalledTimes(3);
  expect(invalidate).toHaveBeenCalledWith({ queryKey: QUERY_KEY, exact: true });
  expect(invalidate).toHaveBeenCalledWith({
    queryKey: storyKeys.commentThreads("acme", STORY_ID),
  });
  expect(invalidate).toHaveBeenCalledWith({
    queryKey: storyKeys.commentState("acme", STORY_ID),
    exact: true,
  });
  expect(cached()!.pages[0].comments.map((comment) => comment.id)).toEqual([
    "second-saved",
    "existing",
  ]);
  client.clear();
});

it("updates deeply nested replies and restores only their content when the action returns an error", async () => {
  const { client, wrapper, cached } = setup();
  client.setQueryData(
    QUERY_KEY,
    makeData(
      [],
      [
        makeComment("parent", {
          subComments: [
            makeComment("reply", {
              parentId: "parent",
              subComments: [makeComment("nested", { parentId: "reply" })],
            }),
          ],
        }),
      ],
    ),
  );
  const response = deferred<ApiResponse<null>>();
  jest.mocked(updateCommentAction).mockReturnValue(response.promise);
  const onSuccess = jest.fn();
  const { result } = renderHook(() => useUpdateCommentMutation(), { wrapper });
  act(() => {
    result.current.mutate(
      {
        storyId: STORY_ID,
        commentId: "nested",
        payload: { content: "Edited", mentions: [] },
      },
      { onSuccess },
    );
  });
  await waitFor(() => {
    expect(updateCommentAction).toHaveBeenCalledTimes(1);
  });
  expect(findCachedComment(cached(), "nested")!.comment.comment).toBe("Edited");
  client.setQueryData<InfiniteCommentsData>(QUERY_KEY, (current) =>
    insertCachedComment(current, makeComment("unrelated")),
  );
  act(() => {
    response.resolve({ error: { message: "Edit rejected" } });
  });
  await waitFor(() => {
    expect(result.current.isError).toBe(true);
  });
  expect(findCachedComment(cached(), "nested")!.comment.comment).toBe(
    "<p>nested</p>",
  );
  expect(findCachedComment(cached(), "unrelated")).toBeDefined();
  expect(onSuccess).not.toHaveBeenCalled();
  expect(toast.error).toHaveBeenCalledWith(
    "Failed to update comment",
    expect.objectContaining({ description: "Edit rejected" }),
  );
  client.clear();
});

it("deletes nested replies immediately and restores the failed deletion without losing a concurrent reply", async () => {
  const { client, wrapper, cached } = setup();
  client.setQueryData(
    QUERY_KEY,
    makeData(
      [],
      [
        makeComment("parent", {
          subComments: [makeComment("reply", { parentId: "parent" })],
        }),
      ],
    ),
  );
  const response = deferred<ApiResponse<null>>();
  jest.mocked(deleteCommentAction).mockReturnValue(response.promise);
  const { result } = renderHook(() => useDeleteCommentMutation(), { wrapper });
  act(() => {
    result.current.mutate({ storyId: STORY_ID, commentId: "reply" });
  });
  await waitFor(() => {
    expect(deleteCommentAction).toHaveBeenCalledTimes(1);
  });
  expect(findCachedComment(cached(), "reply")).toBeUndefined();
  client.setQueryData<InfiniteCommentsData>(QUERY_KEY, (current) =>
    insertCachedComment(
      current,
      makeComment("new-reply", {
        parentId: "parent",
        createdAt: "2026-10-02T12:00:00.000Z",
      }),
    ),
  );
  act(() => {
    response.resolve({ error: { message: "Delete rejected" } });
  });
  await waitFor(() => {
    expect(result.current.isError).toBe(true);
  });
  expect(
    findCachedComment(cached(), "parent")!.comment.subComments.map(
      (comment) => comment.id,
    ),
  ).toEqual(["reply", "new-reply"]);
  client.clear();
});

it("cancels an older read before applying an optimistic edit", async () => {
  const { client, wrapper, cached } = setup();
  const original = makeData([makeComment("comment")]);
  client.setQueryData(QUERY_KEY, original);
  const read = deferred<InfiniteCommentsData>();
  const pendingRead = client
    .fetchQuery({ queryKey: QUERY_KEY, queryFn: () => read.promise })
    .catch(() => undefined);
  const response = deferred<ApiResponse<null>>();
  jest.mocked(updateCommentAction).mockReturnValue(response.promise);
  const { result } = renderHook(() => useUpdateCommentMutation(), { wrapper });
  act(() => {
    result.current.mutate({
      storyId: STORY_ID,
      commentId: "comment",
      payload: { content: "Edited", mentions: [] },
    });
  });
  await waitFor(() => {
    expect(updateCommentAction).toHaveBeenCalledTimes(1);
  });
  read.resolve(original);
  await pendingRead;
  expect(findCachedComment(cached(), "comment")!.comment.comment).toBe(
    "Edited",
  );
  act(() => {
    response.resolve({ data: null });
  });
  await waitFor(() => {
    expect(result.current.isSuccess).toBe(true);
  });
  client.clear();
});

it("cascades a reply deletion through flattened descendants and restores their original container on failure", async () => {
  const { client, wrapper, cached } = setup();
  const original = makeData([
    makeComment("root", {
      subComments: [
        makeComment("reply", { parentId: "root" }),
        makeComment("descendant", {
          parentId: "reply",
          createdAt: "2026-10-02T12:00:00.000Z",
        }),
        makeComment("other-reply", {
          parentId: "root",
          createdAt: "2026-10-03T12:00:00.000Z",
        }),
      ],
    }),
  ]);
  client.setQueryData(QUERY_KEY, original);
  const otherWorkspace = storyKeys.commentsInfinite("other", STORY_ID);
  client.setQueryData(otherWorkspace, original);
  const response = deferred<ApiResponse<null>>();
  jest.mocked(deleteCommentAction).mockReturnValue(response.promise);
  const { result } = renderHook(() => useDeleteCommentMutation(), { wrapper });
  act(() => {
    result.current.mutate({ storyId: STORY_ID, commentId: "reply" });
  });
  await waitFor(() => {
    expect(deleteCommentAction).toHaveBeenCalledTimes(1);
  });
  expect(
    cached()!.pages[0].comments[0].subComments.map(({ id }) => id),
  ).toEqual(["other-reply"]);
  expect(client.getQueryData(otherWorkspace)).toEqual(original);
  act(() => {
    response.resolve({ error: { message: "Delete rejected" } });
  });
  await waitFor(() => {
    expect(result.current.isError).toBe(true);
  });
  expect(cached()).toEqual(original);
  client.clear();
});

it("leaves queued desktop comments to the durable outbox projection", async () => {
  const { client, wrapper, cached } = setup();
  const original = makeData([makeComment("existing")]);
  client.setQueryData(QUERY_KEY, original);
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue({ enqueueOperation: jest.fn() } as unknown as ReturnType<
      typeof getDesktopBridge
    >);
  const comment = makeComment("desktop-comment");
  jest
    .mocked(commentDesktopStory)
    .mockResolvedValue({ data: comment, desktopPending: true });
  const invalidate = jest.spyOn(client, "invalidateQueries");
  const { result } = renderHook(() => useCommentStoryMutation(), { wrapper });
  act(() => {
    result.current.mutate({
      storyId: STORY_ID,
      payload: { comment: "Offline", mentions: [] },
    });
  });
  await waitFor(() => {
    expect(result.current.isSuccess).toBe(true);
  });
  expect(cached()).toEqual(original);
  expect(retainDesktopComment).toHaveBeenCalledWith(
    client,
    expect.objectContaining({ workspaceSlug: "acme" }),
    comment,
  );
  expect(commentStoryAction).not.toHaveBeenCalled();
  expect(invalidate).not.toHaveBeenCalled();
  client.clear();
});

it("does not defer an online edit's refresh behind a desktop enqueue", async () => {
  const { client, wrapper } = setup();
  client.setQueryData(QUERY_KEY, makeData([makeComment("existing")]));
  jest.mocked(getDesktopBridge).mockReturnValue({
    enqueueOperation: jest.fn(),
  } as unknown as ReturnType<typeof getDesktopBridge>);
  const enqueue = deferred<Awaited<ReturnType<typeof commentDesktopStory>>>();
  const edit = deferred<ApiResponse<null>>();
  jest.mocked(commentDesktopStory).mockReturnValue(enqueue.promise);
  jest.mocked(updateCommentAction).mockReturnValue(edit.promise);
  const invalidate = jest.spyOn(client, "invalidateQueries");
  const { result } = renderHook(
    () => ({
      create: useCommentStoryMutation(),
      edit: useUpdateCommentMutation(),
    }),
    { wrapper },
  );
  act(() => {
    result.current.create.mutate({
      storyId: STORY_ID,
      payload: { comment: "Offline", mentions: [] },
    });
    result.current.edit.mutate({
      storyId: STORY_ID,
      commentId: "existing",
      payload: { content: "Edited", mentions: [] },
    });
  });
  await waitFor(() => {
    expect(updateCommentAction).toHaveBeenCalledTimes(1);
    expect(commentDesktopStory).toHaveBeenCalledTimes(1);
  });
  act(() => {
    edit.resolve({ data: null });
  });
  await waitFor(() => {
    expect(result.current.edit.isSuccess).toBe(true);
  });
  expect(invalidate).toHaveBeenCalledTimes(3);
  expect(invalidate).toHaveBeenCalledWith({ queryKey: QUERY_KEY, exact: true });
  expect(invalidate).toHaveBeenCalledWith({
    queryKey: storyKeys.commentThreads("acme", STORY_ID),
  });
  expect(invalidate).toHaveBeenCalledWith({
    queryKey: storyKeys.commentState("acme", STORY_ID),
    exact: true,
  });
  act(() => {
    enqueue.resolve({
      data: makeComment("desktop-comment"),
      desktopPending: true,
    });
  });
  await waitFor(() => {
    expect(result.current.create.isSuccess).toBe(true);
  });
  expect(invalidate).toHaveBeenCalledTimes(3);
  expect(invalidate).toHaveBeenCalledWith({ queryKey: QUERY_KEY, exact: true });
  expect(invalidate).toHaveBeenCalledWith({
    queryKey: storyKeys.commentThreads("acme", STORY_ID),
  });
  expect(invalidate).toHaveBeenCalledWith({
    queryKey: storyKeys.commentState("acme", STORY_ID),
    exact: true,
  });
  client.clear();
});
