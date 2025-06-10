export const statusKeys = {
  all: ["statuses"] as const,
  lists: () => [...statusKeys.all, "list"] as const,
  details: () => [...statusKeys.all, "detail"] as const,
  detail: (id: string) => [...statusKeys.details(), id] as const,
  team: (teamId: string) => [...statusKeys.lists(), teamId] as const,
};

export const sprintKeys = {
  all: ["sprints"] as const,
  lists: () => [...sprintKeys.all, "list"] as const,
  running: () => [...sprintKeys.all, "running"] as const,
  details: () => [...sprintKeys.all, "detail"] as const,
  detail: (id: string) => [...sprintKeys.details(), id] as const,
  team: (teamId: string) => [...sprintKeys.lists(), teamId] as const,
  objective: (objectiveId: string) =>
    [...sprintKeys.lists(), objectiveId] as const,
  analytics: (sprintId: string) =>
    [...sprintKeys.all, "analytics", sprintId] as const,
};

export const memberKeys = {
  all: ["members"] as const,
  lists: () => [...memberKeys.all, "list"] as const,
  details: () => [...memberKeys.all, "detail"] as const,
  detail: (id: string) => [...memberKeys.details(), id] as const,
  team: (teamId: string) => [...memberKeys.lists(), teamId] as const,
};

export const labelKeys = {
  all: ["labels"] as const,
  lists: () => [...labelKeys.all, "list"] as const,
  details: () => [...labelKeys.all, "detail"] as const,
  detail: (id: string) => [...labelKeys.details(), id] as const,
  team: (teamId: string) => [...labelKeys.all, teamId] as const,
};

export const linkKeys = {
  story: (storyId: string) => ["story-links", storyId] as const,
  metadata: (url: string) => ["link-metadata", url] as const,
};

export const linkTags = {
  metadata: (url: string) => `link-metadata-${url}` as const,
};

export const workspaceKeys = {
  all: ["workspaces"] as const,
  lists: () => [...workspaceKeys.all, "list"] as const,
  settings: () => [...workspaceKeys.all, "settings"] as const,
};

export const teamKeys = {
  all: ["teams"] as const,
  lists: () => [...teamKeys.all, "list"] as const,
  details: () => [...teamKeys.all, "detail"] as const,
  detail: (id: string) => [...teamKeys.details(), id] as const,
  public: () => [...teamKeys.lists(), "public"] as const,
  settings: (id: string) => [...teamKeys.all, "settings", id] as const,
};

export const userKeys = {
  all: ["users"] as const,
  profile: () => [...userKeys.all, "profile"] as const,
  automationPreferences: () =>
    [...userKeys.all, "automation-preferences"] as const,
};

export const invitationKeys = {
  pending: ["pending-invitations"] as const,
  mine: ["my-invitations"] as const,
};

export const notificationKeys = {
  all: ["notifications"] as const,
  unread: () => [...notificationKeys.all, "unread"] as const,
  preferences: () => [...notificationKeys.all, "preferences"] as const,
};

export const subscriptionKeys = {
  details: ["subscriptions"] as const,
};

export const analyticsKeys = {
  all: ["analytics"] as const,
  overview: (filters?: Record<string, unknown>) =>
    [...analyticsKeys.all, "overview", filters] as const,
  storyAnalytics: (filters?: Record<string, unknown>) =>
    [...analyticsKeys.all, "story-analytics", filters] as const,
  objectiveProgress: (filters?: Record<string, unknown>) =>
    [...analyticsKeys.all, "objective-progress", filters] as const,
  teamPerformance: (filters?: Record<string, unknown>) =>
    [...analyticsKeys.all, "team-performance", filters] as const,
  sprintAnalytics: (filters?: Record<string, unknown>) =>
    [...analyticsKeys.all, "sprint-analytics", filters] as const,
  timelineTrends: (filters?: Record<string, unknown>) =>
    [...analyticsKeys.all, "timeline-trends", filters] as const,
};
