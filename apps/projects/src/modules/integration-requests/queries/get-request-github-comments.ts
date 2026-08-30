import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { IntegrationRequestGitHubComment } from "../types";

export const getRequestGitHubComments = async (
  requestId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<IntegrationRequestGitHubComment[]>>(
    `integration-requests/${requestId}/github-comments`,
    ctx,
  );
  return response.data ?? [];
};
