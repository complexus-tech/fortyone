export const documentKeys = {
  all: (workspaceSlug: string) => ["documents", workspaceSlug] as const,
  lists: (workspaceSlug: string) =>
    [...documentKeys.all(workspaceSlug), "list"] as const,
  list: (workspaceSlug: string, search = "", scope = "all", limit?: number) =>
    [
      ...documentKeys.lists(workspaceSlug),
      { limit: limit ?? null, scope, search },
    ] as const,
  details: (workspaceSlug: string) =>
    [...documentKeys.all(workspaceSlug), "detail"] as const,
  detail: (workspaceSlug: string, documentId: string) =>
    [...documentKeys.details(workspaceSlug), documentId] as const,
  related: (
    workspaceSlug: string,
    entityType: "story" | "objective",
    entityId: string,
  ) =>
    [
      ...documentKeys.all(workspaceSlug),
      "related",
      entityType,
      entityId,
    ] as const,
};
