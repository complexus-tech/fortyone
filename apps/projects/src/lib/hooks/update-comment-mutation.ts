import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import {
  commentMutationKey,
  findCachedComment,
  invalidateCommentQueries,
  updateCachedComment,
  type InfiniteCommentsData,
} from "@/shared/story/comment-cache";
import type { UpdateComment } from "../actions/comments/update-comment";
import { updateCommentAction } from "../actions/comments/update-comment";

export const useUpdateCommentMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationKey: commentMutationKey(workspaceSlug),
    mutationFn: async ({
      commentId,
      payload,
    }: {
      commentId: string;
      payload: UpdateComment;
      storyId: string;
    }) => {
      const response = await updateCommentAction(
        commentId,
        payload,
        workspaceSlug,
      );
      if (response.error?.message) throw new Error(response.error.message);
      return response;
    },

    onMutate: async ({ commentId, payload, storyId }) => {
      const queryKey = storyKeys.commentsInfinite(workspaceSlug, storyId);
      await queryClient.cancelQueries({ queryKey, exact: true });
      const previous = findCachedComment(
        queryClient.getQueryData<InfiniteCommentsData>(queryKey),
        commentId,
      );
      queryClient.setQueryData<InfiniteCommentsData>(queryKey, (current) =>
        updateCachedComment(current, commentId, (comment) => ({
          ...comment,
          comment: payload.content,
        })),
      );
      return { previousContent: previous?.comment.comment };
    },
    onError: (error, variables, context) => {
      const previousContent = context?.previousContent;
      if (previousContent !== undefined) {
        queryClient.setQueryData<InfiniteCommentsData>(
          storyKeys.commentsInfinite(workspaceSlug, variables.storyId),
          (current) =>
            updateCachedComment(current, variables.commentId, (comment) =>
              comment.comment === variables.payload.content
                ? { ...comment, comment: previousContent }
                : comment,
            ),
        );
      }
      toast.error("Failed to update comment", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSettled: (_response, _error, { storyId }) =>
      invalidateCommentQueries(queryClient, workspaceSlug, storyId),
  });

  return mutation;
};
