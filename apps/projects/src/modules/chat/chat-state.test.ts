/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { QueryClient } from "@tanstack/react-query";
import type { ChatMessage, ChatMessagePage } from "./types";
import { flattenMessages, updateCachedMessage } from "./cache";
import { getMessageInput } from "./message-input";
import { reconcileChatQueries } from "./realtime";
import { chatKeys } from "./constants";

const message = (id: string, sequence: number, version = 1): ChatMessage => ({
  id,
  sequence,
  version,
  conversationId: "team",
  threadRootId: null,
  creationId: id,
  author: {
    id: "person",
    fullName: "Person",
    avatarUrl: null,
    isSystem: false,
  },
  text: id,
  richContent: null,
  createdAt: "2026-10-16T09:00:00Z",
  editedAt: null,
  deletedAt: null,
  replyCount: 0,
  lastReplyAt: null,
  replyAuthors: [],
  reactions: [],
  attachments: [],
  links: [],
  taskRefs: [],
  mentionedUserIds: [],
  agentMentions: [],
  attachmentIds: [],
});
const page = (...items: ChatMessage[]): ChatMessagePage => ({
  items,
  hasOlder: false,
  hasNewer: false,
  olderCursor: null,
  newerCursor: null,
});
afterEach(() => {
  jest.useRealTimers();
});

describe("chat cursor cache", () => {
  it("orders overlapping windows by sequence and never regresses a newer message version", () => {
    expect(
      flattenMessages([
        page(message("b", 2, 3)),
        page(message("a", 1), message("b", 2, 1)),
      ]).map(({ id, version }) => ({ id, version })),
    ).toEqual([
      { id: "a", version: 1 },
      { id: "b", version: 3 },
    ]);
    expect(flattenMessages(undefined)).toEqual([]);
  });
  it("updates all loaded copies without inserting messages into unrelated cursor windows", () => {
    const client = new QueryClient();
    const key = chatKeys.messages("workspace", "team");
    const unrelated = chatKeys.messages("workspace", "other");
    client.setQueryData(key, {
      pages: [page(message("a", 1))],
      pageParams: [{}],
    });
    client.setQueryData(unrelated, {
      pages: [page(message("b", 2))],
      pageParams: [{}],
    });
    const otherData = client.getQueryData(unrelated);
    const tailKey = [...chatKeys.all("workspace"), "latest", "team"];
    client.setQueryData(tailKey, page(message("a", 1)));
    updateCachedMessage(client, "workspace", message("a", 1, 2));
    expect(
      client.getQueryData<{ pages: ChatMessagePage[] }>(key)?.pages[0].items[0]
        .version,
    ).toBe(2);
    updateCachedMessage(client, "workspace", message("a", 1, 1));
    expect(
      client.getQueryData<{ pages: ChatMessagePage[] }>(key)?.pages[0].items[0]
        .version,
    ).toBe(2);
    updateCachedMessage(client, "workspace", message("c", 3));
    expect(
      client.getQueryData<{ pages: ChatMessagePage[] }>(key)?.pages[0].items,
    ).toHaveLength(1);
    expect(client.getQueryData(unrelated)).toBe(otherData);
    const reacted = {
      ...message("a", 1, 2),
      reactions: [{ emoji: "✅", count: 1, reacted: true }],
    };
    updateCachedMessage(client, "workspace", reacted);
    expect(
      client.getQueryData<ChatMessagePage>(tailKey)?.items[0].reactions,
    ).toEqual(reacted.reactions);
    client.clear();
  });
  it("coalesces event bursts while invalidating each affected conversation", async () => {
    jest.useFakeTimers();
    const client = new QueryClient();
    const first = chatKeys.messages("workspace", "first");
    const second = chatKeys.replies("workspace", "second", "root");
    const untouched = chatKeys.messages("workspace", "untouched");
    for (const key of [
      first,
      second,
      untouched,
      chatKeys.conversations("workspace"),
    ])
      client.setQueryData(key, {});
    const invalidate = jest.spyOn(client, "invalidateQueries");
    const promises = Array.from({ length: 50 }, (_, index) =>
      reconcileChatQueries(client, "workspace", index % 2 ? "first" : "second"),
    );
    expect(invalidate).not.toHaveBeenCalled();
    await jest.advanceTimersByTimeAsync(150);
    await Promise.all(promises);
    expect(invalidate).toHaveBeenCalledTimes(1);
    expect(client.getQueryState(first)?.isInvalidated).toBe(true);
    expect(client.getQueryState(second)?.isInvalidated).toBe(true);
    expect(client.getQueryState(untouched)?.isInvalidated).toBe(false);
    client.clear();
  });
});

it("extracts structural mentions only from active message text, not quotes, code or literal @Maya", () => {
  const maya = {
    type: "mention",
    attrs: {
      id: "maya-id",
      label: "Maya",
      mentionKind: "agent",
      agentKey: "maya",
    },
  };
  const human = {
    type: "mention",
    attrs: { id: "human-id", label: "Sarah", mentionKind: "user" },
  };
  const content = {
    type: "doc",
    content: [
      { type: "paragraph", content: [{ type: "text", text: "@Maya" }] },
      { type: "blockquote", content: [maya] },
      { type: "codeBlock", content: [maya] },
      { type: "paragraph", content: [human, human] },
    ],
  };
  expect(getMessageInput({ text: "@Maya Sarah", content })).toMatchObject({
    mentionedUserIds: ["human-id"],
    agentMentions: [],
  });
  content.content.push({ type: "paragraph", content: [maya] });
  expect(
    getMessageInput({ text: "@Maya Sarah", content }).agentMentions,
  ).toEqual(["maya"]);
});
