import type { QueryClient } from "@tanstack/react-query";
import type { Comment } from "@/types/comment";
import { storyKeys } from "./cache-keys";

export type InfiniteCommentsData = {
  pages: {
    comments: Comment[];
    pagination: {
      page: number;
      pageSize: number;
      hasMore: boolean;
      nextPage: number;
    };
  }[];
  pageParams: number[];
};

export const commentMutationKey = (workspaceSlug: string) =>
  ["story-comments", workspaceSlug, "mutation"] as const;

const mapCommentTree = (
  comments: Comment[],
  transform: (comment: Comment) => Comment | null,
): Comment[] => {
  let changed = false;
  const result: Comment[] = [];

  for (const comment of comments) {
    const transformed = transform(comment);
    if (!transformed) {
      changed = true;
      continue;
    }
    const subComments = mapCommentTree(transformed.subComments, transform);
    const next =
      subComments === transformed.subComments
        ? transformed
        : { ...transformed, subComments };
    changed ||= next !== comment;
    result.push(next);
  }

  return changed ? result : comments;
};

const transformCachedComments = (
  data: InfiniteCommentsData | undefined,
  transform: (comment: Comment) => Comment | null,
): InfiniteCommentsData | undefined => {
  if (!data) return data;
  const pages = data.pages.map((page) => {
    const comments = mapCommentTree(page.comments, transform);
    if (comments === page.comments) return page;
    return { ...page, comments };
  });
  return pages.some((page, index) => page !== data.pages[index])
    ? { ...data, pages }
    : data;
};

export const updateCachedComment = (
  data: InfiniteCommentsData | undefined,
  commentId: string,
  update: (comment: Comment) => Comment | null,
) =>
  transformCachedComments(data, (comment) =>
    comment.id === commentId ? update(comment) : comment,
  );

const findInTree = (
  comments: Comment[],
  commentId: string,
): Comment | undefined => {
  for (const comment of comments) {
    if (comment.id === commentId) return comment;
    const reply = findInTree(comment.subComments, commentId);
    if (reply) return reply;
  }
};

export const findCachedComment = (
  data: InfiniteCommentsData | undefined,
  commentId: string,
) => {
  if (!data) return;
  for (const [pageIndex, page] of data.pages.entries()) {
    const comment = findInTree(page.comments, commentId);
    if (comment) return { comment, pageIndex };
  }
};

const compareComments = (left: Comment, right: Comment) =>
  Date.parse(left.createdAt) - Date.parse(right.createdAt) ||
  left.id.localeCompare(right.id);

const insertAtLocation = (
  data: InfiniteCommentsData | undefined,
  comment: Comment,
  parentId: string | null,
  pageIndex: number,
): InfiniteCommentsData | undefined => {
  if (!data?.pages.length || findCachedComment(data, comment.id)) return data;
  if (parentId) {
    return updateCachedComment(data, parentId, (parent) => ({
      ...parent,
      subComments: [...parent.subComments, comment].sort(compareComments),
    }));
  }
  const targetPage = Math.min(pageIndex, data.pages.length - 1);
  return {
    ...data,
    pages: data.pages.map((page, index) =>
      index === targetPage
        ? {
            ...page,
            comments: [...page.comments, comment].sort((left, right) =>
              compareComments(right, left),
            ),
          }
        : page,
    ),
  };
};

export const insertCachedComment = (
  data: InfiniteCommentsData | undefined,
  comment: Comment,
) => insertAtLocation(data, comment, comment.parentId, 0);

type CachedCommentLocation = {
  comment: Comment;
  pageIndex: number;
  containerId: string | null;
};

export const getCachedCommentBranch = (
  data: InfiniteCommentsData | undefined,
  commentId: string,
): CachedCommentLocation[] => {
  const locations: CachedCommentLocation[] = [];
  const children = new Map<string, string[]>();
  const visit = (
    comments: Comment[],
    pageIndex: number,
    containerId: string | null,
  ) => {
    for (const comment of comments) {
      locations.push({ comment, pageIndex, containerId });
      const parentId = comment.parentId ?? containerId;
      if (parentId) {
        const siblings = children.get(parentId) ?? [];
        siblings.push(comment.id);
        children.set(parentId, siblings);
      }
      visit(comment.subComments, pageIndex, comment.id);
    }
  };
  data?.pages.forEach((page, index) => {
    visit(page.comments, index, null);
  });

  // The API can flatten descendants under the root while preserving parentId.
  // Follow those relationships so deletion matches the server's cascade.
  const removedIds = new Set([commentId]);
  for (const id of removedIds) {
    for (const childId of children.get(id) ?? []) removedIds.add(childId);
  }
  return locations.filter(({ comment }) => removedIds.has(comment.id));
};

export const removeCachedComment = (
  data: InfiniteCommentsData | undefined,
  commentId: string,
) => {
  const removedIds = new Set(
    getCachedCommentBranch(data, commentId).map(({ comment }) => comment.id),
  );
  return transformCachedComments(data, (comment) =>
    removedIds.has(comment.id) ? null : comment,
  );
};

export const restoreCachedComments = (
  data: InfiniteCommentsData | undefined,
  removed: CachedCommentLocation[],
) =>
  removed.reduce(
    (current, { comment, pageIndex, containerId }) =>
      insertAtLocation(current, comment, containerId, pageIndex),
    data,
  );

export const invalidateCommentQueries = (
  queryClient: QueryClient,
  workspaceSlug: string,
  storyId: string,
) => {
  const pending = queryClient.isMutating({
    mutationKey: commentMutationKey(workspaceSlug),
    predicate: (mutation) =>
      (mutation.state.variables as { storyId?: string } | undefined)
        ?.storyId === storyId,
  });
  // The settling mutation is still pending here. A later mutation will refetch
  // after its own optimistic change settles, without replacing work in flight.
  if (pending > 1) return;
  return Promise.all([
    queryClient.invalidateQueries({
      queryKey: storyKeys.commentsInfinite(workspaceSlug, storyId),
      exact: true,
    }),
    queryClient.invalidateQueries({
      queryKey: storyKeys.commentThreads(workspaceSlug, storyId),
    }),
    queryClient.invalidateQueries({
      queryKey: storyKeys.commentState(workspaceSlug, storyId),
      exact: true,
    }),
  ]);
};
