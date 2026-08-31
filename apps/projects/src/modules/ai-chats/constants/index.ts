export const aiChatKeys = {
  all: ["ai-chats"] as const,
  lists: () => [...aiChatKeys.all, "list"] as const,
  details: () => [...aiChatKeys.all, "detail"] as const,
  detail: (id: string) => [...aiChatKeys.details(), id] as const,
  messages: (id: string) => [...aiChatKeys.detail(id), "messages"] as const,
  totalMessages: (workspaceSlug: string) =>
    [...aiChatKeys.all, "total-messages", workspaceSlug] as const,
  memories: () => [...aiChatKeys.all, "memories"] as const,
  memory: (id: string) => [...aiChatKeys.memories(), id] as const,
};
