import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import type { Comment } from "@/types";
import { commentStoryAction } from "../actions/comment-story";

export const useCommentStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: {
        comment: string;
        parentId?: string | null;
      };
    }) =>
      commentStoryAction(storyId, {
        comment: payload.comment,
        parentId: payload.parentId,
      }),
    onError: (error, variables) => {
      toast.error("Failed to comment story", {
        description: error.message || "Your comment was not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onMutate: ({ storyId, payload }) => {
      const previousComments = queryClient.getQueryData<Comment[]>(
        storyKeys.comments(storyId),
      );
      if (previousComments) {
        const newComment = {
          id: "new comment",
          userId: "",
          comment: payload.comment,
          storyId,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          parentId: payload.parentId ?? null,
          subComments: [],
        };

        if (payload.parentId) {
          const parentComment = previousComments.find(
            (comment) => comment.id === payload.parentId,
          );
          if (parentComment) {
            parentComment.subComments.push(newComment);
          }
        } else {
          queryClient.setQueryData<Comment[]>(storyKeys.comments(storyId), [
            ...previousComments,
            newComment,
          ]);
        }
      }
    },
    onSuccess: (res, { storyId }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: storyKeys.comments(storyId),
      });
    },
  });

  return mutation;
};
