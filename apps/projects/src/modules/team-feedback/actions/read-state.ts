import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type TeamFeedbackReadState = {
  readAt: string | null;
};

export const setTeamFeedbackReadStateAction = async (
  feedbackId: string,
  isRead: boolean,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<Record<string, never>, ApiResponse<TeamFeedbackReadState>>(
      `feedback/items/${feedbackId}/${isRead ? "read" : "unread"}`,
      {},
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
