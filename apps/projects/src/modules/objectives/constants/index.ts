export const objectiveKeys = {
  all: (workspaceSlug: string) => ["objectives", workspaceSlug] as const,
  statuses: (workspaceSlug: string) => [...objectiveKeys.all(workspaceSlug), "statuses"] as const,
  list: (workspaceSlug: string) => [...objectiveKeys.all(workspaceSlug), "list"] as const,
  team: (workspaceSlug: string, teamId: string) => [...objectiveKeys.all(workspaceSlug), "list", teamId] as const,
  objective: (workspaceSlug: string, objectiveId: string) =>
    [...objectiveKeys.all(workspaceSlug), "list", objectiveId] as const,
  keyResults: (workspaceSlug: string, objectiveId: string) =>
    [...objectiveKeys.all(workspaceSlug), objectiveId, "key-results"] as const,
  analytics: (workspaceSlug: string, objectiveId: string) =>
    [...objectiveKeys.all(workspaceSlug), "analytics", objectiveId] as const,
  activitiesInfinite: (workspaceSlug: string, objectiveId: string) =>
    [
      ...objectiveKeys.objective(workspaceSlug, objectiveId),
      "activities",
      "infinite",
    ] as const,
};
