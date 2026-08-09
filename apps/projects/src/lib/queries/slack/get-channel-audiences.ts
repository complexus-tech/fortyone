import type { SlackChannelAudience } from "@/modules/settings/workspace/integrations/slack/types";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getSlackChannelAudiences = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<SlackChannelAudience[]>>(
    "integrations/slack/channel-audiences",
    ctx,
  );
  return response.data ?? [];
};
