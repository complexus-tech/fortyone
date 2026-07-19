import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { TeamFeedbackItem, UpdateTeamFeedbackStatusInput } from "../types";

export const updateTeamFeedbackStatusAction = async (
  feedbackId: string,
  payload: UpdateTeamFeedbackStatusInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<
      UpdateTeamFeedbackStatusInput,
      ApiResponse<TeamFeedbackItem>
    >(`feedback/items/${feedbackId}/status`, payload, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
