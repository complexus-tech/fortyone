import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { PlanTeamFeedbackInput, PlanTeamFeedbackResult } from "../types";

export const planTeamFeedbackAction = async (
  feedbackId: string,
  payload: PlanTeamFeedbackInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      PlanTeamFeedbackInput,
      ApiResponse<PlanTeamFeedbackResult>
    >(`feedback/items/${feedbackId}/story`, payload, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
