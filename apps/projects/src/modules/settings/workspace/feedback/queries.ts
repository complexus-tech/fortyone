import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { FeedbackPortal, FeedbackReviewer } from "./types";

export const getFeedbackPortals = async (
  ctx: WorkspaceCtx,
): Promise<FeedbackPortal[]> => {
  const response = await get<ApiResponse<FeedbackPortal[]>>(
    "feedback/portals",
    ctx,
  );
  return (response.data ?? []).map((portal) => ({
    ...portal,
    boards: portal.boards ?? [],
  }));
};

export const getFeedbackBoardReviewers = async (
  boardId: string,
  ctx: WorkspaceCtx,
): Promise<FeedbackReviewer[]> => {
  const response = await get<ApiResponse<FeedbackReviewer[]>>(
    `feedback/boards/${boardId}/reviewers`,
    ctx,
  );
  return response.data ?? [];
};
