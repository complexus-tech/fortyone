export const statusKeys = {
  all: (workspaceSlug: string) => ["statuses", workspaceSlug] as const,
  lists: (workspaceSlug: string) => [...statusKeys.all(workspaceSlug), "list"] as const,
  details: (workspaceSlug: string) => [...statusKeys.all(workspaceSlug), "detail"] as const,
  detail: (workspaceSlug: string, id: string) => [...statusKeys.details(workspaceSlug), id] as const,
  team: (workspaceSlug: string, teamId: string) => [...statusKeys.lists(workspaceSlug), teamId] as const,
};

export const sprintKeys = {
  all: (workspaceSlug: string) => ["sprints", workspaceSlug] as const,
  lists: (workspaceSlug: string) => [...sprintKeys.all(workspaceSlug), "list"] as const,
  running: (workspaceSlug: string) => [...sprintKeys.all(workspaceSlug), "running"] as const,
  details: (workspaceSlug: string) => [...sprintKeys.all(workspaceSlug), "detail"] as const,
  detail: (workspaceSlug: string, id: string) => [...sprintKeys.details(workspaceSlug), id] as const,
  team: (workspaceSlug: string, teamId: string) => [...sprintKeys.lists(workspaceSlug), teamId] as const,
  objective: (workspaceSlug: string, objectiveId: string) =>
    [...sprintKeys.lists(workspaceSlug), objectiveId] as const,
  analytics: (workspaceSlug: string, sprintId: string) =>
    [...sprintKeys.all(workspaceSlug), "analytics", sprintId] as const,
};

export const memberKeys = {
  all: (workspaceSlug: string) => ["members", workspaceSlug] as const,
  lists: (workspaceSlug: string) => [...memberKeys.all(workspaceSlug), "list"] as const,
  details: (workspaceSlug: string) => [...memberKeys.all(workspaceSlug), "detail"] as const,
  detail: (workspaceSlug: string, id: string) => [...memberKeys.details(workspaceSlug), id] as const,
  team: (workspaceSlug: string, teamId: string) => [...memberKeys.lists(workspaceSlug), teamId] as const,
};

export const labelKeys = {
  all: (workspaceSlug: string) => ["labels", workspaceSlug] as const,
  lists: (workspaceSlug: string) => [...labelKeys.all(workspaceSlug), "list"] as const,
  details: (workspaceSlug: string) => [...labelKeys.all(workspaceSlug), "detail"] as const,
  detail: (workspaceSlug: string, id: string) => [...labelKeys.details(workspaceSlug), id] as const,
  team: (workspaceSlug: string, teamId: string) => [...labelKeys.all(workspaceSlug), teamId] as const,
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
  settings: (workspaceSlug: string) => [...workspaceKeys.all, "settings", workspaceSlug] as const,
};

export const teamKeys = {
  all: (workspaceSlug: string) => ["teams", workspaceSlug] as const,
  lists: (workspaceSlug: string) => [...teamKeys.all(workspaceSlug), "list"] as const,
  details: (workspaceSlug: string) => [...teamKeys.all(workspaceSlug), "detail"] as const,
  detail: (workspaceSlug: string, id: string) => [...teamKeys.details(workspaceSlug), id] as const,
  public: (workspaceSlug: string) => [...teamKeys.lists(workspaceSlug), "public"] as const,
  settings: (workspaceSlug: string, id: string) => [...teamKeys.all(workspaceSlug), "settings", id] as const,
};

export const userKeys = {
  all: ["users"] as const,
  profile: () => [...userKeys.all, "profile"] as const,
  automationPreferences: (workspaceSlug: string) =>
    [...userKeys.all, "automation-preferences", workspaceSlug] as const,
};

export const invitationKeys = {
  all: ["invitations"] as const,
  pending: (workspaceSlug: string) =>
    [...invitationKeys.all, "pending", workspaceSlug] as const,
  mine: ["my-invitations"] as const,
};

export const notificationKeys = {
  all: (workspaceSlug: string) => ["notifications", workspaceSlug] as const,
  unread: (workspaceSlug: string) => [...notificationKeys.all(workspaceSlug), "unread"] as const,
  preferences: (workspaceSlug: string) => [...notificationKeys.all(workspaceSlug), "preferences"] as const,
};

export const subscriptionKeys = {
  details: (workspaceSlug: string) => ["subscriptions", workspaceSlug] as const,
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
