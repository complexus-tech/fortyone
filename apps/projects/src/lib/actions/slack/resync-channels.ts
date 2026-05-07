import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const resyncSlackChannelsAction = async (workspaceSlug: string) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<Record<string, never>, ApiResponse<null>>(
      "integrations/slack/channels/resync",
      {},
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
