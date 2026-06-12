import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team, TeamsPage } from "../types";

const emptyPublicTeamsPage = (page = 1, pageSize = 15): TeamsPage => ({
  teams: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: 0,
  },
});

const buildPublicTeamsQuery = ({
  page,
  pageSize,
  search,
}: {
  page?: number;
  pageSize?: number;
  search?: string;
}) => {
  const params = new URLSearchParams();
  if (page) {
    params.set("page", String(page));
  }
  if (pageSize) {
    params.set("pageSize", String(pageSize));
  }
  if (search?.trim()) {
    params.set("search", search.trim());
  }

  const query = params.toString();
  return query ? `?${query}` : "";
};

export const getPublicTeams = async (ctx: WorkspaceCtx): Promise<Team[]> => {
  const response = await get<ApiResponse<Team[]>>("teams/public", ctx);
  return response.data!;
};

export const getPublicTeamsPage = async (
  ctx: WorkspaceCtx,
  search = "",
  page = 1,
  pageSize = 15,
) => {
  const response = await get<ApiResponse<TeamsPage>>(
    `teams/public${buildPublicTeamsQuery({ page, pageSize, search })}`,
    ctx,
  );
  return response.data ?? emptyPublicTeamsPage(page, pageSize);
};
