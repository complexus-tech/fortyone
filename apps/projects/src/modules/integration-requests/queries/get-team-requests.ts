import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { IntegrationRequest, IntegrationRequestsPage } from "../types";

const emptyRequestsPage = (
  page = 1,
  pageSize = 25,
): IntegrationRequestsPage => ({
  requests: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: page + 1,
  },
});

export const getTeamIntegrationRequestsPage = async (
  teamId: string,
  ctx: WorkspaceCtx,
  status = "pending",
  page = 1,
  pageSize = 25,
) => {
  const response = await get<ApiResponse<IntegrationRequestsPage>>(
    `teams/${teamId}/integration-requests?status=${status}&page=${page}&pageSize=${pageSize}`,
    ctx,
  );

  return response.data ?? emptyRequestsPage(page, pageSize);
};

export const getTeamIntegrationRequests = async (
  teamId: string,
  ctx: WorkspaceCtx,
  status = "pending",
) => {
  const page = await getTeamIntegrationRequestsPage(teamId, ctx, status);
  return page.requests;
};
