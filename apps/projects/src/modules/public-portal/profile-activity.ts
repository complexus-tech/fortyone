import { get } from "api-client";
import type { ApiResponse } from "@/types";

export type FeedbackProfileActivity = {
  id: string;
  type: "comment" | "feedback";
  feedbackId: string;
  feedbackTitle: string;
  feedbackSlug: string;
  body: string;
  portalSlug: string;
  workspaceName: string;
  workspaceSlug: string;
  createdAt: string;
};

export type FeedbackProfileActivityPage = {
  activities: FeedbackProfileActivity[];
  page: number;
  pageSize: number;
  hasMore: boolean;
  feedbackCount: number;
  commentCount: number;
};

export const getFeedbackProfileActivity = async (page = 1, pageSize = 20) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  const response = await get<ApiResponse<FeedbackProfileActivityPage>>(
    `feedback/contributor/activity?${params.toString()}`,
    { cache: "no-store" },
  );

  if (!response.data) {
    throw new Error("Unable to load feedback activity");
  }
  return response.data;
};
