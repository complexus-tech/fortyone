import type { ApiResponse } from "@/types";
import type {
  FeedbackItem,
  FeedbackPage,
  FeedbackSummary,
  IntakeItem,
  IntakePage,
} from "./types";
import { get } from "@/lib/http";

async function getRequiredData<T>(
  path: string,
  signal?: AbortSignal,
): Promise<T> {
  const response = await get<ApiResponse<T>>(path, { signal });
  if (response.error?.message) throw new Error(response.error.message);
  if (response.data == null)
    throw new Error("The server returned no data. Please try again.");
  return response.data;
}

export const getFeedbackSummaries = (signal?: AbortSignal) =>
  getRequiredData<FeedbackSummary[]>("feedback/team-summaries", signal);

export const getFeedbackPage = (
  teamId: string,
  page: number,
  signal?: AbortSignal,
) =>
  getRequiredData<FeedbackPage>(
    `teams/${encodeURIComponent(teamId)}/feedback?status=active&page=${page}&pageSize=25`,
    signal,
  );

export const getIntakePage = (
  teamId: string,
  page: number,
  signal?: AbortSignal,
) =>
  getRequiredData<IntakePage>(
    `teams/${encodeURIComponent(teamId)}/integration-requests?status=pending&page=${page}&pageSize=25`,
    signal,
  );

export const getFeedbackItem = (id: string, signal?: AbortSignal) =>
  getRequiredData<FeedbackItem>(
    `feedback/items/${encodeURIComponent(id)}`,
    signal,
  );

export const getIntakeItem = (id: string, signal?: AbortSignal) =>
  getRequiredData<IntakeItem>(
    `integration-requests/${encodeURIComponent(id)}`,
    signal,
  );
