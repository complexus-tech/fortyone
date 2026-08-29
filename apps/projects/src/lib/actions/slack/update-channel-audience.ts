import { auth } from "@/auth";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { UpdateSlackChannelAudienceInput } from "@/modules/settings/workspace/integrations/slack/types";

export const updateSlackChannelAudienceAction = async (
  workspaceSlug: string,
  { channelId, isConfigured, teamIds }: UpdateSlackChannelAudienceInput,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<
      { isConfigured: boolean; teamIds: string[] },
      ApiResponse<null>
    >(
      `integrations/slack/channel-audiences/${encodeURIComponent(channelId)}`,
      { isConfigured, teamIds },
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
