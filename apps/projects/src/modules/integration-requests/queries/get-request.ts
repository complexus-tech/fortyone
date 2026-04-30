import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { IntegrationRequest } from "../types";

export const getIntegrationRequest = async (
  requestId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<IntegrationRequest>>(
    `integration-requests/${requestId}`,
    ctx,
  );

  return response.data!;
};
