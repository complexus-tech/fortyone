import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { TeamFeedbackItem } from "../types";

export const getTeamFeedbackItem = async (
  feedbackId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<TeamFeedbackItem>>(
    `feedback/items/${feedbackId}`,
    ctx,
  );

  if (response.error?.message) {
    throw new Error(response.error.message);
  }
  if (!response.data) {
    throw new Error("Feedback was not returned by the server");
  }
  return response.data;
};
