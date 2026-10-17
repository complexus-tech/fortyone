import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ApiResponse } from "@/types/api-response";
import { documentKeys } from "@/shared/documents/keys";
import {
  createDocumentCommentAction,
  replyToDocumentCommentAction,
  resolveDocumentCommentAction,
} from "./actions";
import { getDocumentComments } from "./queries";
import { useDocumentComments, useDocumentCommentMutations } from "./hooks";
import type { DocumentComment, DocumentCommentThread } from "./types";

jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "me", name: "Me" } } }),
}));
jest.mock("./actions", () => ({
  createDocumentCommentAction: jest.fn(),
  replyToDocumentCommentAction: jest.fn(),
  resolveDocumentCommentAction: jest.fn(),
}));
jest.mock("./queries", () => ({ getDocumentComments: jest.fn() }));

const KEY = documentKeys.comments("acme", "document");
const INPUT = {
  body: "New comment",
  quote: "Text",
  anchorStart: 1,
  anchorEnd: 5,
};
const comment = (id: string): DocumentComment => ({
  id,
  body: id,
  createdBy: "author",
  authorName: "Author",
  authorAvatar: null,
  createdAt: "2026-10-16T12:00:00.000Z",
  updatedAt: "2026-10-16T12:00:00.000Z",
});
const thread = (id: string): DocumentCommentThread => ({
  id,
  documentId: "document",
  quote: "Text",
  anchorStart: 1,
  anchorEnd: 5,
  createdBy: "author",
  resolvedAt: null,
  resolvedBy: null,
  createdAt: "2026-10-16T12:00:00.000Z",
  comments: [comment(`${id}-root`)],
});
const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
};
const setup = (...threads: DocumentCommentThread[]) => {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: Infinity },
      mutations: { retry: false },
    },
  });
  if (threads.length) client.setQueryData(KEY, threads);
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return {
    client,
    wrapper,
    cached: () => client.getQueryData<DocumentCommentThread[]>(KEY),
  };
};

