export type TeamFeedbackStatus =
  | "pending"
  | "reviewing"
  | "planned"
  | "in_progress"
  | "completed"
  | "closed";

export type TeamFeedbackListStatus = "active" | "all" | TeamFeedbackStatus;

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
  authorId: string;
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
  relationship: "created_from" | "linked" | "solves";
  isPrimary: boolean;
  createdByUserId: string;
  createdAt: string;
};

export type TeamFeedbackItem = {
  id: string;
  workspaceId: string;
  portalId: string;
  boardId: string;
  authorId: string;
  authorName: string;
  authorAvatar?: string | null;
  title: string;
  description: string;
  slug: string;
  status: TeamFeedbackStatus;
  voteCount: number;
  commentCount: number;
  roadmapSummary?: string | null;
  createdAt: string;
  updatedAt: string;
  board: TeamFeedbackBoard;
  comments: TeamFeedbackComment[];
  storyLinks: TeamFeedbackStoryLink[];
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
};
