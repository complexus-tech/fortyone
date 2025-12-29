import type { UIMessage } from "ai";

export type AiChatSession = {
  id: string;
  userId: string;
  workspaceId: string;
  title: string;
  createdAt: string;
  updatedAt: string;
};

export type AiTotalChatMessages = {
  count: number;
};

export type CreateAiChatPayload = {
  id: string;
  title: string;
  messages: UIMessage[];
};

export type UpdateAiChatPayload = {
  title: string;
};

export type SaveMessagesPayload = {
  id: string;
  messages: UIMessage[];
};

export type Memory = {
  id: string;
  workspaceId: string;
  userId: string;
  content: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateMemoryPayload = {
  content: string;
};

export type UpdateMemoryPayload = {
  content: string;
};
