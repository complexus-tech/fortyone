"use client";

import { useEffect, useState } from "react";
import type { QueryClient, QueryKey } from "@tanstack/react-query";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type {
  PublicPortal,
  PublicPortalViewer,
  PublicRequest,
  PublicRequestComment,
} from "./types";
import type { PublicPortalFilterKey } from "./query-keys";
import { publicPortalKeys } from "./query-keys";
import type { PublicFeedbackPages } from "./client-query";
import {
  createFeedbackAction,
  createFeedbackCommentAction,
  toggleFeedbackVoteAction,
} from "./actions";

const updateRequestInLists = (
  queryClient: QueryClient,
  portalSlug: string,
  requestId: string,
  update: (request: PublicRequest) => PublicRequest,
) => {
  queryClient.setQueriesData<PublicFeedbackPages>(
    { queryKey: publicPortalKeys.feedbackLists(portalSlug) },
    (data) =>
      data
        ? {
            ...data,
            pages: data.pages.map((page) => ({
              ...page,
              requests: page.requests.map((request) =>
                request.id === requestId ? update(request) : request,
              ),
            })),
          }
        : data,
  );
};

const updateRequestInDetails = (
  queryClient: QueryClient,
  portalSlug: string,
  requestId: string,
  update: (request: PublicRequest) => PublicRequest,
) => {
  queryClient.setQueryData<PublicRequest>(
    publicPortalKeys.feedbackDetail(portalSlug, requestId),
    (request) => (request ? update(request) : request),
  );
};

const updateRequestCaches = (
  queryClient: QueryClient,
  portalSlug: string,
  requestId: string,
  update: (request: PublicRequest) => PublicRequest,
) => {
  updateRequestInLists(queryClient, portalSlug, requestId, update);
  updateRequestInDetails(queryClient, portalSlug, requestId, update);
};

const getFilterKey = (queryKey: QueryKey) => {
  const value = queryKey.at(-1);
  if (!value || typeof value !== "object") return null;

  return value as PublicPortalFilterKey;
};

const matchesFilter = (
  request: PublicRequest,
  filters: PublicPortalFilterKey,
) => {
  if (filters.boardId && request.boardId !== filters.boardId) return false;
  if (
    filters.status === "active" &&
    (request.status === "completed" || request.status === "closed")
  ) {
    return false;
  }
  if (
    filters.status &&
    filters.status !== "active" &&
    request.status !== filters.status
  ) {
    return false;
  }

  const search = filters.search.toLowerCase();
  return (
    !search ||
    `${request.title} ${request.description}`.toLowerCase().includes(search)
  );
};

const makeOptimisticRequest = ({
  boardId,
  description,
  title,
  viewer,
}: {
  boardId: string;
  description: string;
  title: string;
  viewer: PublicPortalViewer;
}): PublicRequest => {
  const id = `optimistic-${Date.now()}-${Math.random().toString(36).slice(2)}`;

  return {
    id,
    authorId: viewer.id,
    slug: id,
    title,
    description,
    authorName: viewer.name,
    authorAvatar: viewer.avatarUrl,
    boardId,
    status: "pending",
    voteCount: 0,
    commentCount: 0,
    createdAtLabel: "Just now",
    comments: [],
    storyLinks: [],
  };
};

const insertOptimisticRequest = (
  queryClient: QueryClient,
  portalSlug: string,
  request: PublicRequest,
) => {
  const entries = queryClient.getQueriesData<PublicFeedbackPages>({
    queryKey: publicPortalKeys.feedbackLists(portalSlug),
  });

  entries.forEach(([queryKey, data]) => {
    const filters = getFilterKey(queryKey);
    if (!data || !filters || !matchesFilter(request, filters)) return;

    const lastPageIndex = data.pages.length - 1;
    let inserted = false;
    queryClient.setQueryData<PublicFeedbackPages>(queryKey, {
      ...data,
      pages: data.pages.map((page, pageIndex) => {
        if (filters.sort === "newest" && pageIndex === 0) {
          return { ...page, requests: [request, ...page.requests] };
        }

        if (filters.sort === "oldest" && pageIndex === lastPageIndex) {
          return { ...page, requests: [...page.requests, request] };
        }

        if (filters.sort !== "top" || inserted) return page;

        // Top feedback is ordered by vote count, then newest first. A new
        // request has no votes, so place it before existing zero/negative
        // entries and after every positive entry currently in the cache.
        const insertionIndex = page.requests.findIndex(
          (cachedRequest) => cachedRequest.voteCount <= request.voteCount,
        );
        if (insertionIndex >= 0) {
          inserted = true;
          return {
            ...page,
            requests: [
              ...page.requests.slice(0, insertionIndex),
              request,
              ...page.requests.slice(insertionIndex),
            ],
          };
        }

        if (pageIndex === lastPageIndex) {
          inserted = true;
          return { ...page, requests: [...page.requests, request] };
        }

        return page;
      }),
    });
  });
};

