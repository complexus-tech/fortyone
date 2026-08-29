import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team, TeamsPage } from "../types";

const emptyTeamsPage = (page = 1, pageSize = 15): TeamsPage => ({
  teams: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: 0,
  },
});

const buildTeamsQuery = ({
  page,
  pageSize,
  search,
  joinedOnly,
}: {
  page?: number;
  pageSize?: number;
  search?: string;
  joinedOnly?: boolean;
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
  if (joinedOnly) {
    params.set("joinedOnly", "true");
  }

  const query = params.toString();
  return query ? `?${query}` : "";
};

export const getTeams = async (ctx: WorkspaceCtx): Promise<Team[]> => {
  const response = await get<ApiResponse<Team[]>>("teams", ctx);
  return response.data!;
};

export const getJoinedTeams = async (ctx: WorkspaceCtx): Promise<Team[]> => {
  const response = await get<ApiResponse<Team[]>>("teams?joinedOnly=true", ctx);
  return response.data!;
};

export const getTeamsPage = async (
  ctx: WorkspaceCtx,
  search = "",
  page = 1,
  pageSize = 15,
) => {
  const response = await get<ApiResponse<TeamsPage>>(
    `teams${buildTeamsQuery({ page, pageSize, search })}`,
    ctx,
  );
  return response.data ?? emptyTeamsPage(page, pageSize);
};

export const getJoinedTeamsPage = async (
  ctx: WorkspaceCtx,
  search = "",
  page = 1,
  pageSize = 15,
) => {
  const response = await get<ApiResponse<TeamsPage>>(
    `teams${buildTeamsQuery({
      page,
      pageSize,
      search,
      joinedOnly: true,
    })}`,
    ctx,
  );
  return response.data ?? emptyTeamsPage(page, pageSize);
};
