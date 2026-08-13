export type TeamFeedbackStatus =
  | "pending"
  | "reviewing"
  | "planned"
  | "in_progress"
  | "completed"
  | "closed";

export type TeamFeedbackListStatus = "active" | "trashed" | TeamFeedbackStatus;

export type TeamFeedbackBoard = {
  id: string;
  workspaceId: string;
  portalId: string;
  teamId: string;
  name: string;
  slug: string;
  color: string;
  orderIndex: number;
  createdAt: string;
  updatedAt: string;
};

export type TeamFeedbackComment = {
  id: string;
  workspaceId: string;
  itemId: string;
  authorId: string | null;
  parentId?: string | null;
  authorName: string;
  authorAvatar?: string | null;
  body: string;
  createdAt: string;
  updatedAt: string;
};

export type TeamFeedbackStoryLink = {
  id: string;
  workspaceId: string;
  itemId: string;
  storyId: string;
  storyTitle?: string | null;
  relationship: "created_from" | "linked" | "solves";
  isPrimary: boolean;
  createdByUserId: string;
  createdAt: string;
};

export type StoryFeedbackLink = {
  id: string;
  workspaceId: string;
  itemId: string;
  storyId: string;
  teamId: string;
  feedbackTitle: string;
  relationship: "created_from" | "linked" | "solves";
  isPrimary: boolean;
  createdAt: string;
};

export type TeamFeedbackSummary = {
  teamId: string;
  enabled: boolean;
  totalCount: number;
  unreadCount: number;
};

export type TeamFeedbackPrivateAuthor = {
  contributorId: string;
  userId: string | null;
  kind: "account" | "verified_guest" | "external" | "anonymous";
  displayName: string;
  email: string | null;
  avatarUrl: string | null;
  publicMasked: boolean;
};

export type TeamFeedbackItem = {
  id: string;
  workspaceId: string;
  portalId: string;
  boardId: string;
  authorId: string | null;
  authorName: string;
  authorAvatar?: string | null;
  title: string;
  description: string;
  slug: string;
  status: TeamFeedbackStatus;
  voteCount: number;
  upvoteCount: number;
  downvoteCount: number;
  commentCount: number;
  readAt?: string | null;
  roadmapSummary?: string | null;
  deletedAt?: string | null;
  restoreUntil?: string | null;
  createdAt: string;
  updatedAt: string;
  mergedIntoItemId?: string | null;
  mergedAt?: string | null;
  mergedByUserId?: string | null;
  board: TeamFeedbackBoard;
  comments: TeamFeedbackComment[];
  storyLinks: TeamFeedbackStoryLink[];
};

export type TeamFeedbackMergeResult = {
  sourceItemId: string;
  targetItemId: string;
  portalId: string;
  mergedAt: string;
  mergedByUserId: string;
  movedFollowerCount: number;
  movedUpdateLinkCount: number;
  movedStoryLinkCount: number;
  target: TeamFeedbackItem;
};

export type TeamFeedbackMergeCandidate = {
  id: string;
  slug: string;
  title: string;
  status: TeamFeedbackStatus;
  voteCount: number;
  commentCount: number;
};

export type TeamFeedbackPage = {
  feedback: TeamFeedbackItem[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
    totalCount?: number;
  };
};

export type UpdateTeamFeedbackStatusInput = {
  status: TeamFeedbackStatus;
  roadmapSummary: string | null;
};

export type PlanTeamFeedbackInput = {
  teamId: string;
  storyId?: string;
};

export type PlanTeamFeedbackResult = {
  itemId: string;
  storyId: string;
  linkId: string;
  created: boolean;
};

export type CreateTeamFeedbackCommentInput = {
  body: string;
  parentId?: string;
};
