import { auth } from "@/auth";
import { post, remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const trashTeamFeedbackAction = async (
  feedbackId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await remove<ApiResponse<null>>(`feedback/items/${feedbackId}`, ctx);
  } catch (error) {
    return getApiError(error);
  }
};

export const restoreTeamFeedbackAction = async (
  feedbackId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<null, ApiResponse<null>>(
      `feedback/items/${feedbackId}/restore`,
      null,
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
