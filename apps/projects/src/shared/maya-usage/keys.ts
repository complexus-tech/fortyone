export const mayaUsageKeys = {
  totalMessages: (workspaceSlug: string) =>
    ["ai-chats", "workspace", workspaceSlug, "total-messages"] as const,
};
