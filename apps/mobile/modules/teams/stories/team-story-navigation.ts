type FilterKind = "sprint" | "objective";
type RouteParameter = string | string[] | undefined;

export type TeamStoriesRouteParams = {
  teamId: string;
  sprintId?: string;
  objectiveId?: string;
  openFilter?: FilterKind;
};

export const teamStoriesHref = (params: TeamStoriesRouteParams) => ({
  pathname: "/teams/[teamId]" as const,
  params,
});

const singleParameter = (value: RouteParameter) =>
  typeof value === "string" && value.trim() ? value : undefined;

/** Keep old links working without mounting the retired collection/detail pages. */
export function legacyTeamStoriesHref(
  params: {
    teamId?: RouteParameter;
    sprintId?: RouteParameter;
    objectiveId?: RouteParameter;
  },
  kind: FilterKind,
  collection = false,
) {
  const teamId = singleParameter(params.teamId);
  if (!teamId) return "/" as const;
  if (collection) return teamStoriesHref({ teamId, openFilter: kind });
  const id = singleParameter(
    kind === "sprint" ? params.sprintId : params.objectiveId,
  );
  // Do not silently broaden a malformed entity link into all of a team's work.
  if (!id) return "/" as const;
  return teamStoriesHref(
    kind === "sprint" ? { teamId, sprintId: id } : { teamId, objectiveId: id },
  );
}
