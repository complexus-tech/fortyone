import type { SlackAgentSettings } from "@/modules/settings/workspace/integrations/slack/types";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getSlackAgentSettings = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<SlackAgentSettings>>(
    "integrations/slack/agent-settings",
    ctx,
  );
  return response.data;
};
