import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { Comment } from "@/types";
import { commentStoryAction } from "../actions/comment-story";

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

export const useCommentStoryMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: {
        comment: string;
        parentId?: string | null;
        mentions: string[];
      };
    }) =>
      commentStoryAction(storyId, {
        comment: payload.comment,
        mentions: payload.mentions,
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
      const previousData = queryClient.getQueryData<InfiniteCommentsData>(
        storyKeys.commentsInfinite(workspaceSlug, storyId),
      );

      if (previousData) {
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

        const updatedData = {
          ...previousData,
          pages: previousData.pages.map((page, pageIndex) => {
            if (pageIndex === 0) {
              // Add to first page
              if (payload.parentId) {
                // Add as reply to parent comment
                const updatedComments = page.comments.map((comment) => {
                  if (comment.id === payload.parentId) {
                    return {
                      ...comment,
                      subComments: [newComment, ...comment.subComments],
                    };
                  }
                  return comment;
                });
                return { ...page, comments: updatedComments };
              }
              // Add as new top-level comment
              return {
                ...page,
                comments: [newComment, ...page.comments],
              };
            }
            return page;
          }),
        };

        queryClient.setQueryData<InfiniteCommentsData>(
          storyKeys.commentsInfinite(workspaceSlug, storyId),
          updatedData,
        );
      }
      return { previousData };
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