beforeEach(() => {
  jest.resetAllMocks();
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

it("rolls back only a failed create while another create remains pending", async () => {
  const { client, wrapper, cached } = setup(thread("existing"));
  const first = deferred<ApiResponse<DocumentCommentThread>>();
  const second = deferred<ApiResponse<DocumentCommentThread>>();
  jest
    .mocked(createDocumentCommentAction)
    .mockReturnValueOnce(first.promise)
    .mockReturnValueOnce(second.promise);
  const invalidate = jest.spyOn(client, "invalidateQueries");
  const { result } = renderHook(() => useDocumentCommentMutations("document"), {
    wrapper,
  });
  let failed!: Promise<unknown>;
  let saved!: Promise<unknown>;
  act(() => {
    failed = result.current.create
      .mutateAsync({ ...INPUT, draftTarget: "first-selection" })
      .catch((error: unknown) => error);
    saved = result.current.create.mutateAsync({
      ...INPUT,
      body: "Second",
      draftTarget: "second-selection",
    });
  });
  await waitFor(() => {
    expect(createDocumentCommentAction).toHaveBeenCalledTimes(2);
  });
  expect(createDocumentCommentAction).toHaveBeenNthCalledWith(
    1,
    "document",
    INPUT,
    "acme",
  );
  expect(new Set(cached()!.map(({ id }) => id)).size).toBe(3);
  await act(async () => {
    first.resolve({ error: { message: "Rejected" } });
    await failed;
  });
  expect(await failed).toEqual(new Error("Rejected"));
  expect(cached()!.map((entry) => entry.comments[0].body)).toEqual([
    "Second",
    "existing-root",
  ]);
  expect(invalidate).not.toHaveBeenCalled();
  await act(async () => {
    second.resolve({ data: thread("saved") });
    await saved;
  });
  expect(cached()!.map(({ id }) => id)).toEqual(["saved", "existing"]);
  expect(
    invalidate.mock.calls.filter(
      ([options]) => options?.queryKey?.at(-1) === "comments",
    ),
  ).toHaveLength(1);
});

it("keeps a draft locked across composer remounts and releases it after failure", async () => {
  const { wrapper, cached } = setup(thread("existing"));
  const response = deferred<ApiResponse<DocumentCommentThread>>();
  jest
    .mocked(createDocumentCommentAction)
    .mockReturnValueOnce(response.promise);
  const original = renderHook(() => useDocumentCommentMutations("document"), {
    wrapper,
  });
  const payload = { ...INPUT, draftTarget: "selected-passage" };
  let pending!: Promise<unknown>;
  act(() => {
    pending = original.result.current.create
      .mutateAsync(payload)
      .catch((error: unknown) => error);
  });
  await waitFor(() => {
    expect(createDocumentCommentAction).toHaveBeenCalledTimes(1);
  });
  original.unmount();

  const { result } = renderHook(() => useDocumentCommentMutations("document"), {
    wrapper,
  });
  expect(result.current.create.isPending).toBe(false);
  expect(result.current.isCreating).toBe(true);
  expect(result.current.isCreatingDraft("selected-passage")).toBe(true);
  expect(result.current.isCreatingDraft("other-passage")).toBe(false);
  await act(async () => {
    await expect(result.current.create.mutateAsync(payload)).rejects.toThrow(
      "This comment is already being sent",
    );
  });
  expect(createDocumentCommentAction).toHaveBeenCalledTimes(1);
  expect(cached()).toHaveLength(2);

  await act(async () => {
    response.resolve({ error: { message: "Try again" } });
    await pending;
  });
  await waitFor(() => {
    expect(result.current.isCreatingDraft("selected-passage")).toBe(false);
    expect(result.current.isCreating).toBe(false);
  });
  jest.mocked(createDocumentCommentAction).mockResolvedValueOnce({
    data: thread("saved"),
  });
  await act(async () => {
    await result.current.create.mutateAsync(payload);
  });
  expect(createDocumentCommentAction).toHaveBeenCalledTimes(2);
  expect(cached()!.map(({ id }) => id)).toEqual(["saved", "existing"]);
});

it("rolls back only the failed reply while preserving concurrent and externally added replies", async () => {
  const { client, wrapper, cached } = setup(thread("first"), thread("second"));
  const first = deferred<ApiResponse<DocumentComment>>();
  const second = deferred<ApiResponse<DocumentComment>>();
  jest
    .mocked(replyToDocumentCommentAction)
    .mockReturnValueOnce(first.promise)
    .mockReturnValueOnce(second.promise);
  const { result } = renderHook(() => useDocumentCommentMutations("document"), {
    wrapper,
  });
  let failed!: Promise<unknown>;
  let saved!: Promise<unknown>;
  act(() => {
    failed = result.current.reply
      .mutateAsync({ threadId: "first", body: "Failed" })
      .catch((error: unknown) => error);
    saved = result.current.reply.mutateAsync({
      threadId: "second",
      body: "Pending",
    });
  });
  await waitFor(() => {
    expect(replyToDocumentCommentAction).toHaveBeenCalledTimes(2);
  });
  client.setQueryData<DocumentCommentThread[]>(KEY, (current) =>
    current!.map((entry) =>
      entry.id === "first"
        ? { ...entry, comments: [...entry.comments, comment("external")] }
        : entry,
    ),
  );
  await act(async () => {
    first.resolve({ error: { message: "No access" } });
    await failed;
  });
  expect(cached()![0].comments.map(({ id }) => id)).toEqual([
    "first-root",
    "external",
  ]);
  expect(cached()![1].comments.map(({ body }) => body)).toEqual([
    "second-root",
    "Pending",
  ]);
  await act(async () => {
    second.resolve({ data: comment("saved-reply") });
    await saved;
  });
  expect(cached()![1].comments.map(({ id }) => id)).toEqual([
    "second-root",
    "saved-reply",
  ]);
});

it("restores only resolution fields after failure and keeps added comments", async () => {
  const { client, wrapper, cached } = setup(thread("first"), thread("second"));
  const response = deferred<ApiResponse<null>>();
  jest.mocked(resolveDocumentCommentAction).mockReturnValue(response.promise);
  const { result } = renderHook(() => useDocumentCommentMutations("document"), {
    wrapper,
  });
  let pending!: Promise<unknown>;
  act(() => {
    pending = result.current.resolve
      .mutateAsync({ threadId: "first", resolved: true })
      .catch((error: unknown) => error);
  });
  await waitFor(() => {
    expect(resolveDocumentCommentAction).toHaveBeenCalledTimes(1);
  });
  expect(cached()![0].resolvedBy).toBe("me");
  client.setQueryData<DocumentCommentThread[]>(KEY, (current) =>
    current!.map((entry) =>
      entry.id === "first"
        ? { ...entry, comments: [...entry.comments, comment("external")] }
        : { ...entry, resolvedAt: "2026-10-17T12:00:00Z", resolvedBy: "other" },
    ),
  );
  await act(async () => {
    response.resolve({ error: { message: "Failed" } });
    await pending;
  });
  expect(cached()![0].resolvedAt).toBeNull();
  expect(cached()![0].resolvedBy).toBeNull();
  expect(cached()![0].comments.map(({ id }) => id)).toEqual([
    "first-root",
    "external",
  ]);
  expect(cached()![1].resolvedBy).toBe("other");
});

it("does not overwrite a newer resolution when its own request fails", async () => {
  const { client, wrapper, cached } = setup(thread("first"));
  const response = deferred<ApiResponse<null>>();
  jest.mocked(resolveDocumentCommentAction).mockReturnValue(response.promise);
  const { result } = renderHook(() => useDocumentCommentMutations("document"), {
    wrapper,
  });
  let pending!: Promise<unknown>;
  act(() => {
    pending = result.current.resolve
      .mutateAsync({ threadId: "first", resolved: true })
      .catch((error: unknown) => error);
  });
  await waitFor(() => {
    expect(resolveDocumentCommentAction).toHaveBeenCalledTimes(1);
  });
  client.setQueryData(KEY, [
    {
      ...cached()![0],
      resolvedAt: "2026-10-18T12:00:00Z",
      resolvedBy: "other",
    },
  ]);
  await act(async () => {
    response.resolve({ error: { message: "Failed" } });
    await pending;
  });
  expect(cached()![0].resolvedBy).toBe("other");
});

it("blocks writes to optimistic threads and competing writes across hook instances", async () => {
  const { wrapper, cached } = setup(thread("first"));
  const response = deferred<ApiResponse<DocumentComment>>();
  jest.mocked(replyToDocumentCommentAction).mockReturnValue(response.promise);
  jest.mocked(resolveDocumentCommentAction).mockResolvedValue({ data: null });
  const { result } = renderHook(
    () => ({
      first: useDocumentCommentMutations("document"),
      second: useDocumentCommentMutations("document"),
    }),
    { wrapper },
  );
  let reply!: Promise<unknown>;
  let competing!: Promise<unknown>;
  act(() => {
    reply = result.current.first.reply.mutateAsync({
      threadId: "first",
      body: "Reply",
    });
    competing = result.current.second.resolve
      .mutateAsync({ threadId: "first", resolved: true })
      .catch((error: unknown) => error);
  });
  await waitFor(() => {
    expect(replyToDocumentCommentAction).toHaveBeenCalledTimes(1);
  });
  expect(await competing).toEqual(
    new Error("This thread is already being updated"),
  );
  expect(resolveDocumentCommentAction).not.toHaveBeenCalled();
  await waitFor(() => {
    expect(result.current.second.isThreadPending("first")).toBe(true);
  });
  expect(result.current.first.isThreadPending("optimistic-thread-new")).toBe(
    true,
  );
  await act(async () => {
    await expect(
      result.current.second.reply.mutateAsync({
        threadId: "optimistic-thread-new",
        body: "Bad target",
      }),
    ).rejects.toThrow("finish saving");
    await expect(
      result.current.second.resolve.mutateAsync({
        threadId: "optimistic-thread-new",
        resolved: true,
      }),
    ).rejects.toThrow("finish saving");
  });
  expect(replyToDocumentCommentAction).toHaveBeenCalledTimes(1);
  expect(cached()![0].resolvedAt).toBeNull();
  await act(async () => {
    response.resolve({ data: comment("saved") });
    await reply;
  });
  await waitFor(() => {
    expect(result.current.second.isThreadPending("first")).toBe(false);
  });
  await act(async () => {
    await result.current.second.resolve.mutateAsync({
      threadId: "first",
      resolved: true,
    });
  });
  expect(resolveDocumentCommentAction).toHaveBeenCalledTimes(1);
});

it("reconciles canonical IDs without dropping fresher cached content or duplicating replies", async () => {
  const { client, wrapper, cached } = setup(thread("first"));
  const created = deferred<ApiResponse<DocumentCommentThread>>();
  jest.mocked(createDocumentCommentAction).mockReturnValue(created.promise);
  const { result } = renderHook(() => useDocumentCommentMutations("document"), {
    wrapper,
  });
  let pending!: Promise<unknown>;
  act(() => {
    pending = result.current.create.mutateAsync(INPUT);
  });
  await waitFor(() => {
    expect(createDocumentCommentAction).toHaveBeenCalledTimes(1);
  });
  const canonical = {
    ...thread("saved"),
    comments: [comment("saved-root"), comment("newer")],
  };
  client.setQueryData(KEY, [...cached()!, canonical]);
  await act(async () => {
    created.resolve({ data: thread("saved") });
    await pending;
  });
  expect(cached()!.map(({ id }) => id)).toEqual(["first", "saved"]);
  expect(cached()![1].comments).toEqual(canonical.comments);

  const replied = deferred<ApiResponse<DocumentComment>>();
  jest.mocked(replyToDocumentCommentAction).mockReturnValue(replied.promise);
  act(() => {
    pending = result.current.reply.mutateAsync({
      threadId: "first",
      body: "Reply",
    });
  });
  await waitFor(() => {
    expect(replyToDocumentCommentAction).toHaveBeenCalledTimes(1);
  });
  client.setQueryData<DocumentCommentThread[]>(KEY, (current) =>
    current!.map((entry) =>
      entry.id === "first"
        ? { ...entry, comments: [...entry.comments, comment("saved-reply")] }
        : entry,
    ),
  );
  await act(async () => {
    replied.resolve({ data: comment("saved-reply") });
    await pending;
  });
  expect(cached()![0].comments.map(({ id }) => id)).toEqual([
    "first-root",
    "saved-reply",
  ]);
});

it("holds background and explicit reads until the final document mutation settles", async () => {
  const { wrapper, cached } = setup(thread("existing"));
  const first = deferred<ApiResponse<DocumentCommentThread>>();
  const second = deferred<ApiResponse<DocumentCommentThread>>();
  jest
    .mocked(createDocumentCommentAction)
    .mockReturnValueOnce(first.promise)
    .mockReturnValueOnce(second.promise);
  jest.mocked(getDocumentComments).mockResolvedValue([thread("existing")]);
  const { result } = renderHook(
    () => ({
      comments: useDocumentComments("document"),
      mutations: useDocumentCommentMutations("document"),
    }),
    { wrapper },
  );
  await waitFor(() => {
    expect(result.current.comments.isFetching).toBe(false);
  });
  jest.mocked(getDocumentComments).mockClear();
  let one!: Promise<unknown>;
  let two!: Promise<unknown>;
  act(() => {
    one = result.current.mutations.create.mutateAsync(INPUT);
    two = result.current.mutations.create.mutateAsync({
      ...INPUT,
      body: "Second",
    });
  });
  await waitFor(() => {
    expect(createDocumentCommentAction).toHaveBeenCalledTimes(2);
  });
  await act(async () => {
    await result.current.comments.refetch();
  });
  expect(getDocumentComments).not.toHaveBeenCalled();
  expect(cached()).toHaveLength(3);
  await act(async () => {
    first.resolve({ data: thread("one") });
    await one;
  });
  expect(getDocumentComments).not.toHaveBeenCalled();
  jest
    .mocked(getDocumentComments)
    .mockResolvedValue([thread("two"), thread("one"), thread("existing")]);
  await act(async () => {
    second.resolve({ data: thread("two") });
    await two;
  });
  await waitFor(() => {
    expect(getDocumentComments).toHaveBeenCalledTimes(1);
  });
  expect(cached()!.map(({ id }) => id)).toEqual(["two", "one", "existing"]);
});

it("cancels an older in-flight read before inserting optimistic comments", async () => {
  const { wrapper, cached } = setup();
  const read = deferred<DocumentCommentThread[]>();
  const write = deferred<ApiResponse<DocumentCommentThread>>();
  jest
    .mocked(getDocumentComments)
    .mockReturnValueOnce(read.promise)
    .mockResolvedValue([thread("saved")]);
  jest.mocked(createDocumentCommentAction).mockReturnValue(write.promise);
  const { result } = renderHook(
    () => ({
      comments: useDocumentComments("document"),
      mutations: useDocumentCommentMutations("document"),
    }),
    { wrapper },
  );
  await waitFor(() => {
    expect(getDocumentComments).toHaveBeenCalledTimes(1);
  });
  let pending!: Promise<unknown>;
  act(() => {
    pending = result.current.mutations.create.mutateAsync(INPUT);
  });
  await waitFor(() => {
    expect(createDocumentCommentAction).toHaveBeenCalledTimes(1);
  });
  await act(async () => {
    read.resolve([thread("stale")]);
    await read.promise;
  });
  expect(cached()).toHaveLength(1);
  expect(cached()![0].id).toMatch(/^optimistic-thread-/);
  await act(async () => {
    write.resolve({ data: thread("saved") });
    await pending;
  });
  await waitFor(() => {
    expect(cached()![0].id).toBe("saved");
  });
});
