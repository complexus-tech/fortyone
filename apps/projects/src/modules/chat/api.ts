import type { JSONContent } from "@tiptap/core";
import { z } from "zod";
import { get, patch, post, put, remove } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  ChatAttachment,
  ChatConversation,
  ChatMessage,
  ChatMessageInput,
  ChatMessagePage,
  ChatPage,
  ChatPerson,
  ChatResource,
  ChatResourceFilter,
  ChatResourceSort,
  ChatSearchResult,
  ChatUnreadSummary,
  ChatUnreadThread,
  VoiceNoteUpload,
} from "./types";
import { CHAT_PAGE_SIZE } from "./constants";
import { extractMessageLinks } from "./chat-links";

const personSchema = z.object({
  id: z.string(),
  fullName: z.string(),
  avatarUrl: z.string().nullable().default(null),
  isSystem: z.boolean().default(false),
});
const attachmentSchema = z.object({
  id: z.string(),
  fileName: z.string(),
  mediaType: z.string(),
  sizeBytes: z.number(),
  url: z.string().default(""),
  kind: z.literal("voice-note").optional(),
  durationMs: z.number().int().positive().max(300_000).optional(),
});
const richContentSchema = z
  .custom<JSONContent>(
    (value) => typeof value === "object" && value !== null && "type" in value,
  )
  .nullable()
  .default(null);
const messageSchema = z
  .object({
    id: z.string(),
    conversationId: z.string(),
    threadRootId: z.string().nullable().default(null),
    creationId: z.string(),
    sequence: z.number().int(),
    author: personSchema,
    text: z.string(),
    richContent: richContentSchema,
    createdAt: z.string(),
    editedAt: z.string().nullable().default(null),
    deletedAt: z.string().nullable().default(null),
    version: z.number().int(),
    replyCount: z.number().int().default(0),
    unreadReplyCount: z.number().int().nonnegative().default(0),
    lastReplyAt: z.string().nullable().default(null),
    replyAuthors: z.array(personSchema).default([]),
    reactions: z
      .array(
        z.object({
          emoji: z.string(),
          count: z.number(),
          reacted: z.boolean().default(false),
        }),
      )
      .default([]),
    attachments: z.array(attachmentSchema).default([]),
    links: z
      .array(z.object({ id: z.string(), url: z.string(), title: z.string() }))
      .optional(),
    taskRefs: z
      .array(
        z.object({
          id: z.string(),
          title: z.string(),
          reference: z.string().optional(),
        }),
      )
      .default([]),
    mentionedUserIds: z.array(z.string()).default([]),
    agentMentions: z.array(z.literal("maya")).default([]),
    attachmentIds: z.array(z.string()).default([]),
  })
  .transform((message) => ({
    ...message,
    links: message.deletedAt
      ? []
      : message.links ?? extractMessageLinks(message.text, message.richContent),
  }));
const conversationSchema = z.object({
  id: z.string(),
  kind: z.enum(["team", "dm"]),
  teamId: z.string().nullable().default(null),
  name: z.string(),
  description: z.string().default(""),
  color: z.string().nullable().default(null),
  members: z.array(personSchema).default([]),
  memberCount: z.number().default(0),
  unreadCount: z.number().default(0),
  unreadRootCount: z.number().int().nonnegative().default(0),
  unreadReplyCount: z.number().int().nonnegative().default(0),
  lastReadSequence: z.number().default(0),
  lastMessageAt: z.string().nullable().default(null),
  lastMessageText: z.string().default(""),
});
const unreadSummarySchema = z.object({
  totalUnread: z.number().int().nonnegative(),
  unreadThreads: z.number().int().nonnegative(),
  conversations: z.array(
    z.object({
      conversationId: z.string(),
      unreadCount: z.number().int().nonnegative(),
      unreadRootCount: z.number().int().nonnegative(),
      unreadReplyCount: z.number().int().nonnegative(),
    }),
  ),
});
const unreadThreadSchema = z.intersection(
  messageSchema,
  z.object({
    firstUnreadReplyId: z.string(),
    conversationName: z.string(),
    conversationKind: z.enum(["team", "dm"]),
  }),
);
const messagePageSchema = z.object({
  items: z.array(messageSchema),
  olderCursor: z.string().nullable(),
  newerCursor: z.string().nullable(),
  hasOlder: z.boolean(),
  hasNewer: z.boolean(),
});
const resourceSchema = z.object({
  id: z.string(),
  conversationId: z.string(),
  attachmentId: z.string().default(""),
  description: z.string().default(""),
  kind: z.enum(["file", "link"]),
  messageId: z.string(),
  createdAt: z.string(),
  author: personSchema,
  fileName: z.string().default(""),
  title: z.string().default(""),
  mediaType: z.string().default(""),
  sizeBytes: z.number().default(0),
  url: z.string().default(""),
  attachmentKind: z.literal("voice-note").optional(),
  durationMs: z.number().int().positive().max(300_000).optional(),
});
const pageSchema = <T extends z.ZodType>(item: T) =>
  z.object({ items: z.array(item), nextCursor: z.string().nullable() });
