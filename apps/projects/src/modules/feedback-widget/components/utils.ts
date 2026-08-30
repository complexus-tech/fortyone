import type {
  PublicFeedbackListStatus,
  PublicRequest,
  PublicRequestStatus,
} from "@/shared/feedback-widget/types";
import { feedbackRequestStatusMeta as requestStatusMeta } from "@/shared/feedback-widget/status";
import type { WidgetRoadmapPagination } from "./types";

export const INITIAL_ROADMAP_VISIBLE_COUNT = 3;

export const createInitialRoadmapPagination = (): WidgetRoadmapPagination => ({
  completed: { hasMore: false, nextPage: 2 },
  in_progress: { hasMore: false, nextPage: 2 },
  planned: { hasMore: false, nextPage: 2 },
});

export const roadmapSections = [
  { label: "In progress", status: "in_progress" },
  { label: "Planned", status: "planned" },
  { label: "Completed", status: "completed" },
] as const;

export const replaceRequest = (
  requests: PublicRequest[],
  updatedRequest: PublicRequest,
) =>
  requests.map((request) =>
    request.id === updatedRequest.id ? updatedRequest : request,
  );

export const mergeRequests = (
  current: PublicRequest[],
  incoming: PublicRequest[],
) => {
  const requests = new Map(current.map((request) => [request.id, request]));
  incoming.forEach((request) => {
    requests.set(request.id, request);
  });
  return Array.from(requests.values());
};

export const statusAccent = (status: PublicRequestStatus) => {
  if (status === "in_progress") return "bg-info";
  return requestStatusMeta[status].dotClassName;
};

export const getFeedbackEmptyBody = (
  search: string,
  status: PublicFeedbackListStatus,
) => {
  if (search) return "Try a different phrase or share the idea yourself.";
  if (status !== "active") return "There is no feedback with this status yet.";
  return "New suggestions will appear here as soon as they are shared.";
};
