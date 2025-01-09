import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { Comment } from "@/types";
import { storyKeys } from "@/modules/stories/constants";
import { deleteCommentAction } from "../actions/comments/delete-comment";

export const useDeleteCommentMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      commentId,
      storyId,
    }: {
      commentId: string;
      storyId: string;
    }) => deleteCommentAction(commentId, storyId),
    onError: (_, variables) => {
      toast.error("Failed to delete comment", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => { mutation.mutate(variables); },
        },
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.comments(variables.storyId),
      });
    },
    onMutate: ({ commentId, storyId }) => {
      const previousComments = queryClient.getQueryData<Comment[]>(
        storyKeys.comments(storyId),
      );
      if (previousComments) {
        queryClient.setQueryData<Comment[]>(
          storyKeys.comments(storyId),
          previousComments.filter((comment) => comment.id !== commentId),
        );
      }
    },
    onSettled: (_, __, { storyId }) => {
      queryClient.invalidateQueries({
        queryKey: storyKeys.comments(storyId),
      });
    },
  });

  return mutation;
};
