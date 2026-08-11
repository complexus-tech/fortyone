import { get } from "api-client";
import type { ApiResponse } from "@/types";
import type { PublicRequestStatus } from "./types";

export type FeedbackProfileActivityType = "comment" | "feedback";

export type FeedbackProfileActivity = {
  id: string;
  type: FeedbackProfileActivityType;
  feedbackId: string;
  feedbackTitle: string;
  feedbackSlug: string;
  body: string;
  boardName: string;
  status: PublicRequestStatus | "";
  voteCount: number;
  commentCount: number;
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
  voteScore: number;
  portalCount: number;
};

export const getFeedbackProfileActivity = async (
  type: FeedbackProfileActivityType,
  page = 1,
  pageSize = 20,
) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
    type,
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
