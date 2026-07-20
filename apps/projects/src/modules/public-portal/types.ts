export type PublicPortalTab = "feedback" | "roadmap" | "updates";

export type PublicRequestStatus =
  | "pending"
  | "reviewing"
  | "planned"
  | "in_progress"
  | "completed"
  | "closed";

export type PublicPortalSort = "top" | "newest" | "oldest";

export type PublicPortalFilters = {
  boardId?: string;
  search: string;
  sort: PublicPortalSort;
  status?: PublicRequestStatus;
};

export type PublicRequestBoard = {
  id: string;
  teamId?: string;
  name: string;
  slug?: string;
  color?: string;
  colorClassName: string;
};

export type PublicRequestComment = {
  id: string;
  authorName: string;
  authorAvatar?: string | null;
  body: string;
  createdAtLabel: string;
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

export type PublicFeedbackStoryLink = {
  id: string;
  storyId: string;
  relationship: "created_from" | "linked" | "solves";
};

export type PublicRequest = {
  id: string;
  authorId: string;
  slug: string;
  title: string;
  description: string;
  authorName: string;
  authorAvatar?: string | null;
  boardId: string;
  status: PublicRequestStatus;
  voteCount: number;
  commentCount: number;
  createdAtLabel: string;
  roadmapSummary?: string;
  comments: PublicRequestComment[];
  storyLinks: PublicFeedbackStoryLink[];
};

export type PublicPortalUpdate = {
  id: string;
  title: string;
  body: string;
  status: "published" | "draft";
  publishedAtLabel: string;
  relatedRequestIds: string[];
};

export type PublicPortalWorkspace = {
  name: string;
  slug: string;
  avatarUrl: string | null;
  color: string;
};

export type PublicPortalViewer = {
  id: string;
  name: string;
  email: string;
  avatarUrl: string | null;
  appHref?: string;
  accountHref: string;
  feedbackSetupHref: string;
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

export type PublicPortal = {
  id: string;
  name: string;
  slug: string;
  workspace: PublicPortalWorkspace;
  boards: PublicRequestBoard[];
  requests: PublicRequest[];
  requestsHasMore: boolean;
  updates: PublicPortalUpdate[];
};

export type PublicFeedback = PublicRequest;
export type PublicFeedbackBoard = PublicRequestBoard;
export type PublicFeedbackComment = PublicRequestComment;
export type PublicFeedbackStatus = PublicRequestStatus;
