import type { JSONContent } from "@tiptap/core";

export type ChatPerson = {
  id: string;
  fullName: string;
  avatarUrl: string | null;
  isSystem: boolean;
};

export type ChatReaction = {
  emoji: string;
  count: number;
  reacted: boolean;
  users?: ChatPerson[];
};
export type ChatAttachment = {
  id: string;
  fileName: string;
  mediaType: string;
  sizeBytes: number;
  url: string;
  kind?: "voice-note";
  durationMs?: number;
};
export type VoiceNoteUpload = { kind: "voice-note"; durationMs: number };
export type ChatLink = { id: string; url: string; title: string };
export type ChatTask = { id: string; title: string; reference?: string };

export type ChatMessage = {
  deliveryStatus?: "sending" | "failed";
  id: string;
  conversationId: string;
  threadRootId: string | null;
  creationId: string;
  sequence: number;
  author: ChatPerson;
  text: string;
  richContent: JSONContent | null;
  createdAt: string;
  editedAt: string | null;
  deletedAt: string | null;
  version: number;
  replyCount: number;
  unreadReplyCount: number;
  lastReplyAt: string | null;
  replyAuthors: ChatPerson[];
  reactions: ChatReaction[];
  attachments: ChatAttachment[];
  links: ChatLink[];
  taskRefs: ChatTask[];
  mentionedUserIds: string[];
  agentMentions: "maya"[];
  attachmentIds: string[];
};

export type ChatConversation = {
  id: string;
  kind: "team" | "dm";
  teamId: string | null;
  name: string;
  description: string;
  color: string | null;
  members: ChatPerson[];
  memberCount: number;
  unreadCount: number;
  unreadRootCount: number;
  unreadReplyCount: number;
  lastReadSequence: number;
  lastMessageAt: string | null;
  lastMessageText: string;
};

export type ConversationFeed = ChatConversation & { messages: ChatMessage[] };
export type ChatUnreadSummary = {
  totalUnread: number;
  unreadThreads: number;
  conversations: {
    conversationId: string;
    unreadCount: number;
    unreadRootCount: number;
    unreadReplyCount: number;
  }[];
};
export type ChatUnreadThread = ChatMessage & {
  firstUnreadReplyId: string;
  conversationName: string;
  conversationKind: ChatConversation["kind"];
};
export type ConversationTab = "messages" | "files";
export type ChatPage<T> = { items: T[]; nextCursor: string | null };
export type ChatMessagePage = {
  items: ChatMessage[];
  olderCursor: string | null;
  newerCursor: string | null;
  hasOlder: boolean;
  hasNewer: boolean;
};
export type ChatSearchResult = ChatMessage;
export type ChatResourceFilter = "all" | "files" | "media" | "links";
export type ChatResourceSort = "newest" | "oldest";
export type ChatResource = {
  id: string;
  conversationId: string;
  attachmentId: string;
  description: string;
  kind: "file" | "link";
  messageId: string;
  createdAt: string;
  author: ChatPerson;
  fileName: string;
  title: string;
  mediaType: string;
  sizeBytes: number;
  url: string;
  attachmentKind?: "voice-note";
  durationMs?: number;
};
export type ChatMessageInput = {
  creationId: string;
  threadRootId?: string;
  text: string;
  richContent: JSONContent | null;
  mentionedUserIds: string[];
  agentMentions: "maya"[];
  attachmentIds: string[];
};