const removeOptimisticRequest = (
  queryClient: QueryClient,
  portalSlug: string,
  optimisticId: string,
) => {
  queryClient.setQueriesData<PublicFeedbackPages>(
    { queryKey: publicPortalKeys.feedbackLists(portalSlug) },
    (data) =>
      data
        ? {
            ...data,
            pages: data.pages.map((page) => ({
              ...page,
              requests: page.requests.filter(
                (request) => request.id !== optimisticId,
              ),
            })),
          }
        : data,
  );
};

const replaceOptimisticRequest = (
  queryClient: QueryClient,
  portalSlug: string,
  optimisticId: string,
  request: PublicRequest,
) => {
  queryClient.setQueriesData<PublicFeedbackPages>(
    { queryKey: publicPortalKeys.feedbackLists(portalSlug) },
    (data) =>
      data
        ? {
            ...data,
            pages: data.pages.map((page) => ({
              ...page,
              requests: page.requests.map((cachedRequest) =>
                cachedRequest.id === optimisticId ? request : cachedRequest,
              ),
            })),
          }
        : data,
  );
};

const getActionData = <T>(
  response: {
    data?: T | null;
    error?: { message: string };
  },
  fallbackMessage: string,
) => {
  if (response.error?.message) throw new Error(response.error.message);
  if (!response.data) throw new Error(fallbackMessage);
  return response.data;
};

export const useCreatePublicFeedback = ({
  portal,
  viewer,
}: {
  portal: PublicPortal;
  viewer: PublicPortalViewer;
}) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: Parameters<typeof createFeedbackAction>[0]) =>
      getActionData(
        await createFeedbackAction(input),
        "Unable to submit feedback",
      ),
    onMutate: async (input) => {
      await queryClient.cancelQueries({
        queryKey: publicPortalKeys.feedback(portal.slug),
      });
      const optimisticRequest = makeOptimisticRequest({ ...input, viewer });
      insertOptimisticRequest(queryClient, portal.slug, optimisticRequest);

      return { optimisticRequest };
    },
    onError: (error, _input, context) => {
      if (context) {
        removeOptimisticRequest(
          queryClient,
          portal.slug,
          context.optimisticRequest.id,
        );
      }
      toast.error("Feedback", { description: error.message });
    },
    onSuccess: (request, _input, context) => {
      replaceOptimisticRequest(
        queryClient,
        portal.slug,
        context.optimisticRequest.id,
        request,
      );
      void queryClient.invalidateQueries({
        queryKey: publicPortalKeys.feedback(portal.slug),
      });
    },
    retry: false,
  });
};

