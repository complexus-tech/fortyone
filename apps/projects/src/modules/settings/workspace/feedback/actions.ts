import { post, put, remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  CreateFeedbackBoardInput,
  FeedbackBoard,
  FeedbackPortal,
  FeedbackReviewer,
  UpdateFeedbackPortalInput,
  UpdateFeedbackReviewerInput,
} from "./types";

export const updateFeedbackPortal = async (
  portalId: string,
  input: UpdateFeedbackPortalInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<UpdateFeedbackPortalInput, ApiResponse<FeedbackPortal>>(
      `feedback/portals/${portalId}`,
      input,
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const createFeedbackBoard = async (
  input: CreateFeedbackBoardInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<CreateFeedbackBoardInput, ApiResponse<FeedbackBoard>>(
      "feedback/boards",
      input,
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const deleteFeedbackBoard = async (
  boardId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await remove<ApiResponse<null>>(`feedback/boards/${boardId}`, ctx);
  } catch (error) {
    return getApiError(error);
  }
};

export const updateFeedbackBoardReviewer = async (
  boardId: string,
  userId: string,
  input: UpdateFeedbackReviewerInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<
      UpdateFeedbackReviewerInput,
      ApiResponse<FeedbackReviewer>
    >(`feedback/boards/${boardId}/reviewers/${userId}`, input, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
