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

export const useStoryCommentState = (storyId: string) => {
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: storyKeys.commentState(workspaceSlug, storyId),
    queryFn: () => getCommentState(storyId, { workspaceSlug }),
    staleTime: 0,
    refetchInterval: 30_000,
  });
};

export const useCommentThread = (storyId: string, commentId: string | null) => {
  const { workspaceSlug } = useWorkspacePath();
  const pendingChanges = useIsMutating({
    mutationKey: commentMutationKey(workspaceSlug),
    predicate: (mutation) =>
      (mutation.state.variables as { storyId?: string } | undefined)
        ?.storyId === storyId,
  });
  return useQuery({
    queryKey: storyKeys.commentThread(workspaceSlug, storyId, commentId ?? ""),
    queryFn: () => getCommentThread(storyId, commentId!, { workspaceSlug }),
    enabled: Boolean(commentId) && !pendingChanges,
    retry: false,
    refetchInterval: pendingChanges ? false : 30_000,
  });
};

export const useCommentCollaboration = (storyId: string, commentId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = storyKeys.commentState(workspaceSlug, storyId);
  const mergeState = (
    entry: CommentState,
    field: "reactions" | "resolution",
  ) => {
    queryClient.setQueryData<StoryCommentState>(queryKey, (current) => {
      if (!current) return current;
      const previous = current.comments.find(
        (comment) => comment.commentId === entry.commentId,
      );
      const next = previous
        ? {
            ...previous,
            ...(field === "reactions"
              ? { reactions: entry.reactions }
              : { resolvedAt: entry.resolvedAt, resolvedBy: entry.resolvedBy }),
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
  const invalidate = () => queryClient.invalidateQueries({ queryKey });
  const reaction = useMutation({
    mutationFn: ({
      emoji,
      reacted,
    }: {
      emoji: CommentEmoji;
      reacted: boolean;
    }) => setCommentReaction(commentId, emoji, reacted, { workspaceSlug }),
    onSuccess: (entry) => {
      mergeState(entry, "reactions");
    },
    onError: () => {
      toast.error("Could not update reaction. Please try again.");
    },
    onSettled: invalidate,
  });
  const resolution = useMutation({
    mutationFn: (resolved: boolean) =>
      setCommentResolution(commentId, resolved, { workspaceSlug }),
    onSuccess: (entry) => {
      mergeState(entry, "resolution");
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
          Object.entries(current.threadReads ?? {}).filter(([id]) => id !== threadId),
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
