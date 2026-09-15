import type {
  GroupedStoryParams,
  StoryPriority,
} from "@/modules/stories/types";
import type { StoriesViewOptions } from "@/types/stories-view-options";
import type { TeamStoriesTab } from "../types";

export type TeamStoryAssigneeFilter =
  | { kind: "any" }
  | { kind: "unassigned" }
  | { kind: "members"; ids: string[] };

export type TeamStoryFilters = {
  // Also accepts a local owner key such as "my-work:assigned" when the facets
  // are composed with a different story-list scope instead of a team query.
  teamId: string;
  statusIds: string[];
  priorities: StoryPriority[];
  sprintIds: string[];
  // Null means any objective. The API has no "without an objective" filter.
  objectiveId: string | null;
  assignee: TeamStoryAssigneeFilter;
};

// Filters are local to the current team screen. Resetting does not change its
// tab or persisted display preferences, and creates independent selection arrays.
export function createTeamStoryFilters(teamId: string): TeamStoryFilters {
  return {
    teamId,
    statusIds: [],
    priorities: [],
    sprintIds: [],
    objectiveId: null,
    assignee: { kind: "any" },
  };
}

export function getTeamStoryFilters(
  teamId: string,
  filters: TeamStoryFilters,
): TeamStoryFilters {
  return filters.teamId === teamId ? filters : createTeamStoryFilters(teamId);
}

// A facet counts once, regardless of how many values are selected within it.
export function getTeamStoryFilterCount(filters: TeamStoryFilters): number {
  return [
    filters.statusIds.length > 0,
    filters.priorities.length > 0,
    filters.sprintIds.length > 0,
    filters.objectiveId !== null,
    filters.assignee.kind === "unassigned" ||
      (filters.assignee.kind === "members" && filters.assignee.ids.length > 0),
  ].filter(Boolean).length;
}

function canonicalValues<T extends string>(values: readonly T[]): T[] {
  return [...new Set(values)].sort();
}

export type StoryFilterQueryParams = Pick<
  GroupedStoryParams,
  | "statusIds"
  | "priorities"
  | "sprintIds"
  | "objectiveId"
  | "hasNoAssignee"
  | "assigneeIds"
>;

// Reusable for My Work and other lists: does not supply a team, tab, ownership,
// grouping, or order constraint. The caller owns those scopes independently.
export function buildStoryFilterParams(
  filters: TeamStoryFilters,
): StoryFilterQueryParams {
  const params: StoryFilterQueryParams = {};
  if (filters.statusIds.length) {
    params.statusIds = canonicalValues(filters.statusIds);
  }
  if (filters.priorities.length) {
    params.priorities = canonicalValues(filters.priorities);
  }
  if (filters.sprintIds.length) {
    params.sprintIds = canonicalValues(filters.sprintIds);
  }
  if (filters.objectiveId !== null) params.objectiveId = filters.objectiveId;
  if (filters.assignee.kind === "unassigned") {
    params.hasNoAssignee = true;
  } else if (
    filters.assignee.kind === "members" &&
    filters.assignee.ids.length
  ) {
    params.assigneeIds = canonicalValues(filters.assignee.ids);
  }
  return params;
}

export function buildTeamStoryQuery({
  teamId,
  filters,
  tab,
  viewOptions,
}: {
  teamId: string;
  filters: TeamStoryFilters;
  tab: TeamStoriesTab;
  viewOptions: Pick<
    StoriesViewOptions,
    "groupBy" | "orderBy" | "orderDirection"
  >;
}): GroupedStoryParams {
  const scoped = getTeamStoryFilters(teamId, filters);
  const params: GroupedStoryParams = {
    teamIds: [teamId],
    groupBy: viewOptions.groupBy,
    orderBy: viewOptions.orderBy,
    orderDirection: viewOptions.orderDirection,
    ...buildStoryFilterParams(scoped),
  };

  // Categories and explicit statuses are intersections on the server. Keeping
  // both ensures selecting a status cannot silently change the selected tab.
  if (tab === "active") params.categories = ["started"];
  if (tab === "backlog") params.categories = ["backlog"];
  return params;
}
