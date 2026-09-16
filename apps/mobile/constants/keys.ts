import { useAuthStore } from "@/store/auth";
import { scopedQueryKey } from "@/lib/query-scope";

// The session provider remounts all observers when this scope changes.
const resourceKey = (resource: string) =>
  scopedQueryKey(useAuthStore.getState(), resource);

export const homeKeys = {
  get all() {
    return resourceKey("home");
  },
  overview: () => [...homeKeys.all, "overview"] as const,
};

export const teamKeys = {
  get all() {
    return resourceKey("teams");
  },
  lists: () => [...teamKeys.all, "list"] as const,
  detail: (id: string) => [...teamKeys.all, "detail", id] as const,
};

export const userKeys = {
  get all() {
    return resourceKey("users");
  },
  profile: () => [...userKeys.all, "profile"] as const,
};

export const memberKeys = {
  get all() {
    return resourceKey("members");
  },
  lists: () => [...memberKeys.all, "list"] as const,
  maya: () => [...memberKeys.all, "maya"] as const,
  details: () => [...memberKeys.all, "detail"] as const,
  detail: (id: string) => [...memberKeys.details(), id] as const,
  team: (teamId: string) => [...memberKeys.lists(), teamId] as const,
};

export const storyKeys = {
  get all() {
    return resourceKey("stories");
  },
  lists: () => [...storyKeys.all, "list"] as const,
  mine: () => [...storyKeys.lists(), "mine"] as const,
  team: (teamId: string) => [...storyKeys.lists(), "team", teamId] as const,
  details: () => [...storyKeys.all, "detail"] as const,
  detail: (id: string) => [...storyKeys.details(), id] as const,
  grouped: (params: Record<string, any>) =>
    [...storyKeys.all, "grouped", params] as const,
  mineGrouped: (params: Record<string, any>) =>
    [...storyKeys.mine(), "grouped", params] as const,
  teamGrouped: (teamId: string, params: Record<string, any>) =>
    [...storyKeys.team(teamId), "grouped", params] as const,
  sprintGrouped: (sprintId: string, params: Record<string, any>) =>
    [...storyKeys.all, "sprint", sprintId, "grouped", params] as const,
  objectiveGrouped: (objectiveId: string, params: Record<string, any>) =>
    [...storyKeys.all, "objective", objectiveId, "grouped", params] as const,
  group: (groupKey: string, params: Record<string, any>) =>
    [...storyKeys.all, "group", groupKey, params] as const,
  attachments: (storyId: string) =>
    [...storyKeys.detail(storyId), "attachments"] as const,
  activitiesInfinite: (storyId: string) =>
    [...storyKeys.detail(storyId), "activities", "infinite"] as const,
  commentsInfinite: (storyId: string) =>
    [...storyKeys.detail(storyId), "comments", "infinite"] as const,
};

export const notificationKeys = {
  get all() {
    return resourceKey("notifications");
  },
  lists: () => [...notificationKeys.all, "list"] as const,
  unread: () => [...notificationKeys.all, "unread"] as const,
  preferences: () => [...notificationKeys.all, "preferences"] as const,
};

export const workspaceKeys = {
  get all() {
    return resourceKey("workspace");
  },
  lists: () => [...workspaceKeys.all, "list"] as const,
  settings: () => [...workspaceKeys.all, "settings"] as const,
};

export const sprintKeys = {
  get all() {
    return resourceKey("sprints");
  },
  lists: () => [...sprintKeys.all, "list"] as const,
  details: () => [...sprintKeys.all, "detail"] as const,
  detail: (id: string) => [...sprintKeys.details(), id] as const,
  team: (teamId: string) => [...sprintKeys.lists(), "team", teamId] as const,
  running: () => [...sprintKeys.all, "running"] as const,
};

export const objectiveKeys = {
  get all() {
    return resourceKey("objectives");
  },
  lists: () => [...objectiveKeys.all, "list"] as const,
  details: () => [...objectiveKeys.all, "detail"] as const,
  detail: (id: string) => [...objectiveKeys.details(), id] as const,
  team: (teamId: string) => [...objectiveKeys.lists(), "team", teamId] as const,
  statuses: () => [...objectiveKeys.all, "statuses"] as const,
};

export const statusKeys = {
  get all() {
    return resourceKey("statuses");
  },
  lists: () => [...statusKeys.all, "list"] as const,
  team: (teamId: string) => [...statusKeys.lists(), "team", teamId] as const,
};

export const labelKeys = {
  get all() {
    return resourceKey("labels");
  },
  lists: () => [...labelKeys.all, "list"] as const,
  details: () => [...labelKeys.all, "detail"] as const,
  detail: (id: string) => [...labelKeys.details(), id] as const,
  team: (teamId: string) => [...labelKeys.all, teamId] as const,
};

export const searchKeys = {
  get all() {
    return resourceKey("search");
  },
  query: (params: Record<string, any>) => [...searchKeys.all, params] as const,
};

export const subscriptionKeys = {
  get details() {
    return resourceKey("subscriptions");
  },
};

export const feedbackKeys = {
  get all() {
    return resourceKey("feedback");
  },
  summaries: () => [...feedbackKeys.all, "team-summaries"] as const,
  team: (id: string) => [...feedbackKeys.all, "team", id, "active"] as const,
  detail: (id: string) => [...feedbackKeys.all, "detail", id] as const,
};

export const intakeKeys = {
  get all() {
    return resourceKey("intake");
  },
  team: (id: string) => [...intakeKeys.all, "team", id, "pending"] as const,
  detail: (id: string) => [...intakeKeys.all, "detail", id] as const,
};
