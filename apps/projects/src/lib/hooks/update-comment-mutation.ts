import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import type { Comment } from "@/types";
import type {
  UpdateComment} from "../actions/comments/update-comment";
import {
  updateCommentAction,
} from "../actions/comments/update-comment";

export const useUpdateCommentMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      commentId,
      payload,
      storyId,
    }: {
      commentId: string;
      payload: UpdateComment;
      storyId: string;
    }) => updateCommentAction(commentId, payload, storyId),
    onError: (_, variables) => {
      toast.error("Failed to update comment", {
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
    onMutate: (newComment) => {
      const previousComments = queryClient.getQueryData<Comment[]>(
        storyKeys.comments(newComment.storyId),
      );
      if (previousComments) {
        const updatedComments = previousComments.map((comment) =>
          comment.id === newComment.commentId
            ? { ...comment, comment: newComment.payload.content }
            : comment,
        );
        queryClient.setQueryData<Comment[]>(
          storyKeys.comments(newComment.storyId),
          updatedComments,
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
