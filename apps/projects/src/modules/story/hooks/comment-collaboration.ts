import {
  useIsMutating,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { storyKeys } from "@/shared/story/cache-keys";
import { commentMutationKey } from "@/shared/story/comment-cache";
import type { Comment } from "@/types/comment";
import { isCommentUnavailableError } from "../lib/comment-errors";
import {
  getCommentState,
  getCommentThread,
  markCommentsRead,
  setCommentReaction,
  setCommentResolution,
} from "../queries/comment-collaboration";
import type {
  CommentEmoji,
  CommentReadInput,
  CommentState,
  StoryCommentState,
} from "../types/comment-collaboration";
import { setCachedCommentReaction } from "./comment-reaction-cache";

const collaborationMutationKey = (workspaceSlug: string, storyId: string) =>
  [...storyKeys.commentState(workspaceSlug, storyId), "mutation"] as const;

export const useStoryCommentState = (storyId: string) => {
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const queryKey = storyKeys.commentState(workspaceSlug, storyId);
  const mutationKey = collaborationMutationKey(workspaceSlug, storyId);
  const pendingChanges = useIsMutating({
    mutationKey,
    exact: true,
  });
  return useQuery({
    queryKey,
    queryFn: () => {
      // Explicit retries bypass `enabled`; retain pending reactions there too.
      const cached = queryClient.getQueryData<StoryCommentState>(queryKey);
      if (cached && queryClient.isMutating({ mutationKey, exact: true })) {
        return cached;
      }
      return getCommentState(storyId, { workspaceSlug });
    },
    enabled: !pendingChanges,
    staleTime: 0,
    refetchInterval: pendingChanges ? false : 30_000,
  });
};

export const useCommentThread = (storyId: string, commentId: string | null) => {
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const queryKey = storyKeys.commentThread(
    workspaceSlug,
    storyId,
    commentId ?? "",
  );
  const pendingChanges = useIsMutating({
    mutationKey: commentMutationKey(workspaceSlug),
    predicate: (mutation) =>
      (mutation.state.variables as { storyId?: string } | undefined)
        ?.storyId === storyId,
  });
  return useQuery<Comment | null>({
    queryKey,
    queryFn: async () => {
      try {
        return await getCommentThread(storyId, commentId!, { workspaceSlug });
      } catch (error) {
        if (isCommentUnavailableError(error)) {
          // A later network failure must not restore content after access was
          // revoked or the comment was deleted. Only a successful read can.
          queryClient.setQueryData<Comment | null>(queryKey, null);
        }
        throw error;
      }
    },
    enabled: Boolean(commentId) && !pendingChanges,
    retry: false,
    refetchInterval: pendingChanges ? false : 30_000,
  });
};

export const useCommentCollaboration = (storyId: string, commentId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = storyKeys.commentState(workspaceSlug, storyId);
  const mutationKey = collaborationMutationKey(workspaceSlug, storyId);
  const mergeResolution = (entry: CommentState) => {
    queryClient.setQueryData<StoryCommentState>(queryKey, (current) => {
      if (!current) return current;
      const previous = current.comments.find(
        (comment) => comment.commentId === entry.commentId,
      );
      const next = previous
        ? {
            ...previous,
            resolvedAt: entry.resolvedAt,
            resolvedBy: entry.resolvedBy,
          }
        : entry;
      return {
        ...current,
        comments: previous
          ? current.comments.map((comment) =>
              comment.commentId === next.commentId ? next : comment,
            )
          : [...current.comments, next],
      };
    });
  };
  const invalidate = () => {
    // The settling mutation still counts. The final one reconciles the story
    // without replacing optimistic reactions that are still being saved.
    if (queryClient.isMutating({ mutationKey, exact: true }) > 1) return;
    return queryClient.invalidateQueries({ queryKey, exact: true });
  };
  const reaction = useMutation({
    mutationKey,
    mutationFn: ({
      emoji,
      reacted,
    }: {
      emoji: CommentEmoji;
      reacted: boolean;
    }) => setCommentReaction(commentId, emoji, reacted, { workspaceSlug }),
    onMutate: async ({ emoji, reacted }) => {
      await queryClient.cancelQueries({ queryKey, exact: true });
      const previous = queryClient
        .getQueryData<StoryCommentState>(queryKey)
        ?.comments.find((comment) => comment.commentId === commentId);
      const previousReaction = previous?.reactions.find(
        (item) => item.emoji === emoji,
      );
      const delta =
        Number(reacted) - Number(previousReaction?.reacted ?? false);
      const count = Math.max(0, (previousReaction?.count ?? 0) + delta);
      queryClient.setQueryData<StoryCommentState>(queryKey, (current) =>
        setCachedCommentReaction(
          current,
          commentId,
          emoji,
          count ? { emoji, count, reacted } : undefined,
        ),
      );
      return { previousReaction, hadComment: Boolean(previous) };
    },
    onSuccess: (entry, { emoji }) => {
      queryClient.setQueryData<StoryCommentState>(queryKey, (current) =>
        setCachedCommentReaction(
          current,
          commentId,
          emoji,
          entry.reactions.find((item) => item.emoji === emoji),
        ),
      );
    },
    onError: (_error, { emoji }, context) => {
      if (context) {
        queryClient.setQueryData<StoryCommentState>(queryKey, (current) =>
          setCachedCommentReaction(
            current,
            commentId,
            emoji,
            context.previousReaction,
            !context.hadComment,
          ),
        );
      }
      toast.error("Could not update reaction. Please try again.");
    },
    onSettled: invalidate,
  });
  const resolution = useMutation({
    mutationKey,
    mutationFn: (resolved: boolean) =>
      setCommentResolution(commentId, resolved, { workspaceSlug }),
    onSuccess: (entry) => {
      mergeResolution(entry);
    },
    onError: () => {
      toast.error("Could not update thread. Please try again.");
    },
    onSettled: invalidate,
  });
  return { reaction, resolution };
};

export const useMarkCommentsRead = (storyId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = storyKeys.commentState(workspaceSlug, storyId);
  return useMutation({
    mutationFn: (input: CommentReadInput) =>
      markCommentsRead(storyId, input, { workspaceSlug }),
    onMutate: async ({ observedAt, threadId }) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<StoryCommentState>(queryKey);
      const previousReadAt = threadId
        ? previous?.threadReads?.[threadId]
        : previous?.lastReadAt;
      queryClient.setQueryData<StoryCommentState>(queryKey, (current) => {
        if (!current) return current;
        if (
          previousReadAt &&
          Date.parse(previousReadAt) >= Date.parse(observedAt)
        )
          return current;
        return threadId
          ? {
              ...current,
              threadReads: { ...current.threadReads, [threadId]: observedAt },
            }
          : { ...current, lastReadAt: observedAt };
      });
      return { previousReadAt };
    },
    onSuccess: (result) => {
      queryClient.setQueryData<StoryCommentState>(queryKey, (current) => {
        if (!current) return result;
        const threadReads = { ...current.threadReads };
        for (const [threadId, readAt] of Object.entries(
          result.threadReads ?? {},
        )) {
          threadReads[threadId] = laterReadAt(threadReads[threadId], readAt)!;
        }
        return {
          ...current,
          lastReadAt: laterReadAt(current.lastReadAt, result.lastReadAt),
          threadReads,
          ...(Date.parse(result.observedAt) >= Date.parse(current.observedAt)
            ? { observedAt: result.observedAt, unreadCount: result.unreadCount }
            : {}),
        };
      });
    },
    onError: (_error, { observedAt, threadId }, context) => {
      queryClient.setQueryData<StoryCommentState>(queryKey, (current) => {
        if (!current) return current;
        const readAt = threadId
          ? current.threadReads?.[threadId]
          : current.lastReadAt;
        if (readAt !== observedAt) return current;
        if (!threadId)
          return { ...current, lastReadAt: context?.previousReadAt ?? null };
        const threadReads = Object.fromEntries(
          Object.entries(current.threadReads ?? {}).filter(
            ([id]) => id !== threadId,
          ),
        );
        if (context?.previousReadAt)
          threadReads[threadId] = context.previousReadAt;
        return { ...current, threadReads };
      });
    },
  });
};

const laterReadAt = (
  current: string | null | undefined,
  incoming: string | null | undefined,
) => {
  if (!current) return incoming ?? null;
  return incoming && Date.parse(incoming) > Date.parse(current)
    ? incoming
    : current;
};
