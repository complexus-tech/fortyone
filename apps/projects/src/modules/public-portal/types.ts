export type PublicPortalTab = "feedback" | "roadmap" | "updates";

export type PublicRequestStatus =
  | "pending"
  | "reviewing"
  | "planned"
  | "in_progress"
  | "completed"
  | "closed";

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

export type PublicFeedbackStoryLink = {
  id: string;
  storyId: string;
  relationship: "created_from" | "linked" | "solves";
};

export type PublicRequest = {
  id: string;
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
  name: string;
  email: string;
  avatarUrl: string | null;
  appHref: string;
  accountHref: string;
  notificationsHref: string;
};

export type PublicPortal = {
  id: string;
  name: string;
  slug: string;
  workspace: PublicPortalWorkspace;
  description: string;
  boards: PublicRequestBoard[];
  requests: PublicRequest[];
  requestsHasMore: boolean;
  updates: PublicPortalUpdate[];
};

export type PublicFeedback = PublicRequest;
export type PublicFeedbackBoard = PublicRequestBoard;
export type PublicFeedbackComment = PublicRequestComment;
export type PublicFeedbackStatus = PublicRequestStatus;
