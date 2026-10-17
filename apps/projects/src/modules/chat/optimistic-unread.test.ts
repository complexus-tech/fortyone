/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { InfiniteData } from "@tanstack/react-query";
import { QueryClient } from "@tanstack/react-query";
import type {
  ChatConversation,
  ChatMarkUnreadResult,
  ChatMessage,
  ChatPage,
  ChatUnreadSummary,
  ChatUnreadThread,
} from "./types";
import { chatKeys } from "./constants";
import { mutateUnreadOptimistically } from "./optimistic-unread";

const message = (
  id: string,
  sequence: number,
  fields: Partial<ChatMessage> = {},
): ChatMessage => ({
  id,
  sequence,
  conversationId: "team",
  threadRootId: null,
  creationId: id,
  version: 1,
  author: { id: "peer", fullName: "Sarah", avatarUrl: null, isSystem: false },
  text: "Hello",
  richContent: null,
  createdAt: "2026-10-17T09:00:00Z",
  editedAt: null,
  deletedAt: null,
  replyCount: 0,
  unreadReplyCount: 0,
  lastReplyAt: null,
  replyAuthors: [],
  reactions: [],
  attachments: [],
  links: [],
  taskRefs: [],
  mentionedUserIds: [],
  agentMentions: [],
  attachmentIds: [],
  ...fields,
});
const root = message("root", 1, { replyCount: 4 });
const main = message("main", 5);
const later = message("later", 8);
const reply = message("reply", 6, { threadRootId: root.id });
const laterReply = message("later-reply", 9, { threadRootId: root.id });
const own = message("own", 10, {
  author: { id: "me", fullName: "Me", avatarUrl: null, isSystem: false },
});
const deleted = message("deleted", 7, { deletedAt: "2026-10-17T10:00:00Z" });
const conversation: ChatConversation = {
  id: "team",
  kind: "team",
  teamId: "team-id",
  name: "Product",
  description: "",
  color: null,
  members: [],
  memberCount: 4,
  unreadCount: 0,
  unreadRootCount: 0,
  unreadReplyCount: 0,
  lastReadSequence: 10,
  lastMessageAt: null,
  lastMessageText: "Hello",
};
const summary: ChatUnreadSummary = {
  totalUnread: 3,
  unreadThreads: 1,
  conversations: [
    {
      conversationId: "team",
      unreadCount: 0,
      unreadRootCount: 0,
      unreadReplyCount: 0,
    },
    {
      conversationId: "other",
      unreadCount: 3,
      unreadRootCount: 2,
      unreadReplyCount: 1,
    },
  ],
};
const unreadKey = [...chatKeys.unread("workspace"), "me"];
const threadsKey = [...chatKeys.unreadThreads("workspace"), "me"];
const rootKey = ["chat", "workspace", "message", root.id];
const page = <T>(items: T[]): InfiniteData<ChatPage<T>> => ({
  pages: [{ items, nextCursor: null }],
  pageParams: [undefined],
});
const messagesPage = (items: ChatMessage[]) => ({
  pages: [
    {
      items,
      olderCursor: null,
      newerCursor: null,
      hasOlder: false,
      hasNewer: false,
    },
  ],
  pageParams: [{}],
});
const clients: QueryClient[] = [];
const setup = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity } },
  });
  clients.push(client);
  client.setQueryData(chatKeys.conversation("workspace", "team"), conversation);
  client.setQueryData(
    chatKeys.conversations("workspace"),
    page([conversation]),
  );
  client.setQueryData(
    chatKeys.conversations("workspace", "Product"),
    page([conversation]),
  );
  client.setQueryData(
    chatKeys.messages("workspace", "team"),
    messagesPage([root, main, later, own, deleted]),
  );
  client.setQueryData(
    chatKeys.replies("workspace", "team", root.id),
    messagesPage([reply, laterReply]),
  );
  client.setQueryData(rootKey, root);
  client.setQueryData(threadsKey, page<ChatUnreadThread>([]));
  client.setQueryData(unreadKey, summary);
  client.setQueryData(
    [...chatKeys.unread("workspace"), "other-account"],
    summary,
  );
  client.setQueryData(
    [...chatKeys.unreadThreads("workspace"), "other-account"],
    page<ChatUnreadThread>([]),
  );
  client.setQueryData(chatKeys.conversation("elsewhere", "team"), conversation);
  return client;
};
const deferred = <T>() => {
  let resolve!: (value: T) => void;
  let reject!: (reason: Error) => void;
  const promise = new Promise<T>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
};
const readConversation = (client: QueryClient) =>
  client.getQueryData<ChatConversation>(
    chatKeys.conversation("workspace", "team"),
  )!;
