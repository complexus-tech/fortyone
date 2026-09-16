export type TeamPagination = {
  page: number;
  nextPage: number;
  hasMore: boolean;
  totalCount?: number;
};

export type FeedbackSummary = {
  teamId: string;
  enabled: boolean;
  totalCount: number;
  unreadCount: number;
};

export type FeedbackItem = {
  id: string;
  title: string;
  description: string;
  status:
    | "pending"
    | "reviewing"
    | "planned"
    | "in_progress"
    | "completed"
    | "closed";
  authorName: string;
  authorAvatar?: string | null;
  voteCount: number;
  commentCount: number;
  board?: { teamId: string; name: string } | null;
  comments:
    | {
        id: string;
        authorName: string;
        body: string;
        createdAt: string;
      }[]
    | null;
  storyLinks:
    | { id: string; storyId: string; storyTitle?: string | null }[]
    | null;
};

export type IntakeItem = {
  id: string;
  teamId: string;
  title: string;
  description?: string | null;
  provider: "github" | "slack" | "intercom";
  status: "pending" | "accepted" | "declined";
  priority: string;
  acceptedStoryId?: string | null;
  createdAt: string;
};

export type FeedbackPage = {
  feedback: FeedbackItem[];
  pagination: TeamPagination;
};
export type IntakePage = { requests: IntakeItem[]; pagination: TeamPagination };

export const feedbackStatusLabels: Record<FeedbackItem["status"], string> = {
  pending: "Pending",
  reviewing: "Reviewing",
  planned: "Planned",
  in_progress: "In progress",
  completed: "Completed",
  closed: "Closed",
};