const decode =
  <T>(schema: z.ZodType<T>) =>
  (value: unknown) => {
    const envelope = z
      .object({
        data: schema.nullable().optional(),
        error: z.object({ message: z.string() }).optional(),
      })
      .parse(value);
    if (envelope.error) throw new Error(envelope.error.message);
    if (!envelope.data)
      throw new Error("Chat is unavailable. Please try again.");
    return envelope.data;
  };
const queryString = (values: Record<string, string | number | undefined>) => {
  const search = new URLSearchParams();
  Object.entries(values).forEach(([key, value]) => {
    if (value !== undefined && value !== "") search.set(key, String(value));
  });
  return search.size ? `?${search.toString()}` : "";
};
const base = "chat";

export const getConversations = (
  ctx: WorkspaceCtx,
  search = "",
  cursor?: string,
  signal?: AbortSignal,
): Promise<ChatPage<ChatConversation>> =>
  get(
    `${base}/conversations${queryString({ search, cursor, limit: CHAT_PAGE_SIZE })}`,
    ctx,
    { signal },
    decode(pageSchema(conversationSchema)),
  );
export const getUnreadSummary = (
  ctx: WorkspaceCtx,
  signal?: AbortSignal,
): Promise<ChatUnreadSummary> =>
  get(`${base}/unread`, ctx, { signal }, decode(unreadSummarySchema));
export const getUnreadThreads = (
  ctx: WorkspaceCtx,
  cursor?: string,
  signal?: AbortSignal,
): Promise<ChatPage<ChatUnreadThread>> =>
  get(
    `${base}/threads/unread${queryString({ cursor, limit: CHAT_PAGE_SIZE })}`,
    ctx,
    { signal },
    decode(pageSchema(unreadThreadSchema)),
  );
export const getConversation = (
  ctx: WorkspaceCtx,
  conversationId: string,
  signal?: AbortSignal,
): Promise<ChatConversation> =>
  get(
    `${base}/conversations/${conversationId}`,
    ctx,
    { signal },
    decode(conversationSchema),
  );
export const createDirectConversation = (
  ctx: WorkspaceCtx,
  userId: string,
): Promise<ChatConversation> =>
  post(
    `${base}/conversations/direct`,
    { userId },
    ctx,
    { retry: 0 },
    decode(conversationSchema),
  );
export const getMessages = (
  ctx: WorkspaceCtx,
  conversationId: string,
  cursor: { before?: string; after?: string; around?: string } = {},
  signal?: AbortSignal,
): Promise<ChatMessagePage> =>
  get(
    `${base}/conversations/${conversationId}/messages${queryString({ ...cursor, limit: CHAT_PAGE_SIZE })}`,
    ctx,
    { signal },
    decode(messagePageSchema),
  );
export const getReplies = (
  ctx: WorkspaceCtx,
  rootId: string,
  cursor: { before?: string; after?: string; around?: string } = {},
  signal?: AbortSignal,
): Promise<ChatMessagePage> =>
  get(
    `${base}/messages/${rootId}/replies${queryString({ ...cursor, limit: CHAT_PAGE_SIZE })}`,
    ctx,
    { signal },
    decode(messagePageSchema),
  );
