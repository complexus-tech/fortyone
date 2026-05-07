import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SlackIntegration } from "@/modules/settings/workspace/integrations/slack/types";

export const getSlackIntegration = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<SlackIntegration>>(
    "integrations/slack",
    ctx,
  );
  return response.data!;
};
