import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SlackRequestLog } from "@/modules/settings/workspace/integrations/slack/types";

export const getSlackRequestLogs = async (ctx: WorkspaceCtx, limit = 20) => {
  const response = await get<ApiResponse<SlackRequestLog[]>>(
    `integrations/slack/logs?limit=${limit}`,
    ctx,
  );
  return response.data ?? [];
};
