import type { PublicPortalFilters, PublicRequestStatus } from "./types";

export type PublicPortalFilterKey = {
  boardId: string | null;
  search: string;
  sort: PublicPortalFilters["sort"];
  status: PublicRequestStatus | null;
};

export const toPublicPortalFilterKey = (
  filters: PublicPortalFilters,
): PublicPortalFilterKey => ({
  boardId: filters.boardId ?? null,
  search: filters.search.trim(),
  sort: filters.sort,
  status: filters.status ?? null,
});

export const publicPortalKeys = {
  all: ["public-portal"] as const,
  portal: (portalSlug: string) =>
    [...publicPortalKeys.all, portalSlug] as const,
  feedback: (portalSlug: string) =>
    [...publicPortalKeys.portal(portalSlug), "feedback"] as const,
  feedbackLists: (portalSlug: string) =>
    [...publicPortalKeys.feedback(portalSlug), "list"] as const,
  feedbackList: (portalSlug: string, filters: PublicPortalFilters) =>
    [
      ...publicPortalKeys.feedbackLists(portalSlug),
      toPublicPortalFilterKey(filters),
    ] as const,
  feedbackDetails: (portalSlug: string) =>
    [...publicPortalKeys.feedback(portalSlug), "detail"] as const,
  feedbackDetail: (portalSlug: string, requestId: string) =>
    [...publicPortalKeys.feedbackDetails(portalSlug), requestId] as const,
  authorFeedback: (portalSlug: string, authorId: string) =>
    [...publicPortalKeys.feedback(portalSlug), "author", authorId] as const,
  authorProfile: (portalSlug: string, authorId: string) =>
    [
      ...publicPortalKeys.authorFeedback(portalSlug, authorId),
      "profile",
    ] as const,
  authorComments: (portalSlug: string, authorId: string) =>
    [
      ...publicPortalKeys.authorFeedback(portalSlug, authorId),
      "comments",
    ] as const,
  notifications: (portalSlug: string) =>
    [...publicPortalKeys.portal(portalSlug), "notifications"] as const,
  notificationLists: (portalSlug: string) =>
    [...publicPortalKeys.notifications(portalSlug), "list"] as const,
  notificationList: (portalSlug: string, unreadOnly = false) =>
    [
      ...publicPortalKeys.notificationLists(portalSlug),
      { unreadOnly },
    ] as const,
  notificationUnreadCount: (portalSlug: string) =>
    [...publicPortalKeys.notifications(portalSlug), "unread"] as const,
  roadmaps: (portalSlug: string) =>
    [...publicPortalKeys.portal(portalSlug), "roadmap"] as const,
  roadmap: (portalSlug: string, status: PublicRequestStatus) =>
    [...publicPortalKeys.roadmaps(portalSlug), status] as const,
};
