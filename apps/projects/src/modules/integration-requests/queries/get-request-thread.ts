import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { IntegrationRequestThreadActivity } from "../types";

export const getIntegrationRequestThread = async (
  requestId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<IntegrationRequestThreadActivity>>(
    `integration-requests/${requestId}/thread`,
    ctx,
  );
  return response.data!;
};
