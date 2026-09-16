import type { UIMessage } from "ai";

export type MayaMessage = UIMessage;

export type MayaSession = {
  id: string;
  userId: string;
  workspaceId: string;
  title: string;
  createdAt: string;
  updatedAt: string;
};

export type MayaApproval = {
  id: string;
  toolCallId: string;
  toolName: string;
  input: unknown;
};

export type MayaScreenContext = {
  storyId?: string;
  storyReference?: string;
};
