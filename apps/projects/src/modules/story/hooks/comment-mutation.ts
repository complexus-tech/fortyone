import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { storyKeys } from "@/modules/stories/constants";
import type { Comment } from "@/types";
import {
  commentMutationKey,
  insertCachedComment,
  invalidateCommentQueries,
  removeCachedComment,
  type InfiniteCommentsData,
} from "@/shared/story/comment-cache";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { useDesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";
import {
  commentDesktopStory,
  isDesktopPending,
} from "@/shell/desktop/story-offline";
import { retainDesktopComment } from "@/shell/desktop/outbox-projection";
import type { StoryCommentPayload } from "@/shared/story/comment-payload";
import { commentStoryAction } from "../actions/comment-story";

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
      payload: StoryCommentPayload;
    },
    { optimisticId?: string }
  >({
    mutationKey: getDesktopBridge()?.enqueueOperation
      ? ["desktop-comment", workspaceSlug]
      : commentMutationKey(workspaceSlug),
    networkMode: getDesktopBridge()?.enqueueOperation ? "always" : "online",
    mutationFn: async ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: StoryCommentPayload;
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
      const optimisticId = context?.optimisticId;
      if (optimisticId) {
        queryClient.setQueryData<InfiniteCommentsData>(
          storyKeys.commentsInfinite(workspaceSlug, variables.storyId),
          (current) => removeCachedComment(current, optimisticId),
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
    onMutate: async ({ storyId, payload }) => {
      if (getDesktopBridge()?.enqueueOperation) return {};
      const queryKey = storyKeys.commentsInfinite(workspaceSlug, storyId);
      await queryClient.cancelQueries({ queryKey, exact: true });
      const optimisticId = `pending-comment:${crypto.randomUUID()}`;
      const createdAt = new Date().toISOString();
      const newComment: Comment = {
        id: optimisticId,
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
        createdAt,
        updatedAt: createdAt,
        parentId: payload.parentId ?? null,
        subComments: [],
      };
      queryClient.setQueryData<InfiniteCommentsData>(queryKey, (current) =>
        insertCachedComment(current, newComment),
      );
      return { optimisticId };
    },
    onSuccess: (res, { storyId }, context) => {
      if (isDesktopPending(res) && res.data && desktopScope) {
        retainDesktopComment(queryClient, desktopScope, res.data);
        return;
      }

      if (res.data) {
        const comment = res.data;
        queryClient.setQueryData<InfiniteCommentsData>(
          storyKeys.commentsInfinite(workspaceSlug, storyId),
          (current) =>
            insertCachedComment(
              context.optimisticId
                ? removeCachedComment(current, context.optimisticId)
                : current,
              comment,
            ),
        );
      }
    },
    onSettled: (res, _error, { storyId }) => {
      if (isDesktopPending(res)) return;
      return Promise.all([
        invalidateCommentQueries(queryClient, workspaceSlug, storyId),
        queryClient.invalidateQueries({
          queryKey: storyKeys.mayaRuns(workspaceSlug, storyId),
        }),
      ]);
    },
  });

  return mutation;
};