const readSummary = (client: QueryClient) =>
  client.getQueryData<ChatUnreadSummary>(unreadKey)!;
const mark = (
  client: QueryClient,
  target: ChatMessage,
  execute: () => Promise<ChatMarkUnreadResult>,
) =>
  mutateUnreadOptimistically(
    client,
    { workspaceSlug: "Workspace", accountId: "me", message: target },
    execute,
  );
const result = (
  target: ChatMessage,
  count: number,
  extra: Partial<ChatMarkUnreadResult> = {},
): ChatMarkUnreadResult => ({
  conversationId: target.conversationId,
  threadRootId: target.threadRootId,
  lastReadSequence: target.sequence - 1,
  unreadCount: count,
  firstUnreadReplyId: target.threadRootId ? target.id : null,
  summary: {
    ...summary,
    totalUnread: 3 + count,
    unreadThreads: target.threadRootId ? 2 : 1,
    conversations: [
      {
        conversationId: "team",
        unreadCount: count,
        unreadRootCount: target.threadRootId ? 0 : count,
        unreadReplyCount: target.threadRootId ? count : 0,
      },
      summary.conversations[1],
    ],
  },
  ...extra,
});
const tick = () =>
  new Promise<void>((resolve) => {
    setTimeout(resolve, 0);
  });
afterEach(() => {
  clients.splice(0).forEach((client) => {
    client.clear();
  });
});

