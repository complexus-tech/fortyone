export const homeKeys = {
  all: ["home"] as const,
  overview: () => [...homeKeys.all, "overview"] as const,
};

export const teamKeys = {
  all: ["teams"] as const,
  lists: () => [...teamKeys.all, "list"] as const,
  detail: (id: string) => [...teamKeys.all, "detail", id] as const,
};

export const storyKeys = {
  all: ["stories"] as const,
  lists: () => [...storyKeys.all, "list"] as const,
  mine: () => [...storyKeys.lists(), "mine"] as const,
  team: (teamId: string) => [...storyKeys.lists(), "team", teamId] as const,
  details: () => [...storyKeys.all, "detail"] as const,
  detail: (id: string) => [...storyKeys.details(), id] as const,
  grouped: (params: Record<string, any>) =>
    [...storyKeys.all, "grouped", params] as const,
  group: (groupKey: string, params: Record<string, any>) =>
    [...storyKeys.all, "group", groupKey, params] as const,
};
