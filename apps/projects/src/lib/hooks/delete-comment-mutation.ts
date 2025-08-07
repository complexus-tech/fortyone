import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { Comment } from "@/types";
import { storyKeys } from "@/modules/stories/constants";
import { deleteCommentAction } from "../actions/comments/delete-comment";

export const useDeleteCommentMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ commentId }: { commentId: string; storyId: string }) =>
      deleteCommentAction(commentId),

    onMutate: ({ commentId, storyId }) => {
      const previousComments = queryClient.getQueryData<Comment[]>(
        storyKeys.commentsInfinite(storyId),
      );
      if (previousComments) {
        queryClient.setQueryData<Comment[]>(
          storyKeys.commentsInfinite(storyId),
          previousComments.filter((comment) => comment.id !== commentId),
        );
      }
      return { previousComments };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(
        storyKeys.commentsInfinite(variables.storyId),
        context?.previousComments,
      );
      toast.error("Failed to delete comment", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.commentsInfinite(variables.storyId),
      });
    },
    onSuccess: (res, { storyId }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: storyKeys.commentsInfinite(storyId),
      });
    },
  });

  return mutation;
};
