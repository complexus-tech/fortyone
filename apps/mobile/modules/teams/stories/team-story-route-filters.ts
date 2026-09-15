import type { TeamStoryFilters } from "./team-story-filters";
import { createTeamStoryFilters } from "./team-story-filters";

type RouteParameter = string | string[];
type TeamStoryRouteParams = {
  sprintId?: RouteParameter;
  objectiveId?: RouteParameter;
  openFilter?: RouteParameter;
};

export type TeamStoryRouteFilters = {
  filters: TeamStoryFilters;
  initialFacet?: "sprint" | "objective";
  error?: string;
};

const UUID_PATTERN = /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i;
const NIL_UUID = "00000000-0000-0000-0000-000000000000";

const isEntityId = (value: unknown): value is string =>
  typeof value === "string" &&
  value.length === 36 &&
  UUID_PATTERN.test(value) &&
  value !== NIL_UUID;

/** An error is terminal for this route: render it instead of querying filters. */
export function parseTeamStoryRouteFilters(
  teamId: string,
  params: TeamStoryRouteParams,
): TeamStoryRouteFilters {
  const filters = createTeamStoryFilters(teamId);
  const { sprintId, objectiveId, openFilter } = params;

  // Check all supplied parameters before applying any part of the route. A bad
  // objective must not silently turn a combined link into a sprint-only query.
  if (sprintId !== undefined && !isEntityId(sprintId)) {
    return { filters, error: "This link contains an invalid sprint filter." };
  }
  if (objectiveId !== undefined && !isEntityId(objectiveId)) {
    return {
      filters,
      error: "This link contains an invalid objective filter.",
    };
  }
  if (
    openFilter !== undefined &&
    openFilter !== "sprint" &&
    openFilter !== "objective"
  ) {
    return {
      filters,
      error: "This link contains an invalid filter selection.",
    };
  }

  if (sprintId !== undefined) filters.sprintIds = [sprintId.toLowerCase()];
  if (objectiveId !== undefined)
    filters.objectiveId = objectiveId.toLowerCase();
  return { filters, ...(openFilter ? { initialFacet: openFilter } : {}) };
}
