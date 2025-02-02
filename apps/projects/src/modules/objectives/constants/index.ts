export const objectiveKeys = {
  all: ["objectives"] as const,
  statuses: () => [...objectiveKeys.all, "statuses"] as const,
  list: () => [...objectiveKeys.all, "list"] as const,
  team: (teamId: string) => [...objectiveKeys.all, "list", teamId] as const,
  objective: (objectiveId: string) =>
    [...objectiveKeys.all, "list", objectiveId] as const,
  keyResults: (objectiveId: string) =>
    [...objectiveKeys.all, objectiveId, "key-results"] as const,
};

export const objectiveTags = {
  all: "objectives" as const,
  statuses: () => `${objectiveTags.all}-statuses` as const,
  list: () => `${objectiveTags.all}-list` as const,
  team: (teamId: string) => `${objectiveTags.list()}-${teamId}` as const,
  objective: (objectiveId: string) =>
    `${objectiveTags.list()}-${objectiveId}` as const,
  keyResults: (objectiveId: string) =>
    `${objectiveTags.objective(objectiveId)}-key-results` as const,
};
