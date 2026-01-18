export const storyKeys = {
  all: (workspaceSlug: string) => ["stories", workspaceSlug] as const,
  lists: (workspaceSlug: string) =>
    [...storyKeys.all(workspaceSlug), "list"] as const,
  mine: (workspaceSlug: string) =>
    [...storyKeys.all(workspaceSlug), "list", "mine"] as const,
  teams: (workspaceSlug: string) =>
    [...storyKeys.all(workspaceSlug), "list", "team"] as const,
  team: (workspaceSlug: string, teamId: string) =>
    [...storyKeys.teams(workspaceSlug), teamId] as const,
  list: (workspaceSlug: string, filter: string) =>
    [...storyKeys.lists(workspaceSlug), filter] as const,
  details: (workspaceSlug: string) =>
    [...storyKeys.all(workspaceSlug), "detail"] as const,
  detail: (workspaceSlug: string, id: string) =>
    [...storyKeys.details(workspaceSlug), id] as const,
  activitiesInfinite: (workspaceSlug: string, id: string) =>
    [...storyKeys.detail(workspaceSlug, id), "activities", "infinite"] as const,
  objectives: (workspaceSlug: string) =>
    [...storyKeys.all(workspaceSlug), "objectives"] as const,
  objective: (workspaceSlug: string, objectiveId: string) =>
    [...storyKeys.objectives(workspaceSlug), objectiveId] as const,
  sprints: (workspaceSlug: string) =>
    [...storyKeys.all(workspaceSlug), "sprints"] as const,
  sprint: (workspaceSlug: string, sprintId: string) =>
    [...storyKeys.sprints(workspaceSlug), sprintId] as const,
  commentsInfinite: (workspaceSlug: string, id: string) =>
    [...storyKeys.detail(workspaceSlug, id), "comments", "infinite"] as const,
  attachments: (workspaceSlug: string, id: string) =>
    [...storyKeys.detail(workspaceSlug, id), "attachments"] as const,
  total: (workspaceSlug: string) =>
    ["totalStories", workspaceSlug] as const,
  // Grouped stories
  grouped: (workspaceSlug: string) =>
    [...storyKeys.all(workspaceSlug), "grouped"] as const,
  mineGrouped: (workspaceSlug: string, params: Record<string, unknown>) =>
    [...storyKeys.mine(workspaceSlug), "grouped", params] as const,
  teamGrouped: (
    workspaceSlug: string,
    teamId: string,
    params: Record<string, unknown>,
  ) => [...storyKeys.team(workspaceSlug, teamId), "grouped", params] as const,
  objectiveGrouped: (
    workspaceSlug: string,
    objectiveId: string,
    params: Record<string, unknown>,
  ) =>
    [
      ...storyKeys.objective(workspaceSlug, objectiveId),
      "grouped",
      params,
    ] as const,
  sprintGrouped: (
    workspaceSlug: string,
    sprintId: string,
    params: Record<string, unknown>,
  ) =>
    [...storyKeys.sprint(workspaceSlug, sprintId), "grouped", params] as const,
  groupStories: (
    workspaceSlug: string,
    groupKey: string,
    params: Record<string, unknown>,
  ) => [...storyKeys.all(workspaceSlug), "group", groupKey, params] as const,
};

