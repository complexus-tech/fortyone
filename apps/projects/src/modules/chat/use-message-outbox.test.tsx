import { randomUUID } from "node:crypto";
import type { ReactNode } from "react";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import * as api from "./api";
import { chatKeys } from "./constants";
import type { ChatMessage } from "./types";
import { withOptimisticReplySummary } from "./optimistic-reply-summaries";
import { useMessageOutbox } from "./use-message-outbox";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({
    data: { user: { id: "me", name: "Joseph", image: null } },
  }),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
jest.mock("./api");
jest.mock("./realtime", () => ({
  reconcileChatQueries: jest.fn().mockResolvedValue(undefined),
}));
const mocked = jest.mocked(api);
const root: Awaited<ReturnType<typeof api.getMessage>> = {
  id: "root",
  conversationId: "team",
  threadRootId: null,
  creationId: "root-creation",
  sequence: 1,
  version: 1,
  author: { id: "author", fullName: "Sarah", avatarUrl: null, isSystem: false },
  text: "Discussion",
  richContent: null,
  createdAt: "2026-10-17T09:00:00Z",
  editedAt: null,
  deletedAt: null,
  mentionedUserIds: [],
  agentMentions: [],
  attachmentIds: [],
  attachments: [],
  links: [],
  taskRefs: [],
  reactions: [],
  replyCount: 10,
  unreadReplyCount: 3,
  replyAuthors: [],
  lastReplyAt: null,
};
const rootKey = [...chatKeys.all("workspace"), "message", root.id];
const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
};
let client: QueryClient;
const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);
beforeEach(() => {
  Object.defineProperty(globalThis.crypto, "randomUUID", {
    configurable: true,
    value: randomUUID,
  });
  client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity } },
  });
  client.setQueryData(rootKey, root);
});
afterEach(() => {
  cleanup();
  client.clear();
  jest.clearAllMocks();
});

