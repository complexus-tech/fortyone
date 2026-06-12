import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective, ObjectivesPage } from "../types";

const emptyTeamObjectivesPage = (
  page = 1,
  pageSize = 15,
): ObjectivesPage => ({
  objectives: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: 0,
  },
});

export const getTeamObjectivesPage = async (
  teamId: string,
  ctx: WorkspaceCtx,
  search = "",
  page = 1,
  pageSize = 15,
) => {
  if (!teamId) return emptyTeamObjectivesPage(page, pageSize);

  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
    teamId,
  });
  if (search.trim()) {
    params.set("search", search.trim());
  }

  const objectives = await get<ApiResponse<ObjectivesPage>>(
    `objectives?${params.toString()}`,
    ctx,
  );
  return objectives.data ?? emptyTeamObjectivesPage(page, pageSize);
};

export const getTeamObjectives = async (
  teamId: string,
  ctx: WorkspaceCtx,
  search = "",
) => {
  const params = new URLSearchParams({ teamId });
  if (search.trim()) {
    params.set("search", search.trim());
  }
  const objectives = await get<ApiResponse<Objective[]>>(
    `objectives?${params.toString()}`,
    ctx,
  );
  return objectives.data ?? [];
};
