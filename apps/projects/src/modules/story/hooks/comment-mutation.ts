import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { storyKeys } from "@/modules/stories/constants";
import type { Comment } from "@/types";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { useDesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";
import {
  commentDesktopStory,
  isDesktopPending,
} from "@/shell/desktop/story-offline";
import { retainDesktopComment } from "@/shell/desktop/outbox-projection";
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
  const desktopScope = useDesktopWorkspaceScope();
  const { workspaceSlug } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const { data: session } = useSession();

  const mutation = useMutation<
    | Awaited<ReturnType<typeof commentStoryAction>>
    | Awaited<ReturnType<typeof commentDesktopStory>>,
    Error,
    {
      storyId: string;
      payload: {
        comment: string;
        parentId?: string | null;
        mentions: string[];
      };
    },
    { previousData: InfiniteCommentsData | undefined }
  >({
    networkMode: getDesktopBridge()?.enqueueOperation ? "always" : "online",
    mutationFn: async ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: {
        comment: string;
        parentId?: string | null;
        mentions: string[];
      };
    }) => {
      if (getDesktopBridge()?.enqueueOperation) {
        if (!desktopScope)
          throw new Error("The desktop workspace is not ready yet.");
        return commentDesktopStory(desktopScope, session, storyId, payload);
      }
      const response = await commentStoryAction(
        storyId,
        payload,
        workspaceSlug,
      );
      if (response.error?.message) throw new Error(response.error.message);
      return response;
    },
    onError: (error, variables, context) => {
      if (context?.previousData) {
        queryClient.setQueryData(
          storyKeys.commentsInfinite(workspaceSlug, variables.storyId),
          context.previousData,
        );
      }
      toast.error(`Failed to comment on ${getTermDisplay("storyTerm")}`, {
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
      if (getDesktopBridge()?.enqueueOperation)
        return { previousData: undefined };
      const previousData = queryClient.getQueryData<InfiniteCommentsData>(
        storyKeys.commentsInfinite(workspaceSlug, storyId),
      );

      if (previousData) {
        const newComment = {
          id: "new comment",
          userId: session?.user.id ?? "",
          user: {
            id: session?.user.id ?? "",
            username: session?.user.username ?? "",
            fullName: session?.user.name ?? "",
            avatarUrl: session?.user.image ?? null,
            isActive: true,
            isSystem: false,
          },
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
      if (isDesktopPending(res) && res.data && desktopScope) {
        retainDesktopComment(queryClient, desktopScope, res.data);
        return;
      }

      queryClient.invalidateQueries({
        queryKey: storyKeys.commentsInfinite(workspaceSlug, storyId),
      });
    },
  });

  return mutation;
};
