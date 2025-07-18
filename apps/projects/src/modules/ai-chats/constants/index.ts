export const aiChatKeys = {
  all: ["ai-chats"] as const,
  lists: () => [...aiChatKeys.all, "list"] as const,
  details: () => [...aiChatKeys.all, "detail"] as const,
  detail: (id: string) => [...aiChatKeys.details(), id] as const,
  messages: (id: string) => [...aiChatKeys.detail(id), "messages"] as const,
};
