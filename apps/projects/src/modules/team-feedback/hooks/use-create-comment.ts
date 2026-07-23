import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { feedbackKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import { createTeamFeedbackCommentAction } from "../actions/comment";
import type {
  CreateTeamFeedbackCommentInput,
  TeamFeedbackComment,
  TeamFeedbackItem,
} from "../types";

export const useCreateTeamFeedbackComment = (feedbackId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const detailKey = feedbackKeys.detail(workspaceSlug, feedbackId);
  const mutationKey = [
    ...feedbackKeys.all(workspaceSlug),
    "create-comment",
    feedbackId,
  ] as const;

  return useMutation({
    mutationKey,
    mutationFn: async (input: CreateTeamFeedbackCommentInput) => {
      const response = await createTeamFeedbackCommentAction(
        feedbackId,
        input,
        workspaceSlug,
      );
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      if (!response.data) {
        throw new Error("The comment was not returned by the server");
      }
      return response.data;
    },
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: detailKey });
      const optimisticComment: TeamFeedbackComment = {
        id: `optimistic-${Date.now()}-${Math.random().toString(36).slice(2)}`,
        workspaceId: "",
        itemId: feedbackId,
        authorId: session?.user.id ?? "",
        parentId: input.parentId,
        authorName: session?.user.name ?? "You",
        authorAvatar: session?.user.image,
        body: input.body,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      queryClient.setQueryData<TeamFeedbackItem>(detailKey, (feedback) =>
        feedback
          ? {
              ...feedback,
              commentCount: feedback.commentCount + 1,
              comments: [optimisticComment, ...feedback.comments],
            }
          : feedback,
      );

      return { optimisticComment };
    },
    onError: (error, _input, context) => {
      if (context) {
        queryClient.setQueryData<TeamFeedbackItem>(detailKey, (feedback) => {
          if (!feedback) return feedback;
          const hasOptimisticComment = feedback.comments.some(
            (comment) => comment.id === context.optimisticComment.id,
          );
          if (!hasOptimisticComment) return feedback;

          return {
            ...feedback,
            commentCount: Math.max(0, feedback.commentCount - 1),
            comments: feedback.comments.filter(
              (comment) => comment.id !== context.optimisticComment.id,
            ),
          };
        });
      }
      toast.error("Failed to add comment", {
        description: error.message,
      });
    },
    onSuccess: (comment, _input, context) => {
      queryClient.setQueryData<TeamFeedbackItem>(detailKey, (feedback) =>
        feedback
          ? {
              ...feedback,
              comments: feedback.comments.map((cachedComment) =>
                cachedComment.id === context.optimisticComment.id
                  ? comment
                  : cachedComment,
              ),
            }
          : feedback,
      );
    },
    onSettled: () => {
      if (queryClient.isMutating({ mutationKey }) === 1) {
        void Promise.all([
          queryClient.invalidateQueries({ queryKey: detailKey }),
          queryClient.invalidateQueries({
            queryKey: feedbackKeys.lists(workspaceSlug),
          }),
        ]);
      }
    },
  });
};
