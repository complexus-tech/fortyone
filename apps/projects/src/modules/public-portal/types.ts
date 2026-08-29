export type PublicPortalTab = "feedback" | "roadmap" | "updates";

export type FeedbackParticipationMode =
  | "account_required"
  | "verified_guest"
  | "anonymous_allowed";

export type FeedbackGuestIdentityPolicy =
  | "show_identity"
  | "allow_public_masking"
  | "always_mask_guests";

export type PublicParticipantKind =
  | "account"
  | "verified_guest"
  | "external"
  | "anonymous";

export type PublicRequestStatus =
  | "pending"
  | "reviewing"
  | "planned"
  | "in_progress"
  | "completed"
  | "closed";

export type PublicFeedbackListStatus = "active" | PublicRequestStatus;

export type PublicPortalSort = "top" | "newest" | "oldest";

export type PublicPortalFilters = {
  boardId?: string;
  search: string;
  sort: PublicPortalSort;
  status?: PublicFeedbackListStatus;
};

export type PublicRequestBoard = {
  id: string;
  teamId?: string;
  name: string;
  slug?: string;
  color?: string;
};

export type PublicRequestComment = {
  id: string;
  parentId?: string | null;
  participantKind?: PublicParticipantKind;
  authorMasked?: boolean;
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
  authorId: string | null;
  slug: string;
  title: string;
  description: string;
  authorMasked?: boolean;
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
  participantKind?: PublicParticipantKind;
  following?: boolean;
  viewerVote?: -1 | 0 | 1;
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

export type PublicPortalUpdate = {
  id: string;
  slug: string;
  title: string;
  summary?: string | null;
  body: string;
  coverImageUrl?: string | null;
  publishedAt: string;
  publishedAtLabel: string;
  linkedItems: {
    id: string;
    slug: string;
    title: string;
    status: PublicRequestStatus;
  }[];
};

export type PublicPortalWorkspace = {
  name: string;
  slug: string;
  avatarUrl: string | null;
  color: string;
};

export type PublicPortalViewer = {
  kind: "account";
  id: string;
  name: string;
  email: string;
  avatarUrl: string | null;
  appHref?: string;
  accountHref: string;
  feedbackSetupHref: string;
  canReceiveUpdates: true;
};

export type PublicPortalGuestParticipant = {
  kind: "verified_guest" | "external";
  id: string;
  name: string;
  displayName: string;
  email?: string;
  avatarUrl: string | null;
  masked: boolean;
  canReceiveUpdates: true;
  sessionExpiresAt: string;
  unreadUpdateCount: number;
};

export type PublicPortalAnonymousParticipant = {
  kind: "anonymous";
  canReceiveUpdates: false;
};

export type PublicPortalParticipant =
  | PublicPortalViewer
  | PublicPortalGuestParticipant
  | PublicPortalAnonymousParticipant;

export type LegacyPublicPortalViewer = Omit<
  PublicPortalViewer,
  "kind" | "canReceiveUpdates"
>;

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
  participationMode: FeedbackParticipationMode;
  guestIdentityPolicy: FeedbackGuestIdentityPolicy;
  hasPublishedUpdates: boolean;
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
