import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  FeedbackParticipationMode,
  FeedbackPortal,
  FeedbackReviewer,
} from "./types";

type ApiFeedbackPortal = Omit<FeedbackPortal, "participationMode"> & {
  participationMode?: FeedbackParticipationMode;
};

export const getFeedbackPortals = async (
  ctx: WorkspaceCtx,
): Promise<FeedbackPortal[]> => {
  const response = await get<ApiResponse<ApiFeedbackPortal[]>>(
    "feedback/portals",
    ctx,
  );
  return (response.data ?? []).map((portal) => ({
    ...portal,
    boards: portal.boards ?? [],
    participationMode: portal.participationMode ?? "account_required",
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
