import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  IntegrationRequest,
  IntegrationRequestProvider,
  IntegrationRequestStatus,
  IntegrationRequestsPage,
} from "../types";

export type IntegrationRequestListFilters = {
  search?: string;
  status?: IntegrationRequestStatus;
  provider?: IntegrationRequestProvider;
  priority?: IntegrationRequest["priority"];
  assigneeId?: string;
  createdAfter?: string;
  createdBefore?: string;
};

const emptyRequestsPage = (
  page = 1,
  pageSize = 25,
): IntegrationRequestsPage => ({
  requests: [],
  pagination: {
    page,
    pageSize,
    totalCount: 0,
    hasMore: false,
    nextPage: page + 1,
  },
});

export const getTeamIntegrationRequestsPage = async (
  teamId: string,
  ctx: WorkspaceCtx,
  status: IntegrationRequestStatus = "pending",
  page = 1,
  pageSize = 25,
  filters: Omit<IntegrationRequestListFilters, "status"> = {},
) => {
  const params = new URLSearchParams({
    status,
    page: String(page),
    pageSize: String(pageSize),
  });
  if (filters.search?.trim()) params.set("search", filters.search.trim());
  if (filters.provider) params.set("provider", filters.provider);
  if (filters.priority) params.set("priority", filters.priority);
  if (filters.assigneeId) params.set("assigneeId", filters.assigneeId);
  if (filters.createdAfter) params.set("createdAfter", filters.createdAfter);
  if (filters.createdBefore) params.set("createdBefore", filters.createdBefore);

  const response = await get<ApiResponse<IntegrationRequestsPage>>(
    `teams/${teamId}/integration-requests?${params.toString()}`,
    ctx,
  );

  return response.data ?? emptyRequestsPage(page, pageSize);
};

export const getTeamIntegrationRequests = async (
  teamId: string,
  ctx: WorkspaceCtx,
  status: IntegrationRequestStatus = "pending",
) => {
  return getTeamIntegrationRequestsPage(teamId, ctx, status);
};
