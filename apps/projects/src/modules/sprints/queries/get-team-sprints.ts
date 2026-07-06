import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { Sprint, SprintsPage } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

const emptyTeamSprintsPage = (page = 1, pageSize = 15): SprintsPage => ({
  sprints: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: 0,
  },
});

export const getTeamSprintsPage = async (
  teamId: string,
  ctx: WorkspaceCtx,
  search = "",
  page = 1,
  pageSize = 15,
) => {
  if (!teamId) return emptyTeamSprintsPage(page, pageSize);

  const query = stringify(
    {
      search: search.trim() || undefined,
      teamId,
      page,
      pageSize,
    },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );

  const sprints = await get<ApiResponse<SprintsPage>>(`sprints${query}`, ctx);
  return sprints.data ?? emptyTeamSprintsPage(page, pageSize);
};

export const getTeamSprints = async (
  teamId: string,
  ctx: WorkspaceCtx,
  search = "",
) => {
  if (!teamId) return [];
  const query = stringify(
    { search: search.trim() || undefined, teamId },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );

  const sprints = await get<ApiResponse<Sprint[]>>(`sprints${query}`, ctx);
  return sprints.data!;
};
