import { auth } from "@/auth";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type UpdateSlackChannelAudienceInput = {
  teamIds: string[];
};

export const updateSlackChannelAudienceAction = async (
  workspaceSlug: string,
  channelId: string,
  input: UpdateSlackChannelAudienceInput,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<UpdateSlackChannelAudienceInput, ApiResponse<null>>(
      `integrations/slack/channel-audiences/${encodeURIComponent(channelId)}`,
      input,
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
