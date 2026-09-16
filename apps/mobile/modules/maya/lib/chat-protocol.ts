import type { MayaApproval, MayaMessage, MayaSession } from "../types";

const CHAT_ID = /^[A-Za-z0-9_-]{16}$/;
const UUID =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export const isEntityId = (value: unknown): value is string =>
  typeof value === "string" && UUID.test(value);

export const assertChatId = (id: string) => {
  if (!CHAT_ID.test(id)) throw new Error("This conversation link is invalid.");
  return id;
};

export const asRecord = (value: unknown): Record<string, unknown> =>
  value !== null && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};

export const parseMayaSessions = (
  value: unknown,
  userId: string,
): MayaSession[] => {
  if (!Array.isArray(value))
    throw new Error("Maya returned invalid conversation history.");
  return value.map((entry) => {
    const session = asRecord(entry);
    if (
      typeof session.id !== "string" ||
      !CHAT_ID.test(session.id) ||
      session.userId !== userId ||
      !isEntityId(session.workspaceId) ||
      typeof session.title !== "string" ||
      typeof session.createdAt !== "string" ||
      !Number.isFinite(Date.parse(session.createdAt)) ||
      typeof session.updatedAt !== "string" ||
      !Number.isFinite(Date.parse(session.updatedAt))
    )
      throw new Error("Maya returned invalid conversation history.");
    return session as MayaSession;
  });
};

export const getPendingMayaApprovals = (
  messages: MayaMessage[],
): MayaApproval[] => {
  const last = messages.at(-1);
  if (last?.role !== "assistant") return [];
  return last.parts.flatMap((part) => {
    const tool = asRecord(part);
    const approval = asRecord(tool.approval);
    if (
      !part.type.startsWith("tool-") ||
      tool.state !== "approval-requested" ||
      typeof tool.toolCallId !== "string" ||
      typeof approval.id !== "string"
    )
      return [];
    return [
      {
        id: approval.id,
        toolCallId: tool.toolCallId,
        toolName: part.type.slice(5),
        input: tool.input,
      },
    ];
  });
};

export const hasUnresolvedMayaApprovals = (messages: MayaMessage[]) =>
  messages.at(-1)?.parts.some((part) => {
    const tool = asRecord(part);
    return (
      part.type.startsWith("tool-") &&
      (tool.state === "approval-requested" ||
        tool.state === "approval-responded")
    );
  }) ?? false;

export const HISTORICAL_ATTACHMENT_PLACEHOLDER =
  "[Historical attachment omitted from this request.]";

/** Only an explicitly submitted new user turn may carry attachment bytes. */
export const prepareMobileMayaMessages = (
  messages: MayaMessage[],
  { preserveLastUserFiles = false }: { preserveLastUserFiles?: boolean } = {},
): MayaMessage[] =>
  messages.map((message, index) => {
    if (
      preserveLastUserFiles &&
      index === messages.length - 1 &&
      message.role === "user"
    )
      return message;
    if (!message.parts.some((part) => part.type === "file")) return message;
    return {
      ...message,
      parts: message.parts.map((part) =>
        part.type === "file"
          ? { type: "text" as const, text: HISTORICAL_ATTACHMENT_PLACEHOLDER }
          : part,
      ),
    };
  });

export const getMayaScreenPath = (reference?: string) =>
  reference && /^[A-Za-z][A-Za-z0-9_-]{0,30}-?\d+$/.test(reference)
    ? `/work/${encodeURIComponent(reference)}`
    : "/maya";
