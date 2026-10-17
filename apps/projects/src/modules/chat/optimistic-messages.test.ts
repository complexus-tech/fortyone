/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { QueryClient } from "@tanstack/react-query";
import type { ChatMessage, ChatPerson } from "./types";
import { chatKeys } from "./constants";
import {
  getCachedMessage,
  getQueryMessages,
  updateCachedMessage,
} from "./cache";
import {
  commitMessageReaction,
  deletedMessage,
  mutateMessageOptimistically,
  setMessageReaction,
} from "./optimistic-messages";

const person: ChatPerson = {
  id: "me",
  fullName: "Joseph",
  avatarUrl: null,
  isSystem: false,
};
const other: ChatPerson = { ...person, id: "other", fullName: "Sarah" };
const message: ChatMessage = {
  id: "reply",
  conversationId: "team",
  threadRootId: "root",
  creationId: "creation",
  sequence: 2,
  version: 1,
  author: person,
  text: "Hello",
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
  replyCount: 0,
  unreadReplyCount: 0,
  replyAuthors: [],
  lastReplyAt: null,
};
const clients: QueryClient[] = [];
const setup = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  clients.push(client);
  const page = {
    items: [message],
    olderCursor: null,
    newerCursor: null,
    hasOlder: false,
    hasNewer: false,
  };
  for (const key of [
    chatKeys.messages("workspace", "team"),
    chatKeys.replies("workspace", "team", "root"),
    chatKeys.search("workspace", "team", "Hello"),
  ])
    client.setQueryData(key, { pages: [page], pageParams: [{}] });
  client.setQueryData(chatKeys.unreadThreads("workspace"), {
    pages: [
      {
        items: [
          {
            ...message,
            firstUnreadReplyId: "unread",
            conversationName: "Team",
          },
        ],
        nextCursor: null,
      },
    ],
    pageParams: [undefined],
  });
  client.setQueryData(["chat", "workspace", "latest", "team"], page);
  client.setQueryData(["chat", "workspace", "message", "reply"], message);
  client.setQueryData(["chat", "another", "message", "reply"], message);
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
const read = (client: QueryClient) =>
  getCachedMessage(client, "workspace", message.id)!;
const toggle = (
  client: QueryClient,
  emoji: string,
  request: Promise<ChatMessage>,
) => {
  const original = read(client);
  const selected = !original.reactions.some(
    (item) => item.emoji === emoji && item.reacted,
  );
  return mutateMessageOptimistically(client, "workspace", original, {
    apply: (current) => setMessageReaction(current, emoji, selected, person),
    execute: () => request,
    commit: (current, response, refreshed) =>
      commitMessageReaction(
        current,
        response,
        emoji,
        selected,
        person,
        refreshed,
      ),
  });
};
afterEach(() => {
  clients.splice(0).forEach((client) => {
    client.clear();
  });
});

