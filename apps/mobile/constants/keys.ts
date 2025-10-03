export const homeKeys = {
  all: ["home"] as const,
  overview: () => [...homeKeys.all, "overview"] as const,
};

export const teamKeys = {
  all: ["teams"] as const,
  lists: () => [...teamKeys.all, "list"] as const,
  detail: (id: string) => [...teamKeys.all, "detail", id] as const,
};

export const userKeys = {
  all: ["users"] as const,
  profile: () => [...userKeys.all, "profile"] as const,
};

export const memberKeys = {
  all: ["members"] as const,
  lists: () => [...memberKeys.all, "list"] as const,
  details: () => [...memberKeys.all, "detail"] as const,
  detail: (id: string) => [...memberKeys.details(), id] as const,
  team: (teamId: string) => [...memberKeys.lists(), teamId] as const,
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

export const notificationKeys = {
  all: ["notifications"] as const,
  unread: () => [...notificationKeys.all, "unread"] as const,
  preferences: () => [...notificationKeys.all, "preferences"] as const,
};

export const sprintKeys = {
  all: ["sprints"] as const,
  lists: () => [...sprintKeys.all, "list"] as const,
  details: () => [...sprintKeys.all, "detail"] as const,
  detail: (id: string) => [...sprintKeys.details(), id] as const,
  team: (teamId: string) => [...sprintKeys.lists(), "team", teamId] as const,
  running: () => [...sprintKeys.all, "running"] as const,
};

export const objectiveKeys = {
  all: ["objectives"] as const,
  lists: () => [...objectiveKeys.all, "list"] as const,
  details: () => [...objectiveKeys.all, "detail"] as const,
  detail: (id: string) => [...objectiveKeys.details(), id] as const,
  team: (teamId: string) => [...objectiveKeys.lists(), "team", teamId] as const,
  statuses: () => [...objectiveKeys.all, "statuses"] as const,
};

export const statusKeys = {
  all: ["statuses"] as const,
  lists: () => [...statusKeys.all, "list"] as const,
  team: (teamId: string) => [...statusKeys.lists(), "team", teamId] as const,
};
