import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { get } from "@/lib/http";
import type { TeamFeedbackSummary } from "../types";

export const getTeamFeedbackSummaries = async (
  ctx: WorkspaceCtx,
): Promise<TeamFeedbackSummary[]> => {
  const response = await get<ApiResponse<TeamFeedbackSummary[]>>(
    "feedback/team-summaries",
    ctx,
  );

  if (response.error?.message) {
    throw new Error(response.error.message);
  }
  return response.data ?? [];
};