export const usePublicFeedbackVote = ({
  portalSlug,
  request,
}: {
  portalSlug: string;
  request: PublicRequest;
}) => {
  const queryClient = useQueryClient();
  const [vote, setVote] = useState<-1 | 0 | 1>(0);
  const [voteCount, setVoteCount] = useState(request.voteCount);

  useEffect(() => {
    setVoteCount(request.voteCount);
  }, [request.voteCount]);

  const mutation = useMutation({
    mutationFn: async (direction: -1 | 1) =>
      getActionData(
        await toggleFeedbackVoteAction({
          itemId: request.id,
          itemSlug: request.slug,
          portalSlug,
          vote: direction,
        }),
        "Unable to save vote",
      ),
    onMutate: async (direction) => {
      await queryClient.cancelQueries({
        queryKey: publicPortalKeys.portal(portalSlug),
      });
      const previous = { vote, voteCount };
      const nextVote = vote === direction ? 0 : direction;
      const nextVoteCount = voteCount + nextVote - vote;

      setVote(nextVote);
      setVoteCount(nextVoteCount);
      updateRequestCaches(queryClient, portalSlug, request.id, (cached) => ({
        ...cached,
        voteCount: nextVoteCount,
      }));

      return { nextVote, nextVoteCount, previous };
    },
    onError: (error, _direction, context) => {
      if (context) {
        setVote((currentVote) =>
          currentVote === context.nextVote
            ? context.previous.vote
            : currentVote,
        );
        setVoteCount((currentVoteCount) =>
          currentVoteCount === context.nextVoteCount
            ? context.previous.voteCount
            : currentVoteCount,
        );
        updateRequestCaches(queryClient, portalSlug, request.id, (cached) =>
          cached.voteCount === context.nextVoteCount
            ? { ...cached, voteCount: context.previous.voteCount }
            : cached,
        );
      }
      toast.error("Vote", { description: error.message });
    },
    onSuccess: (result) => {
      setVote(result.vote);
      setVoteCount(result.voteCount);
      updateRequestCaches(queryClient, portalSlug, request.id, (cached) => ({
        ...cached,
        voteCount: result.voteCount,
      }));
      void Promise.all([
        queryClient.invalidateQueries({
          queryKey: publicPortalKeys.feedback(portalSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: publicPortalKeys.roadmaps(portalSlug),
        }),
      ]);
    },
    retry: false,
  });

  return { mutation, vote, voteCount };
};

const toPublicComment = (
  comment: NonNullable<
    Awaited<ReturnType<typeof createFeedbackCommentAction>>["data"]
  >,
): PublicRequestComment => ({
  id: comment.id,
  parentId: comment.parentId,
  authorName: comment.authorName,
  authorAvatar: comment.authorAvatar,
  body: comment.body,
  createdAtLabel: "Just now",
});

export const useCreatePublicFeedbackComment = ({
  portalSlug,
  request,
  viewer,
}: {
  portalSlug: string;
  request: PublicRequest;
  viewer: PublicPortalViewer;
}) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      body,
      parentId,
    }: {
      body: string;
      parentId?: string;
    }) =>
      getActionData(
        await createFeedbackCommentAction({
          body,
          itemId: request.id,
          itemSlug: request.slug,
          ...(parentId ? { parentId } : {}),
          portalSlug,
        }),
        "Unable to add comment",
      ),
    onMutate: async ({ body, parentId }) => {
      await queryClient.cancelQueries({
        queryKey: publicPortalKeys.feedback(portalSlug),
      });
      const optimisticComment: PublicRequestComment = {
        id: `optimistic-${Date.now()}-${Math.random().toString(36).slice(2)}`,
        parentId,
        authorName: viewer.name,
        authorAvatar: viewer.avatarUrl,
        body,
        createdAtLabel: "Just now",
      };

      updateRequestCaches(queryClient, portalSlug, request.id, (cached) => ({
        ...cached,
        commentCount: cached.commentCount + 1,
        comments: [optimisticComment, ...cached.comments],
      }));

      return { optimisticComment };
    },
    onError: (error, _body, context) => {
      if (context) {
        updateRequestCaches(queryClient, portalSlug, request.id, (cached) => {
          const hasOptimisticComment = cached.comments.some(
            (comment) => comment.id === context.optimisticComment.id,
          );
          if (!hasOptimisticComment) return cached;

          return {
            ...cached,
            commentCount: Math.max(0, cached.commentCount - 1),
            comments: cached.comments.filter(
              (comment) => comment.id !== context.optimisticComment.id,
            ),
          };
        });
      }
      toast.error("Comment", { description: error.message });
    },
    onSuccess: (comment, _body, context) => {
      const confirmedComment = toPublicComment(comment);
      updateRequestCaches(queryClient, portalSlug, request.id, (cached) => ({
        ...cached,
        comments: cached.comments.map((cachedComment) =>
          cachedComment.id === context.optimisticComment.id
            ? confirmedComment
            : cachedComment,
        ),
      }));
      void queryClient.invalidateQueries({
        queryKey: publicPortalKeys.feedback(portalSlug),
      });
    },
    retry: false,
  });
};
