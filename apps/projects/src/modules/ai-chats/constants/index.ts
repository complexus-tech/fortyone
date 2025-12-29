export const aiChatKeys = {
  all: ["ai-chats"] as const,
  lists: () => [...aiChatKeys.all, "list"] as const,
  details: () => [...aiChatKeys.all, "detail"] as const,
  detail: (id: string) => [...aiChatKeys.details(), id] as const,
  messages: (id: string) => [...aiChatKeys.detail(id), "messages"] as const,
  totalMessages: () => [...aiChatKeys.all, "total-messages"] as const,
  memory: () => [...aiChatKeys.all, "memory"] as const,
};
