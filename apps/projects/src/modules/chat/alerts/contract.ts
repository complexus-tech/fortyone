export type ChatNotificationEvent = {
  kind: "chat-notification";
  notificationId: string;
  eventId: string;
  workspaceId: string;
  userId: string;
  actorId: string;
  conversationId: string;
  messageId: string;
  threadRootId: string | null;
  createdAt: string;
  occurredAt: string;
  title: string;
  body: string;
};

const UUID = /^[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}$/i;
const isId = (value: unknown): value is string =>
  typeof value === "string" && UUID.test(value);
const isTimestamp = (value: unknown): value is string =>
  typeof value === "string" &&
  value.length <= 64 &&
  Number.isFinite(Date.parse(value));

export const parseChatNotification = (
  value: Record<string, unknown>,
): ChatNotificationEvent | null => {
  if (
    value.type !== "chat.notification" ||
    !isId(value.notificationId) ||
    !isId(value.eventId) ||
    !isId(value.workspaceId) ||
    !isId(value.userId) ||
    !isId(value.actorId) ||
    !isId(value.conversationId) ||
    !isId(value.messageId) ||
    (value.threadRootId !== null &&
      value.threadRootId !== undefined &&
      !isId(value.threadRootId)) ||
    !isTimestamp(value.createdAt) ||
    !isTimestamp(value.occurredAt) ||
    typeof value.title !== "string" ||
    !value.title.trim() ||
    value.title.length > 256 ||
    typeof value.body !== "string" ||
    value.body.length > 1024
  )
    return null;

  return {
    kind: "chat-notification",
    notificationId: value.notificationId,
    eventId: value.eventId,
    workspaceId: value.workspaceId,
    userId: value.userId,
    actorId: value.actorId,
    conversationId: value.conversationId,
    messageId: value.messageId,
    threadRootId: value.threadRootId ?? null,
    createdAt: value.createdAt,
    occurredAt: value.occurredAt,
    title: value.title,
    body: value.body,
  };
};
