import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { TeamFeedbackListStatus, TeamFeedbackPage } from "../types";

export const getTeamFeedbackPage = async (
  teamId: string,
  ctx: WorkspaceCtx,
  status: TeamFeedbackListStatus = "active",
  search = "",
  page = 1,
  pageSize = 25,
) => {
  const params = new URLSearchParams({
    status,
    page: String(page),
    pageSize: String(pageSize),
  });
  if (search.trim()) params.set("search", search.trim());
  const response = await get<ApiResponse<TeamFeedbackPage>>(
    `teams/${teamId}/feedback?${params.toString()}`,
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

export const getTeamFeedback = async (
  teamId: string,
  ctx: WorkspaceCtx,
  status: TeamFeedbackListStatus = "active",
  search = "",
) => {
  const page = await getTeamFeedbackPage(teamId, ctx, status, search);
  return page.feedback;
};
