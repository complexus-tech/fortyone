import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  CreateTeamFeedbackCommentInput,
  TeamFeedbackComment,
} from "../types";

export const createTeamFeedbackCommentAction = async (
  feedbackId: string,
  payload: CreateTeamFeedbackCommentInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      CreateTeamFeedbackCommentInput,
      ApiResponse<TeamFeedbackComment>
    >(`feedback/items/${feedbackId}/comments`, payload, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