describe("optimistic message mutations", () => {
  it("updates conversation, thread, search, latest, detail, and unread-thread caches immediately without crossing workspaces", async () => {
    const client = setup();
    const request = deferred<ChatMessage>();
    const delivery = toggle(client, "👍", request.promise);
    for (const query of client
      .getQueryCache()
      .findAll({ queryKey: chatKeys.all("workspace") })) {
      expect(
        getQueryMessages(query.queryKey, query.state.data)[0].reactions,
      ).toEqual([{ emoji: "👍", count: 1, reacted: true, users: [person] }]);
    }
    expect(
      client.getQueryData(["chat", "another", "message", "reply"]),
    ).toEqual(message);
    expect(
      client.getQueryData(chatKeys.unreadThreads("workspace")),
    ).toMatchObject({
      pages: [
        { items: [{ firstUnreadReplyId: "unread", conversationName: "Team" }] },
      ],
    });
    request.resolve(setMessageReaction(message, "👍", true, person));
    await delivery;
  });

  it("keeps rapid toggles responsive and serializes their server writes", async () => {
    const client = setup();
    const first = deferred<ChatMessage>();
    const second = deferred<ChatMessage>();
    const sendFirst = jest.fn(() => first.promise);
    const sendSecond = jest.fn(() => second.promise);
    const one = mutateMessageOptimistically(client, "workspace", message, {
      apply: (current) => setMessageReaction(current, "👍", true, person),
      execute: sendFirst,
    });
    const two = mutateMessageOptimistically(client, "workspace", read(client), {
      apply: (current) => setMessageReaction(current, "👍", false, person),
      execute: sendSecond,
    });
    expect(read(client).reactions).toEqual([]);
    await new Promise((resolve) => {
      setTimeout(resolve, 0);
    });
    expect(sendFirst).toHaveBeenCalledTimes(1);
    expect(sendSecond).not.toHaveBeenCalled();
    first.resolve(setMessageReaction(message, "👍", true, person));
    await one;
    expect(read(client).reactions).toEqual([]);
    second.resolve(message);
    await two;
    expect(sendSecond).toHaveBeenCalledTimes(1);
    expect(read(client).reactions).toEqual([]);
  });

  it("rolls back only the failed change while preserving another queued reaction and realtime updates", async () => {
    const client = setup();
    const failed = deferred<ChatMessage>();
    const following = deferred<ChatMessage>();
    const first = toggle(client, "👍", failed.promise).catch(
      (error: unknown) => error,
    );
    const second = toggle(client, "🎉", following.promise);
    const refreshed = {
      ...message,
      replyCount: 2,
      reactions: [{ emoji: "👀", count: 1, reacted: false, users: [other] }],
    };
    client.setQueryData(["chat", "workspace", "latest", "team"], {
      items: [refreshed],
    });
    expect(read(client).reactions.map((item) => item.emoji)).toEqual([
      "👀",
      "👍",
      "🎉",
    ]);
    failed.reject(new Error("Offline"));
    await first;
    expect(read(client).replyCount).toBe(2);
    expect(read(client).reactions.map((item) => item.emoji)).toEqual([
      "👀",
      "🎉",
    ]);
    following.resolve(setMessageReaction(refreshed, "🎉", true, person));
    await second;
    expect(read(client).reactions.map((item) => item.emoji)).toEqual([
      "👀",
      "🎉",
    ]);
  });

  it("rebases a pending reaction over a refetch that already includes it without double counting", async () => {
    const client = setup();
    const request = deferred<ChatMessage>();
    const delivery = toggle(client, "👍", request.promise);
    await new Promise((resolve) => {
      setTimeout(resolve, 0);
    });
    const refreshed = {
      ...message,
      reactions: [
        { emoji: "👍", count: 2, reacted: true, users: [other, person] },
      ],
    };
    client.setQueryData(["chat", "workspace", "latest", "team"], {
      items: [refreshed],
    });
    expect(read(client).reactions[0].count).toBe(2);
    request.resolve(setMessageReaction(message, "👍", true, person));
    await delivery;
    expect(read(client).reactions[0].count).toBe(2);
  });

  it.each(["edit", "delete"] as const)(
    "restores an unsuccessful %s without discarding a newer server edit",
    async (kind) => {
      const client = setup();
      const request = deferred<ChatMessage>();
      const delivery = mutateMessageOptimistically(
        client,
        "workspace",
        message,
        {
          apply: (current) =>
            kind === "edit"
              ? { ...current, text: "My edit" }
              : deletedMessage(current, "now"),
          execute: () => request.promise,
        },
      ).catch((error: unknown) => error);
      expect(read(client).text).toBe(kind === "edit" ? "My edit" : "");
      const refreshed = { ...message, text: "Newer server edit", version: 2 };
      client.setQueryData(["chat", "workspace", "message", "reply"], refreshed);
      request.reject(new Error("Conflict"));
      await delivery;
      expect(read(client)).toMatchObject({
        text: "Newer server edit",
        version: 2,
        deletedAt: null,
      });
    },
  );

  it("never lets a stale message response downgrade the single-message cache", () => {
    const client = setup();
    updateCachedMessage(client, "workspace", {
      ...message,
      version: 3,
      text: "Current",
    });
    updateCachedMessage(client, "workspace", {
      ...message,
      version: 2,
      text: "Stale",
    });
    expect(
      client.getQueryData(["chat", "workspace", "message", "reply"]),
    ).toMatchObject({ version: 3, text: "Current" });
  });

  it("does not repopulate cleared account caches after a request finishes", async () => {
    const client = setup();
    const request = deferred<ChatMessage>();
    const delivery = toggle(client, "👍", request.promise).catch(
      (error: unknown) => error,
    );
    await Promise.resolve();
    await Promise.resolve();
    await Promise.resolve();
    client.clear();
    request.resolve(setMessageReaction(message, "👍", true, person));
    await delivery;
    expect(client.getQueryCache().getAll()).toEqual([]);
  });
});
