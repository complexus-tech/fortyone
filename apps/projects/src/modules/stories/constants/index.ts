export const storyKeys = {
  all: ["stories"] as const,
  lists: () => [...storyKeys.all, "list"] as const,
  teams: () => [...storyKeys.all, "list", "team"] as const,
  mine: () => [...storyKeys.all, "list", "mine"] as const,
  team: (teamId: string) => [...storyKeys.teams(), teamId] as const,
  list: (filter: string) => [...storyKeys.lists(), filter] as const,
  details: () => [...storyKeys.all, "detail"] as const,
  detail: (id: string) => [...storyKeys.details(), id] as const,
};

export const storyTags = {
  all: "stories" as const,
  teams: () => `${storyTags.all}-list-team` as const,
  mine: () => `${storyTags.all}-list-mine` as const,
  team: (teamId: string) => `${storyTags.teams()}-${teamId}` as const,
  details: () => `${storyTags.all}-detail` as const,
  detail: (id: string) => `${storyTags.details()}-${id}` as const,
};
