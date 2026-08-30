import type * as FeedbackWidget from "@/shared/feedback-widget/types";

export type PublicPortalTab = "feedback" | "roadmap" | "updates";

/**
 * Portal aliases preserve the public-portal API while the hosted portal and
 * embedded widget consume the same feedback contract.
 */
export type FeedbackParticipationMode =
  FeedbackWidget.FeedbackParticipationMode;
export type FeedbackGuestIdentityPolicy =
  FeedbackWidget.FeedbackGuestIdentityPolicy;
export type PublicParticipantKind = FeedbackWidget.PublicParticipantKind;
export type PublicRequestStatus = FeedbackWidget.PublicRequestStatus;
export type PublicFeedbackListStatus = FeedbackWidget.PublicFeedbackListStatus;
export type PublicPortalSort = FeedbackWidget.PublicPortalSort;
export type PublicRequestBoard = FeedbackWidget.PublicRequestBoard;
export type PublicRequestComment = FeedbackWidget.PublicRequestComment;
export type PublicFeedbackStoryLink = FeedbackWidget.PublicFeedbackStoryLink;
export type PublicRequest = FeedbackWidget.PublicRequest;
export type PublicPortalUpdate = FeedbackWidget.PublicPortalUpdate;
export type PublicPortalWorkspace = FeedbackWidget.PublicPortalWorkspace;
export type PublicPortalViewer = FeedbackWidget.PublicPortalViewer;
export type PublicPortalGuestParticipant =
  FeedbackWidget.PublicPortalGuestParticipant;
export type PublicPortalAnonymousParticipant =
  FeedbackWidget.PublicPortalAnonymousParticipant;
export type PublicPortalParticipant = FeedbackWidget.PublicPortalParticipant;
export type LegacyPublicPortalViewer = FeedbackWidget.LegacyPublicPortalViewer;
export type PublicPortal = FeedbackWidget.PublicPortal;

export type PublicPortalFilters = {
  boardId?: string;
  search: string;
  sort: PublicPortalSort;
  status?: PublicFeedbackListStatus;
};

export type SimilarPublicFeedback = {
  id: string;
  slug: string;
  title: string;
  authorId: string | null;
  authorName: string;
  authorAvatar?: string | null;
  status?: PublicRequestStatus;
  voteCount: number;
  commentCount: number;
  confidence: number;
  isDuplicate: boolean;
};

export type PublicContributorStats = {
  feedbackCount: number;
  commentCount: number;
  voteScore: number;
};

export type PublicContributor = {
  id: string;
  name: string;
  avatarUrl?: string | null;
  joinedAt: string;
  stats: PublicContributorStats;
};

export type PublicContributorComment = {
  id: string;
  body: string;
  createdAtLabel: string;
  feedback: {
    id: string;
    title: string;
    slug: string;
  };
};

export type PublicContributorCommentsPage = {
  comments: PublicContributorComment[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
  };
};

export type PublicPortalNotification = {
  id: string;
  type: "feedback_comment" | "feedback_status_update";
  title: string;
  message: {
    template: string;
    variables: Partial<
      Record<
        string,
        {
          type: string;
          value: string;
        }
      >
    >;
  };
  actor: {
    id: string;
    name: string;
    avatarUrl: string | null;
  };
  feedback: {
    id: string;
    title: string;
    slug: string;
    path: string;
  };
  createdAt: string;
  readAt: string | null;
};

export type PublicPortalNotificationsPage = {
  notifications: PublicPortalNotification[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
  };
};

export type PublicFeedback = PublicRequest;
export type PublicFeedbackBoard = PublicRequestBoard;
export type PublicFeedbackComment = PublicRequestComment;
export type PublicFeedbackStatus = PublicRequestStatus;
