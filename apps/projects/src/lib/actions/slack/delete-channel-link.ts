import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteSlackChannelLinkAction = async (
  linkId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await remove<ApiResponse<null>>(
      `integrations/slack/channel-links/${linkId}`,
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
