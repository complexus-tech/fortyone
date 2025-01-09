import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import { commentStoryAction } from "../actions/comment-story";
import { Comment } from "@/types";

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
        comment: payload?.comment,
        parentId: payload?.parentId,
      }),
    onError: (_, variables) => {
      toast.error("Failed to comment story", {
        description: "Your comment was not saved",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(variables),
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
          storyId: storyId,
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
    onSettled: (_, __, { storyId }) => {
      queryClient.invalidateQueries({
        queryKey: storyKeys.comments(storyId!),
      });
    },
  });

  return mutation;
};
