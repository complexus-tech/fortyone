import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import {
  commentMutationKey,
  getCachedCommentBranch,
  invalidateCommentQueries,
  removeCachedComment,
  restoreCachedComments,
  type InfiniteCommentsData,
} from "@/shared/story/comment-cache";
import { deleteCommentAction } from "../actions/comments/delete-comment";

export const useDeleteCommentMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationKey: commentMutationKey(workspaceSlug),
    mutationFn: async ({
      commentId,
    }: {
      commentId: string;
      storyId: string;
    }) => {
      const response = await deleteCommentAction(commentId, workspaceSlug);
      if (response.error?.message) throw new Error(response.error.message);
      return response;
    },

    onMutate: async ({ commentId, storyId }) => {
      const queryKey = storyKeys.commentsInfinite(workspaceSlug, storyId);
      await queryClient.cancelQueries({ queryKey, exact: true });
      const removed = getCachedCommentBranch(
        queryClient.getQueryData<InfiniteCommentsData>(queryKey),
        commentId,
      );
      queryClient.setQueryData<InfiniteCommentsData>(queryKey, (current) =>
        removeCachedComment(current, commentId),
      );
      return { removed };
    },
    onError: (error, variables, context) => {
      const removed = context?.removed;
      if (removed) {
        queryClient.setQueryData<InfiniteCommentsData>(
          storyKeys.commentsInfinite(workspaceSlug, variables.storyId),
          (current) => restoreCachedComments(current, removed),
        );
      }
      toast.error("Failed to delete comment", {
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
