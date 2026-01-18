import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { Comment } from "@/types";
import type { UpdateComment } from "../actions/comments/update-comment";
import { updateCommentAction } from "../actions/comments/update-comment";

type InfiniteCommentsData = {
  pages: {
    comments: Comment[];
    pagination: {
      page: number;
      pageSize: number;
      hasMore: boolean;
      nextPage: number;
    };
  }[];
  pageParams: number[];
};

export const useUpdateCommentMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: ({
      commentId,
      payload,
    }: {
      commentId: string;
      payload: UpdateComment;
      storyId: string;
    }) => updateCommentAction(commentId, payload, workspaceSlug),

    onMutate: (newComment) => {
      const previousData = queryClient.getQueryData<InfiniteCommentsData>(
        storyKeys.commentsInfinite(workspaceSlug, newComment.storyId),
      );

      if (previousData) {
        const updatedData = {
          ...previousData,
          pages: previousData.pages.map((page) => ({
            ...page,
            comments: page.comments.map((comment) =>
              comment.id === newComment.commentId
                ? { ...comment, comment: newComment.payload.content }
                : comment,
            ),
          })),
        };

        queryClient.setQueryData<InfiniteCommentsData>(
          storyKeys.commentsInfinite(workspaceSlug, newComment.storyId),
          updatedData,
        );
      }
      return { previousData };
    },
    onError: (error, variables, context) => {
      if (context?.previousData) {
        queryClient.setQueryData(
          storyKeys.commentsInfinite(workspaceSlug, variables.storyId),
          context.previousData,
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
      queryClient.invalidateQueries({
        queryKey: storyKeys.commentsInfinite(workspaceSlug, variables.storyId),
      });
    },
    onSuccess: (res, { storyId }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: storyKeys.commentsInfinite(workspaceSlug, storyId),
      });
    },
  });

  return mutation;
};