export const getMessage = (
  ctx: WorkspaceCtx,
  messageId: string,
  signal?: AbortSignal,
): Promise<ChatMessage> =>
  get(`${base}/messages/${messageId}`, ctx, { signal }, decode(messageSchema));
export const createMessage = (
  ctx: WorkspaceCtx,
  conversationId: string,
  input: ChatMessageInput,
): Promise<ChatMessage> =>
  post(
    `${base}/conversations/${conversationId}/messages`,
    input,
    ctx,
    { retry: 0 },
    decode(messageSchema),
  );
export const editMessage = (
  ctx: WorkspaceCtx,
  messageId: string,
  input: Omit<ChatMessageInput, "creationId" | "threadRootId"> & {
    version: number;
  },
): Promise<ChatMessage> =>
  patch(
    `${base}/messages/${messageId}`,
    input,
    ctx,
    { retry: 0 },
    decode(messageSchema),
  );
export const deleteMessage = (
  ctx: WorkspaceCtx,
  messageId: string,
  version: number,
) =>
  remove(
    `${base}/messages/${messageId}${queryString({ version })}`,
    ctx,
    { retry: 0 },
    decode(
      z.object({
        id: z.string(),
        conversationId: z.string(),
        threadRootId: z.string().nullable(),
        version: z.number(),
        deletedAt: z.string(),
      }),
    ),
  );
export const setReaction = async (
  ctx: WorkspaceCtx,
  messageId: string,
  emoji: string,
  selected: boolean,
) => {
  const path = `${base}/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`;
  if (selected) return put(path, {}, ctx, { retry: 0 }, decode(messageSchema));
  return remove(path, ctx, { retry: 0 }, decode(messageSchema));
};
export const markRead = (
  ctx: WorkspaceCtx,
  conversationId: string,
  throughMessageId: string,
  threadRootId?: string,
) =>
  post(
    `${base}/conversations/${conversationId}/read`,
    { throughMessageId, threadRootId },
    ctx,
    { retry: 0 },
  );
export const searchMessages = (
  ctx: WorkspaceCtx,
  conversationId: string,
  search: string,
  cursor?: string,
  signal?: AbortSignal,
): Promise<ChatPage<ChatSearchResult>> =>
  get(
    `${base}/conversations/${conversationId}/search${queryString({ q: search, cursor, limit: CHAT_PAGE_SIZE })}`,
    ctx,
    { signal },
    decode(pageSchema(messageSchema)),
  );
export const getResources = (
  ctx: WorkspaceCtx,
  conversationId: string,
  kind: ChatResourceFilter,
  search: string,
  cursor?: string,
  signal?: AbortSignal,
  sort: ChatResourceSort = "newest",
): Promise<ChatPage<ChatResource>> =>
  get(
    `${base}/conversations/${conversationId}/resources${queryString({ kind, sort, q: search, cursor, limit: CHAT_PAGE_SIZE })}`,
    ctx,
    { signal },
    decode(pageSchema(resourceSchema)),
  );
export const uploadAttachment = (
  ctx: WorkspaceCtx,
  conversationId: string,
  file: File,
  voiceNote?: VoiceNoteUpload,
): Promise<ChatAttachment> => {
  const data = new FormData();
  data.append("file", file);
  if (voiceNote) {
    data.append("kind", voiceNote.kind);
    data.append("durationMs", String(voiceNote.durationMs));
  }
  return post(
    `${base}/conversations/${conversationId}/attachments`,
    data,
    ctx,
    { timeout: 300_000, retry: 0 },
    decode(attachmentSchema),
  );
};
export const deleteAttachment = (
  ctx: WorkspaceCtx,
  conversationId: string,
  attachmentId: string,
) =>
  remove<ApiResponse<unknown>>(
    `${base}/conversations/${conversationId}/attachments/${attachmentId}`,
    ctx,
    { retry: 0 },
  );

export const getConversationMembers = (
  ctx: WorkspaceCtx,
  conversationId: string,
  search: string,
  signal?: AbortSignal,
): Promise<ChatPage<ChatPerson>> =>
  get(
    `${base}/conversations/${conversationId}/members${queryString({ q: search, limit: 15 })}`,
    ctx,
    { signal },
    decode(pageSchema(personSchema)),
  );
