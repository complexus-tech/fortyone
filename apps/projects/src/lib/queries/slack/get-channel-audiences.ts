import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SlackChannelAudience } from "@/modules/settings/workspace/integrations/slack/types";

export const getSlackChannelAudiences = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<SlackChannelAudience[]>>(
    "integrations/slack/channel-audiences",
    ctx,
  );
  const audiences = response.data ?? [];
  const hasExplicitConfigurationState = audiences.every(
    (audience) =>
      typeof (audience as { isConfigured?: unknown }).isConfigured ===
      "boolean",
  );
  if (!hasExplicitConfigurationState) {
    throw new Error("Slack channel configuration is not available");
  }
  return audiences;
};
