export const CHAT_PAGE_SIZE = 50;
export const CHAT_MAX_PAGES = 8;

export const chatKeys = {
  all: (workspaceSlug: string) =>
    ["chat", workspaceSlug.toLowerCase()] as const,
  conversations: (workspaceSlug: string, search = "") =>
    [...chatKeys.all(workspaceSlug), "conversations", search] as const,
  conversation: (workspaceSlug: string, conversationId: string) =>
    [...chatKeys.all(workspaceSlug), "conversation", conversationId] as const,
  messages: (
    workspaceSlug: string,
    conversationId: string,
    anchor = "latest",
  ) =>
    [
      ...chatKeys.all(workspaceSlug),
      "messages",
      conversationId,
      anchor,
    ] as const,
  replies: (workspaceSlug: string, conversationId: string, rootId: string) =>
    [
      ...chatKeys.all(workspaceSlug),
      "replies",
      conversationId,
      rootId,
    ] as const,
  search: (workspaceSlug: string, conversationId: string, search: string) =>
    [...chatKeys.all(workspaceSlug), "search", conversationId, search] as const,
  resources: (workspaceSlug: string, conversationId: string, kind: string) =>
    [
      ...chatKeys.all(workspaceSlug),
      "resources",
      conversationId,
      kind,
    ] as const,
  mayaRuns: (workspaceSlug: string, conversationId: string) =>
    [...chatKeys.all(workspaceSlug), "maya-runs", conversationId] as const,
};
