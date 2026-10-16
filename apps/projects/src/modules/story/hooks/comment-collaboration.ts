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
  return useMutation({
    mutationFn: (observedAt: string) =>
      markCommentsRead(storyId, observedAt, { workspaceSlug }),
    onSuccess: (result) => {
      queryClient.setQueryData<StoryCommentState>(
        storyKeys.commentState(workspaceSlug, storyId),
        (current) => {
          if (!current) return result;
          // An older tab or request must not move the read marker backwards.
          if (
            current.lastReadAt &&
            result.lastReadAt &&
            Date.parse(current.lastReadAt) > Date.parse(result.lastReadAt)
          )
            return current;
          return {
            ...current,
            lastReadAt: result.lastReadAt,
            unreadCount: result.unreadCount,
          };
        },
      );
    },
  });
};