describe("reply outbox summaries", () => {
  it("keeps an immediate summary until the authoritative parent catches up", async () => {
    const sendResponse = deferred<ChatMessage>();
    const parentResponse =
      deferred<Awaited<ReturnType<typeof api.getMessage>>>();
    mocked.createMessage.mockReturnValue(sendResponse.promise);
    mocked.getMessage.mockReturnValue(parentResponse.promise);
    const { result } = renderHook(useMessageOutbox, { wrapper });
    let sent!: NonNullable<ReturnType<typeof result.current.enqueue>>;
    act(() => {
      sent = result.current.enqueue(
        "team",
        { text: "Reply", content: null },
        "root",
      )!;
    });
    await waitFor(() => {
      expect(result.current.outgoing).toHaveLength(1);
    });
    expect(
      withOptimisticReplySummary(root, result.current.outgoing).replyCount,
    ).toBe(11);
    expect(client.getQueryData<ChatMessage>(rootKey)?.replyCount).toBe(10);
    const delivered = {
      ...sent.pending,
      id: "reply",
      deliveryStatus: undefined,
      sequence: 2,
      version: 1,
    };
    await act(async () => {
      sendResponse.resolve(delivered);
      await sent.delivery;
    });
    act(() => {
      result.current.acknowledge([delivered]);
    });
    await waitFor(() => {
      expect(result.current.outgoing[0]?.acknowledged).toBe(true);
    });
    expect(
      withOptimisticReplySummary(root, result.current.outgoing).replyCount,
    ).toBe(11);
    await act(async () => {
      parentResponse.resolve({ ...root, replyCount: 11 });
    });
    await waitFor(() => {
      expect(result.current.outgoing).toHaveLength(0);
    });
    expect(client.getQueryData<ChatMessage>(rootKey)?.replyCount).toBe(11);
  });

  it("keeps the confirmed prefix while a second reply is still pending", async () => {
    const firstResponse = deferred<ChatMessage>();
    const secondResponse = deferred<ChatMessage>();
    const parentResponse =
      deferred<Awaited<ReturnType<typeof api.getMessage>>>();
    mocked.createMessage
      .mockReturnValueOnce(firstResponse.promise)
      .mockReturnValueOnce(secondResponse.promise);
    mocked.getMessage.mockReturnValue(parentResponse.promise);
    const { result } = renderHook(useMessageOutbox, { wrapper });
    let first!: NonNullable<ReturnType<typeof result.current.enqueue>>;
    let second!: NonNullable<ReturnType<typeof result.current.enqueue>>;
    act(() => {
      first = result.current.enqueue(
        "team",
        { text: "First", content: null },
        "root",
      )!;
      second = result.current.enqueue(
        "team",
        { text: "Second", content: null },
        "root",
      )!;
    });
    const firstDelivered = {
      ...first.pending,
      id: "first",
      deliveryStatus: undefined,
      sequence: 2,
      version: 1,
    };
    await act(async () => {
      firstResponse.resolve(firstDelivered);
      await first.delivery;
    });
    const afterFirst = { ...root, replyCount: 11 };
    client.setQueryData(rootKey, afterFirst);
    act(() => {
      result.current.acknowledge([firstDelivered]);
    });
    await waitFor(() => {
      expect(result.current.outgoing[0]?.acknowledged).toBe(true);
    });
    expect(
      withOptimisticReplySummary(afterFirst, result.current.outgoing)
        .replyCount,
    ).toBe(12);
    expect(mocked.getMessage).not.toHaveBeenCalled();

    const secondDelivered = {
      ...second.pending,
      id: "second",
      deliveryStatus: undefined,
      sequence: 3,
      version: 1,
    };
    await act(async () => {
      secondResponse.resolve(secondDelivered);
      await second.delivery;
    });
    act(() => {
      result.current.acknowledge([secondDelivered]);
    });
    expect(
      withOptimisticReplySummary(afterFirst, result.current.outgoing)
        .replyCount,
    ).toBe(12);
    expect(mocked.getMessage).toHaveBeenCalledTimes(1);
    await act(async () => {
      parentResponse.resolve({ ...root, replyCount: 12 });
    });
    await waitFor(() => {
      expect(result.current.outgoing).toHaveLength(0);
    });
    expect(client.getQueryData<ChatMessage>(rootKey)?.replyCount).toBe(12);
  });

  it("does not let an older root read retire or overwrite a newer reply cohort", async () => {
    const firstResponse = deferred<ChatMessage>();
    const secondResponse = deferred<ChatMessage>();
    const firstParent = deferred<Awaited<ReturnType<typeof api.getMessage>>>();
    const secondParent = deferred<Awaited<ReturnType<typeof api.getMessage>>>();
    mocked.createMessage
      .mockReturnValueOnce(firstResponse.promise)
      .mockReturnValueOnce(secondResponse.promise);
    mocked.getMessage
      .mockReturnValueOnce(firstParent.promise)
      .mockReturnValueOnce(secondParent.promise);
    const { result } = renderHook(useMessageOutbox, { wrapper });
    let first!: NonNullable<ReturnType<typeof result.current.enqueue>>;
    let second!: NonNullable<ReturnType<typeof result.current.enqueue>>;
    act(() => {
      first = result.current.enqueue(
        "team",
        { text: "First", content: null },
        "root",
      )!;
    });
    const deliveredFirst = {
      ...first.pending,
      id: "first",
      deliveryStatus: undefined,
      sequence: 2,
    };
    await act(async () => {
      firstResponse.resolve(deliveredFirst);
      await first.delivery;
    });
    expect(mocked.getMessage).toHaveBeenCalledTimes(1);
    act(() => {
      result.current.acknowledge([deliveredFirst]);
      second = result.current.enqueue(
        "team",
        { text: "Second", content: null },
        "root",
      )!;
    });
    const deliveredSecond = {
      ...second.pending,
      id: "second",
      deliveryStatus: undefined,
      sequence: 3,
    };
    await act(async () => {
      secondResponse.resolve(deliveredSecond);
      await second.delivery;
    });
    act(() => {
      result.current.acknowledge([deliveredSecond]);
    });
    await act(async () => {
      secondParent.resolve({ ...root, replyCount: 12 });
    });
    await waitFor(() => {
      expect(result.current.outgoing).toHaveLength(0);
    });
    await act(async () => {
      firstParent.resolve({ ...root, replyCount: 11 });
    });
    expect(client.getQueryData<ChatMessage>(rootKey)?.replyCount).toBe(12);
  });

  it("removes a failed first send from a cohort while keeping a later pending reply", async () => {
    const firstResponse = deferred<ChatMessage>();
    const secondResponse = deferred<ChatMessage>();
    mocked.createMessage
      .mockReturnValueOnce(firstResponse.promise)
      .mockReturnValueOnce(secondResponse.promise);
    mocked.getMessage.mockResolvedValue({ ...root, replyCount: 11 });
    const { result } = renderHook(useMessageOutbox, { wrapper });
    let first!: NonNullable<ReturnType<typeof result.current.enqueue>>;
    let second!: NonNullable<ReturnType<typeof result.current.enqueue>>;
    act(() => {
      first = result.current.enqueue(
        "team",
        { text: "First", content: null },
        "root",
      )!;
      second = result.current.enqueue(
        "team",
        { text: "Second", content: null },
        "root",
      )!;
    });
    await act(async () => {
      firstResponse.reject(new Error("Offline"));
      await first.delivery;
    });
    await waitFor(() => {
      expect(result.current.outgoing[0]?.deliveryStatus).toBe("failed");
    });
    expect(
      withOptimisticReplySummary(root, result.current.outgoing).replyCount,
    ).toBe(11);
    expect(mocked.getMessage).not.toHaveBeenCalled();
    await act(async () => {
      secondResponse.resolve({
        ...second.pending,
        id: "second",
        deliveryStatus: undefined,
        sequence: 2,
      });
      await second.delivery;
    });
    await waitFor(() => {
      expect(
        result.current.outgoing.every((item) => !item.optimisticReplySummary),
      ).toBe(true);
    });
    expect(client.getQueryData<ChatMessage>(rootKey)?.replyCount).toBe(11);
  });

  it("removes a failed send from the summary without undoing concurrent server changes", async () => {
    const sendResponse = deferred<ChatMessage>();
    mocked.createMessage.mockReturnValue(sendResponse.promise);
    const { result } = renderHook(useMessageOutbox, { wrapper });
    let sent!: NonNullable<ReturnType<typeof result.current.enqueue>>;
    act(() => {
      sent = result.current.enqueue(
        "team",
        { text: "Reply", content: null },
        "root",
      )!;
    });
    await waitFor(() => {
      expect(result.current.outgoing).toHaveLength(1);
    });
    const newer = { ...root, replyCount: 12 };
    client.setQueryData(rootKey, newer);
    await act(async () => {
      sendResponse.reject(new Error("Offline"));
      await sent.delivery;
    });
    await waitFor(() => {
      expect(result.current.outgoing[0]?.deliveryStatus).toBe("failed");
    });
    expect(withOptimisticReplySummary(newer, result.current.outgoing)).toBe(
      newer,
    );
    expect(client.getQueryData<ChatMessage>(rootKey)?.replyCount).toBe(12);
  });

  it("does not turn an SSE-confirmed reply into a failure when its HTTP response is lost", async () => {
    const sendResponse = deferred<ChatMessage>();
    mocked.createMessage.mockReturnValue(sendResponse.promise);
    mocked.getMessage.mockResolvedValue({ ...root, replyCount: 11 });
    const { result } = renderHook(useMessageOutbox, { wrapper });
    let sent!: NonNullable<ReturnType<typeof result.current.enqueue>>;
    act(() => {
      sent = result.current.enqueue(
        "team",
        { text: "Reply", content: null },
        "root",
      )!;
    });
    const delivered = {
      ...sent.pending,
      id: "reply",
      sequence: 2,
      deliveryStatus: undefined,
      version: 1,
    };
    act(() => {
      result.current.acknowledge([delivered]);
    });
    await act(async () => {
      sendResponse.reject(new Error("Connection lost"));
      await sent.delivery;
    });
    await waitFor(() => {
      expect(result.current.outgoing).toHaveLength(0);
    });
    expect(client.getQueryData<ChatMessage>(rootKey)?.replyCount).toBe(11);
    expect(toast.error).not.toHaveBeenCalled();
  });
});
