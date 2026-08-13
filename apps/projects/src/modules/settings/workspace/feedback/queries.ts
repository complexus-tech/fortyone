import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  FeedbackItemCandidate,
  FeedbackParticipationMode,
  FeedbackPortal,
  FeedbackReviewer,
  FeedbackUpdate,
  FeedbackWidgetSettings,
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

export const getFeedbackWidgetSettings = async (
  portalId: string,
  ctx: WorkspaceCtx,
): Promise<FeedbackWidgetSettings> => {
  const response = await get<ApiResponse<FeedbackWidgetSettings>>(
    `feedback/portals/${portalId}/widget-settings`,
    ctx,
  );
  if (!response.data) {
    throw new Error("Feedback widget settings were not returned");
  }
  return response.data;
};

export const getFeedbackUpdates = async (
  ctx: WorkspaceCtx,
): Promise<FeedbackUpdate[]> => {
  const response = await get<
    ApiResponse<{
      hasMore: boolean;
      page: number;
      pageSize: number;
      updates: FeedbackUpdate[];
    }>
  >("feedback/updates?page=1&pageSize=100", ctx);
  return response.data?.updates ?? [];
};

export const getFeedbackUpdateCandidates = async (
  portalId: string,
  search: string,
  ctx: WorkspaceCtx,
) => {
  const params = new URLSearchParams({ limit: "30" });
  const normalizedSearch = search.trim();
  if (normalizedSearch) params.set("search", normalizedSearch);

  const response = await get<
    ApiResponse<{
      candidates: FeedbackItemCandidate[];
      hasMore: boolean;
    }>
  >(`feedback/portals/${portalId}/item-candidates?${params.toString()}`, ctx);
  if (response.error?.message) {
    throw new Error(response.error.message);
  }
  if (!response.data) {
    throw new Error("Feedback candidates were not returned");
  }
  return response.data;
};
