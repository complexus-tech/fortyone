import { mayaUsageKeys } from "@/shared/maya-usage/keys";

export const aiChatKeys = {
  all: ["ai-chats"] as const,
  workspace: (workspaceSlug: string) =>
    [...aiChatKeys.all, "workspace", workspaceSlug] as const,
  lists: (workspaceSlug: string) =>
    [...aiChatKeys.workspace(workspaceSlug), "list"] as const,
  details: (workspaceSlug: string) =>
    [...aiChatKeys.workspace(workspaceSlug), "detail"] as const,
  detail: (workspaceSlug: string, id: string) =>
    [...aiChatKeys.details(workspaceSlug), id] as const,
  messages: (workspaceSlug: string, id: string) =>
    [...aiChatKeys.detail(workspaceSlug, id), "messages"] as const,
  totalMessages: mayaUsageKeys.totalMessages,
  memories: (workspaceSlug: string) =>
    [...aiChatKeys.workspace(workspaceSlug), "memories"] as const,
  memory: (workspaceSlug: string, id: string) =>
    [...aiChatKeys.memories(workspaceSlug), id] as const,
};
