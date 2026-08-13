import { post, put, remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  CreateFeedbackBoardInput,
  FeedbackBoard,
  FeedbackPortal,
  FeedbackReviewer,
  FeedbackUpdate,
  FeedbackWidgetSettings,
  FeedbackWidgetSigningSecret,
  UpsertFeedbackUpdateInput,
  UpdateFeedbackPortalInput,
  UpdateFeedbackReviewerInput,
  UpdateFeedbackWidgetSettingsInput,
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

const getWorkspaceContext = async (workspaceSlug: string) => {
  const session = await auth();
  return { session: session!, workspaceSlug };
};

export const updateFeedbackWidgetSettings = async (
  portalId: string,
  input: UpdateFeedbackWidgetSettingsInput,
  workspaceSlug: string,
) => {
  try {
    return await put<
      UpdateFeedbackWidgetSettingsInput,
      ApiResponse<FeedbackWidgetSettings>
    >(
      `feedback/portals/${portalId}/widget-settings`,
      input,
      await getWorkspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const createFeedbackWidgetSigningSecret = async (
  portalId: string,
  workspaceSlug: string,
) => {
  try {
    return await post<
      Record<string, never>,
      ApiResponse<FeedbackWidgetSigningSecret>
    >(
      `feedback/portals/${portalId}/widget-settings/signing-secret`,
      {},
      await getWorkspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const rotateFeedbackWidgetSigningSecret = async (
  portalId: string,
  workspaceSlug: string,
) => {
  try {
    return await post<
      Record<string, never>,
      ApiResponse<FeedbackWidgetSigningSecret>
    >(
      `feedback/portals/${portalId}/widget-settings/signing-secret/rotate`,
      {},
      await getWorkspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const createFeedbackUpdate = async (
  input: UpsertFeedbackUpdateInput,
  workspaceSlug: string,
) => {
  try {
    return await post<UpsertFeedbackUpdateInput, ApiResponse<FeedbackUpdate>>(
      "feedback/updates",
      input,
      await getWorkspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const updateFeedbackUpdate = async (
  updateId: string,
  input: UpsertFeedbackUpdateInput,
  workspaceSlug: string,
) => {
  try {
    return await put<UpsertFeedbackUpdateInput, ApiResponse<FeedbackUpdate>>(
      `feedback/updates/${updateId}`,
      input,
      await getWorkspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const deleteFeedbackUpdate = async (
  updateId: string,
  workspaceSlug: string,
) => {
  try {
    return await remove<ApiResponse<null>>(
      `feedback/updates/${updateId}`,
      await getWorkspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

const transitionFeedbackUpdate = async (
  updateId: string,
  transition: "publish" | "unpublish",
  workspaceSlug: string,
) => {
  try {
    return await post<Record<string, never>, ApiResponse<FeedbackUpdate>>(
      `feedback/updates/${updateId}/${transition}`,
      {},
      await getWorkspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const publishFeedbackUpdate = async (
  updateId: string,
  workspaceSlug: string,
) => transitionFeedbackUpdate(updateId, "publish", workspaceSlug);

export const unpublishFeedbackUpdate = async (
  updateId: string,
  workspaceSlug: string,
) => transitionFeedbackUpdate(updateId, "unpublish", workspaceSlug);
