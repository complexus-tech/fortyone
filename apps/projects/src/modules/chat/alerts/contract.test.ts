import { parseChatNotification } from "./contract";

export const chatEvent = {
  type: "chat.notification",
  notificationId: "10000000-0000-4000-8000-000000000001",
  eventId: "10000000-0000-4000-8000-000000000002",
  workspaceId: "10000000-0000-4000-8000-000000000003",
  userId: "10000000-0000-4000-8000-000000000004",
  actorId: "10000000-0000-4000-8000-000000000005",
  conversationId: "10000000-0000-4000-8000-000000000006",
  messageId: "10000000-0000-4000-8000-000000000007",
  threadRootId: null,
  createdAt: "2026-10-17T12:00:00Z",
  occurredAt: "2026-10-17T12:00:00Z",
  title: "New chat message",
  body: "You have a new message.",
};

describe("chat notification contract", () => {
  it("retains recipient, event identity and exact conversation/thread destination", () => {
    expect(
      parseChatNotification({ ...chatEvent, threadRootId: chatEvent.eventId }),
    ).toEqual({
      ...Object.fromEntries(
        Object.entries(chatEvent).filter(([key]) => key !== "type"),
      ),
      kind: "chat-notification",
      threadRootId: chatEvent.eventId,
    });
  });

  it("accepts reaction copy through the existing chat notification contract", () => {
    const reaction = {
      ...chatEvent,
      threadRootId: "10000000-0000-4000-8000-000000000008",
      title: "Product",
      body: "Sarah reacted 🙌🏽 to your message: Ready for review",
    };

    expect(parseChatNotification(reaction)).toEqual({
      ...Object.fromEntries(
        Object.entries(reaction).filter(([key]) => key !== "type"),
      ),
      kind: "chat-notification",
    });
  });

  it.each([
    { actorId: "missing" },
    { conversationId: "https://outside.test" },
    { threadRootId: 12 },
    { occurredAt: "invalid" },
    { title: " " },
    { body: "x".repeat(1025) },
    { type: "chat.updated" },
  ])("rejects malformed delivery metadata: %o", (change) => {
    expect(parseChatNotification({ ...chatEvent, ...change })).toBeNull();
  });
});
