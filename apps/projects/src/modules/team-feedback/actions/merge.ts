import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { TeamFeedbackMergeResult } from "../types";

export const mergeTeamFeedbackAction = async (
  sourceItemId: string,
  targetItemId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      { targetItemId: string },
      ApiResponse<TeamFeedbackMergeResult>
    >(
      `feedback/items/${encodeURIComponent(sourceItemId)}/merge`,
      { targetItemId },
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
