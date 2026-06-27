export type PublicPortalTab = "requests" | "roadmap" | "updates";

export type PublicRequestStatus =
  | "pending"
  | "reviewing"
  | "planned"
  | "in_progress"
  | "completed"
  | "closed";

export type PublicRequestBoard = {
  id: string;
  name: string;
  colorClassName: string;
};

export type PublicRequestComment = {
  id: string;
  authorName: string;
  body: string;
  createdAtLabel: string;
};

export type PublicRequest = {
  id: string;
  slug: string;
  title: string;
  description: string;
  authorName: string;
  boardId: string;
  status: PublicRequestStatus;
  voteCount: number;
  commentCount: number;
  createdAtLabel: string;
  roadmapSummary?: string;
  comments: PublicRequestComment[];
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
  name: string;
  slug: string;
  workspace: PublicPortalWorkspace;
  description: string;
  boards: PublicRequestBoard[];
  requests: PublicRequest[];
  updates: PublicPortalUpdate[];
};
