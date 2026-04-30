import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { IntegrationRequest } from "../types";

export const getTeamIntegrationRequests = async (
  teamId: string,
  ctx: WorkspaceCtx,
  status = "pending",
) => {
  const response = await get<ApiResponse<IntegrationRequest[]>>(
    `teams/${teamId}/integration-requests?status=${status}`,
    ctx,
  );

  return response.data ?? [];
};