describe("optimistic mark unread", () => {
  it("rewinds only main messages across all lists and updates account-scoped totals immediately", async () => {
    const client = setup();
    const request = deferred<ChatMarkUnreadResult>();
    const delivery = mark(client, main, () => request.promise);
    expect(readConversation(client)).toMatchObject({
      lastReadSequence: 4,
      unreadRootCount: 2,
      unreadReplyCount: 0,
      unreadCount: 2,
    });
    for (const key of [
      chatKeys.conversations("workspace"),
      chatKeys.conversations("workspace", "Product"),
    ])
      expect(client.getQueryData(key)).toMatchObject({
        pages: [{ items: [{ unreadRootCount: 2, lastReadSequence: 4 }] }],
      });
    expect(readSummary(client)).toMatchObject({
      totalUnread: 5,
      unreadThreads: 1,
    });
    expect(
      client.getQueryData([...chatKeys.unread("workspace"), "other-account"]),
    ).toEqual(summary);
    expect(
      client.getQueryData(chatKeys.conversation("elsewhere", "team")),
    ).toEqual(conversation);
    expect(client.getQueryData(rootKey)).toEqual(root);
    request.resolve(result(main, 4));
    await delivery;
    expect(readSummary(client).totalUnread).toBe(7);
    expect(readConversation(client).unreadRootCount).toBe(4);
  });

  it("marks thread replies optimistically without rewinding the main stream and adds its unread-thread row", async () => {
    const client = setup();
    const request = deferred<ChatMarkUnreadResult>();
    const delivery = mark(client, reply, () => request.promise);
    expect(readConversation(client)).toMatchObject({
      lastReadSequence: 10,
      unreadRootCount: 0,
      unreadReplyCount: 2,
    });
    expect(client.getQueryData(rootKey)).toMatchObject({ unreadReplyCount: 2 });
    expect(
      client.getQueryData(chatKeys.messages("workspace", "team")),
    ).toMatchObject({
      pages: [
        {
          items: expect.arrayContaining([
            expect.objectContaining({ id: "root", unreadReplyCount: 2 }),
          ]),
        },
      ],
    });
    expect(readSummary(client)).toMatchObject({
      totalUnread: 5,
      unreadThreads: 2,
    });
    expect(client.getQueryData(threadsKey)).toMatchObject({
      pageParams: [undefined],
      pages: [
        {
          nextCursor: null,
          items: [
            {
              id: "root",
              firstUnreadReplyId: "reply",
              unreadReplyCount: 2,
              conversationName: "Product",
              conversationKind: "team",
            },
          ],
        },
      ],
    });
    expect(
      client.getQueryData([
        ...chatKeys.unreadThreads("workspace"),
        "other-account",
      ]),
    ).toEqual(page([]));
    request.resolve(result(reply, 3));
    await delivery;
    expect(readConversation(client).unreadReplyCount).toBe(3);
    expect(client.getQueryData(rootKey)).toMatchObject({ unreadReplyCount: 3 });
    expect(readSummary(client).unreadThreads).toBe(2);
  });

  it("preserves existing older unread counts when marking a later message", async () => {
    const client = setup();
    client.setQueryData(chatKeys.conversation("workspace", "team"), {
      ...conversation,
      lastReadSequence: 0,
      unreadRootCount: 3,
      unreadCount: 3,
    });
    client.setQueryData(
      chatKeys.conversations("workspace"),
      page([
        {
          ...conversation,
          lastReadSequence: 0,
          unreadRootCount: 3,
          unreadCount: 3,
        },
      ]),
    );
    const request = deferred<ChatMarkUnreadResult>();
    const delivery = mark(client, later, () => request.promise);
    expect(readConversation(client)).toMatchObject({
      lastReadSequence: 0,
      unreadRootCount: 3,
    });
    request.resolve(result(later, 3, { lastReadSequence: 0 }));
    await delivery;
    expect(readConversation(client)).toMatchObject({
      lastReadSequence: 0,
      unreadRootCount: 3,
    });
  });

  it("rolls back only unread fields while preserving a concurrent local message edit and unrelated counters", async () => {
    const client = setup();
    const request = deferred<ChatMarkUnreadResult>();
    const delivery = mark(client, reply, () => request.promise);
    client.setQueryData<ChatMessage>(rootKey, (current) => ({
      ...current!,
      text: "Edited while marking",
      version: 2,
    }));
    client.setQueryData<ChatUnreadSummary>(unreadKey, (current) => ({
      ...current!,
      totalUnread: current!.totalUnread + 2,
      conversations: current!.conversations.map((item) =>
        item.conversationId === "other"
          ? { ...item, unreadRootCount: 4, unreadCount: 5 }
          : item,
      ),
    }));
    request.reject(new Error("offline"));
    await expect(delivery).rejects.toThrow("offline");
    expect(client.getQueryData(rootKey)).toMatchObject({
      text: "Edited while marking",
      version: 2,
      unreadReplyCount: 0,
    });
    expect(readConversation(client).unreadCount).toBe(0);
    expect(readSummary(client)).toMatchObject({
      totalUnread: 5,
      unreadThreads: 1,
    });
    expect(client.getQueryData(threadsKey)).toEqual(page([]));
  });

  it("rebases pending intention over an authorized SSE refresh and restores that refresh on failure", async () => {
    const client = setup();
    const request = deferred<ChatMarkUnreadResult>();
    const delivery = mark(client, main, () => request.promise);
    const refreshed = {
      ...conversation,
      lastReadSequence: 7,
      unreadCount: 1,
      unreadRootCount: 1,
      name: "Renamed",
    };
    await client.fetchQuery({
      queryKey: chatKeys.conversation("workspace", "team"),
      queryFn: async () => refreshed,
    });
    expect(readConversation(client)).toMatchObject({
      name: "Renamed",
      unreadRootCount: 2,
      lastReadSequence: 4,
    });
    request.reject(new Error("offline"));
    await expect(delivery).rejects.toThrow("offline");
    expect(readConversation(client)).toEqual(refreshed);
  });

  it("keeps later optimistic actions when an earlier mark fails and serializes requests per conversation", async () => {
    const client = setup();
    const first = deferred<ChatMarkUnreadResult>();
    const second = deferred<ChatMarkUnreadResult>();
    const sendFirst = jest.fn(() => first.promise);
    const sendSecond = jest.fn(() => second.promise);
    const one = mark(client, main, sendFirst);
    const two = mark(client, reply, sendSecond);
    expect(readConversation(client)).toMatchObject({
      unreadRootCount: 2,
      unreadReplyCount: 2,
      unreadCount: 4,
    });
    await tick();
    expect(sendFirst).toHaveBeenCalledTimes(1);
    expect(sendSecond).not.toHaveBeenCalled();
    first.reject(new Error("offline"));
    await expect(one).rejects.toThrow("offline");
    expect(readConversation(client)).toMatchObject({
      unreadRootCount: 0,
      unreadReplyCount: 2,
      unreadCount: 2,
    });
    second.resolve(result(reply, 2));
    await two;
    expect(sendSecond).toHaveBeenCalledTimes(1);
    expect(readSummary(client).unreadThreads).toBe(2);
  });

  it("does not double-count rapid repeated marks on the same thread", async () => {
    const client = setup();
    const first = deferred<ChatMarkUnreadResult>();
    const second = deferred<ChatMarkUnreadResult>();
    const one = mark(client, reply, () => first.promise);
    const two = mark(client, reply, () => second.promise);
    expect(readConversation(client).unreadReplyCount).toBe(2);
    expect(readSummary(client).unreadThreads).toBe(2);
    first.resolve(result(reply, 2));
    await one;
    expect(readConversation(client).unreadReplyCount).toBe(2);
    second.resolve(result(reply, 2));
    await two;
    expect(readSummary(client).unreadThreads).toBe(2);
  });

  it("does not repopulate cleared account caches after a late successful response", async () => {
    const client = setup();
    const request = deferred<ChatMarkUnreadResult>();
    const delivery = mark(client, main, () => request.promise);
    await tick();
    client.clear();
    request.resolve(result(main, 2));
    await delivery;
    expect(client.getQueryCache().getAll()).toHaveLength(0);
  });
  it("accepts a zero authoritative count after a concurrent deletion and removes the optimistic unread thread", async () => {
    const client = setup();
    const request = deferred<ChatMarkUnreadResult>();
    const delivery = mark(client, reply, () => request.promise);
    expect(readSummary(client).unreadThreads).toBe(2);
    request.resolve(result(reply, 0, { firstUnreadReplyId: null }));
    await delivery;
    expect(readConversation(client)).toMatchObject({
      unreadReplyCount: 0,
      unreadCount: 0,
    });
    expect(client.getQueryData(rootKey)).toMatchObject({ unreadReplyCount: 0 });
    expect(readSummary(client)).toMatchObject({
      totalUnread: 3,
      unreadThreads: 1,
    });
    expect(client.getQueryData(threadsKey)).toEqual(page([]));
  });

  it("keeps a newer fetched count when an older mutation response completes", async () => {
    const client = setup();
    const request = deferred<ChatMarkUnreadResult>();
    const delivery = mark(client, reply, () => request.promise);
    await tick();
    await client.fetchQuery({
      queryKey: rootKey,
      queryFn: async () => ({ ...root, unreadReplyCount: 3 }),
    });
    request.resolve(result(reply, 2));
    await delivery;
    expect(client.getQueryData(rootKey)).toMatchObject({ unreadReplyCount: 3 });
    expect(readConversation(client).unreadReplyCount).toBe(3);
  });

  it("does not cancel or project other conversation message queries", async () => {
    const client = setup();
    const otherKey = chatKeys.messages("workspace", "other");
    const otherMessages = messagesPage([
      message("other-message", 1, { conversationId: "other" }),
    ]);
    client.setQueryData(otherKey, otherMessages);
    const reads = deferred<typeof otherMessages>();
    const read = client.fetchQuery({
      queryKey: otherKey,
      queryFn: () => reads.promise,
    });
    const request = deferred<ChatMarkUnreadResult>();
    const delivery = mark(client, main, () => request.promise);
    expect(client.getQueryState(otherKey)?.fetchStatus).toBe("fetching");
    expect(client.getQueryData(otherKey)).toBe(otherMessages);
    request.resolve(result(main, 2));
    reads.resolve(otherMessages);
    await Promise.all([delivery, read]);
  });
  it("does not restore stale snapshots from a settled conversation when another room is pending", async () => {
    const client = setup();
    const other = { ...conversation, id: "other", name: "Another team" };
    client.setQueryData(chatKeys.conversation("workspace", "other"), other);
    const first = deferred<ChatMarkUnreadResult>();
    const second = deferred<ChatMarkUnreadResult>();
    const otherMessage = message("other-message", 4, {
      conversationId: "other",
    });
    const one = mark(client, main, () => first.promise);
    const two = mark(client, otherMessage, () => second.promise);
    first.resolve(result(main, 2));
    await one;
    client.setQueryData<ChatConversation>(
      chatKeys.conversation("workspace", "team"),
      (current) => ({ ...current!, name: "Changed after success" }),
    );
    second.resolve(result(otherMessage, 1));
    await two;
    expect(readConversation(client).name).toBe("Changed after success");
  });
});
