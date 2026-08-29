import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  TeamFeedbackItem,
  TeamFeedbackMergeCandidate,
  TeamFeedbackPrivateAuthor,
} from "../types";

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

export const getTeamFeedbackPrivateAuthor = async (
  feedbackId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<TeamFeedbackPrivateAuthor>>(
    `feedback/items/${feedbackId}/private-author`,
    ctx,
  );

  if (response.error?.message) {
    throw new Error(response.error.message);
  }
  if (!response.data) {
    throw new Error("Feedback author identity was not returned by the server");
  }
  return response.data;
};

export const getTeamFeedbackMergeCandidates = async (
  feedbackId: string,
  search: string,
  ctx: WorkspaceCtx,
) => {
  const params = new URLSearchParams({ limit: "30" });
  const normalizedSearch = search.trim();
  if (normalizedSearch) params.set("search", normalizedSearch);

  const response = await get<
    ApiResponse<{
      candidates: TeamFeedbackMergeCandidate[];
      hasMore: boolean;
    }>
  >(`feedback/items/${feedbackId}/merge-candidates?${params.toString()}`, ctx);

  if (response.error?.message) {
    throw new Error(response.error.message);
  }
  if (!response.data) {
    throw new Error("Merge candidates were not returned by the server");
  }
  return response.data;
};
