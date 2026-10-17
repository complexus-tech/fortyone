import type { ChatMessage } from "./types";
import type { OutgoingChatMessage } from "./optimistic-reply-summaries";
import { withOptimisticReplySummary } from "./optimistic-reply-summaries";

const replySummaryRoot: ChatMessage = {
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
  lastReplyAt: "2026-10-17T09:00:00Z",
};

const reply = (id: string): OutgoingChatMessage => ({
  ...replySummaryRoot,
  id,
  creationId: id,
  threadRootId: replySummaryRoot.id,
  createdAt: "2026-10-17T10:00:00Z",
  author: { id: "me", fullName: "Joseph", avatarUrl: null, isSystem: false },
  optimisticReplySummary: { baseReplyCount: 10 },
  deliveryStatus: "sending",
});

describe("optimistic reply summaries", () => {
  it("immediately adds pending replies and their author without changing unread counts", () => {
    const summary = withOptimisticReplySummary(replySummaryRoot, [
      reply("one"),
      reply("two"),
    ]);
    expect(summary.replyCount).toBe(12);
    expect(summary.replyAuthors.map((author) => author.fullName)).toEqual([
      "Joseph",
    ]);
    expect(summary.lastReplyAt).toBe("2026-10-17T10:00:00Z");
    expect(summary.unreadReplyCount).toBe(3);
    expect(replySummaryRoot.replyCount).toBe(10);
  });

  it("does not double count after realtime has already observed the send", () => {
    const updated = { ...replySummaryRoot, replyCount: 11 };
    expect(withOptimisticReplySummary(updated, [reply("one")]).replyCount).toBe(
      11,
    );
    const newer = {
      ...updated,
      replyCount: 15,
      lastReplyAt: "2026-10-17T11:00:00Z",
    };
    const summary = withOptimisticReplySummary(newer, [reply("one")]);
    expect(summary.replyCount).toBe(15);
    expect(summary.lastReplyAt).toBe(newer.lastReplyAt);
  });

  it("removes only failed sends from the summary and preserves other pending sends", () => {
    expect(
      withOptimisticReplySummary(replySummaryRoot, [
        { ...reply("failed"), deliveryStatus: "failed" },
        reply("pending"),
      ]).replyCount,
    ).toBe(11);
    expect(
      withOptimisticReplySummary(replySummaryRoot, [
        { ...reply("failed"), deliveryStatus: "failed" },
      ]),
    ).toBe(replySummaryRoot);
  });

  it("handles a newer baseline without counting an earlier send twice", () => {
    const first = reply("first");
    const second = {
      ...reply("second"),
      optimisticReplySummary: { baseReplyCount: 11 },
    };
    expect(
      withOptimisticReplySummary({ ...replySummaryRoot, replyCount: 11 }, [
        first,
        second,
      ]).replyCount,
    ).toBe(12);
  });

  it("ignores other conversations, other threads and retired overlays", () => {
    expect(
      withOptimisticReplySummary(replySummaryRoot, [
        { ...reply("other-team"), conversationId: "another-team" },
        { ...reply("other-thread"), threadRootId: "another-root" },
        { ...reply("settled"), optimisticReplySummary: undefined },
      ]),
    ).toBe(replySummaryRoot);
  });
});
