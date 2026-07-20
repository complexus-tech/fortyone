import type {
  PublicContributor,
  PublicContributorCommentsPage,
  PublicFeedbackStoryLink,
  PublicPortal,
  PublicPortalWorkspace,
  PublicRequestBoard,
  PublicRequestComment,
  PublicRequestStatus,
} from "./types";

type ApiPortal = {
  id: string;
  name: string;
  slug: string;
  itemsHasMore?: boolean;
  boards?: ApiBoard[];
  items?: ApiFeedbackItem[];
};

type ApiBoard = {
  id: string;
  teamId?: string;
  name: string;
  slug: string;
  color: string;
};

export type ApiFeedbackItem = {
  id: string;
  boardId: string;
  authorId: string;
  authorName: string;
  authorAvatar?: string | null;
  title: string;
  description: string;
  slug: string;
  status: PublicRequestStatus;
  voteCount: number;
  commentCount: number;
  roadmapSummary?: string | null;
  createdAt: string;
  comments?: ApiFeedbackComment[];
  storyLinks?: ApiFeedbackStoryLink[];
};

type ApiFeedbackComment = {
  id: string;
  authorName: string;
  authorAvatar?: string | null;
  body: string;
  createdAt: string;
};

export type ApiContributor = {
  id: string;
  name: string;
  avatarUrl?: string | null;
  joinedAt: string;
  stats: {
    feedbackCount: number;
    commentCount: number;
    voteScore: number;
  };
};

export type ApiContributorCommentsPage = {
  comments: {
    id: string;
    body: string;
    createdAt: string;
    feedback: {
      id: string;
      title: string;
      slug: string;
    };
  }[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
  };
};

type ApiFeedbackStoryLink = {
  id: string;
  storyId: string;
  relationship: PublicFeedbackStoryLink["relationship"];
};

const dateLabel = (value: string) => {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) return "";

  const diff = Date.now() - date.getTime();
  const minute = 60 * 1000;
  const hour = 60 * minute;
  const day = 24 * hour;

  if (diff < minute) return "Just now";
  if (diff < hour) return `${Math.floor(diff / minute)} minutes ago`;
  if (diff < day) return `${Math.floor(diff / hour)} hours ago`;
  if (diff < 2 * day) return "Yesterday";

  return date.toLocaleDateString(undefined, {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
};

export const toPublicContributor = (
  contributor: ApiContributor,
): PublicContributor => ({
  id: contributor.id,
  name: contributor.name,
  avatarUrl: contributor.avatarUrl,
  joinedAt: contributor.joinedAt,
  stats: contributor.stats,
});

export const toPublicContributorCommentsPage = (
  page: ApiContributorCommentsPage,
): PublicContributorCommentsPage => ({
  comments: page.comments.map((comment) => ({
    id: comment.id,
    body: comment.body,
    createdAtLabel: dateLabel(comment.createdAt),
    feedback: comment.feedback,
  })),
  pagination: page.pagination,
});

export const toPublicRequest = (
  item: ApiFeedbackItem,
): PublicPortal["requests"][number] => {
  const comments: PublicRequestComment[] = (item.comments ?? []).map(
    (comment) => ({
      id: comment.id,
      authorName: comment.authorName,
      authorAvatar: comment.authorAvatar,
      body: comment.body,
      createdAtLabel: dateLabel(comment.createdAt),
    }),
  );

  return {
    id: item.id,
    authorId: item.authorId,
    slug: item.slug,
    title: item.title,
    description: item.description,
    authorName: item.authorName,
    authorAvatar: item.authorAvatar,
    boardId: item.boardId,
    status: item.status,
    voteCount: item.voteCount,
    commentCount: item.commentCount,
    createdAtLabel: dateLabel(item.createdAt),
    roadmapSummary: item.roadmapSummary ?? undefined,
    comments,
    storyLinks: item.storyLinks ?? [],
  };
};

export const toPublicPortal = (
  apiPortal: ApiPortal,
  workspace?: PublicPortalWorkspace,
): PublicPortal => {
  const boards: PublicRequestBoard[] = (apiPortal.boards ?? []).map(
    (board) => ({
      id: board.id,
      teamId: board.teamId,
      name: board.name,
      slug: board.slug,
      color: board.color,
    }),
  );

  return {
    id: apiPortal.id,
    name: apiPortal.name,
    slug: apiPortal.slug,
    workspace: workspace ?? {
      avatarUrl: null,
      color: "var(--primary)",
      name: apiPortal.name,
      slug: apiPortal.slug,
    },
    boards,
    requests: (apiPortal.items ?? []).map(toPublicRequest),
    requestsHasMore: apiPortal.itemsHasMore ?? false,
    updates: [],
  };
};

export type { ApiPortal };
