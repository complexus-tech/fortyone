import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { Comment } from "@/types";
import { storyKeys } from "@/modules/stories/constants";
import { deleteCommentAction } from "../actions/comments/delete-comment";

type InfiniteCommentsData = {
  pages: Array<{
    comments: Comment[];
    pagination: {
      page: number;
      pageSize: number;
      hasMore: boolean;
      nextPage: number;
    };
  }>;
  pageParams: number[];
};

export const useDeleteCommentMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ commentId }: { commentId: string; storyId: string }) =>
      deleteCommentAction(commentId),

    onMutate: ({ commentId, storyId }) => {
      const previousData = queryClient.getQueryData<InfiniteCommentsData>(
        storyKeys.commentsInfinite(storyId),
      );

      if (previousData) {
        const updatedData = {
          ...previousData,
          pages: previousData.pages.map((page) => ({
            ...page,
            comments: page.comments.filter(
              (comment) => comment.id !== commentId,
            ),
          })),
        };

        queryClient.setQueryData<InfiniteCommentsData>(
          storyKeys.commentsInfinite(storyId),
          updatedData,
        );
      }
      return { previousData };
    },
    onError: (error, variables, context) => {
      if (context?.previousData) {
        queryClient.setQueryData(
          storyKeys.commentsInfinite(variables.storyId),
          context.previousData,
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
