import type { InfiniteData, QueryKey } from "@tanstack/react-query";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { feedbackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { setTeamFeedbackReadStateAction } from "../actions/read-state";
import type { TeamFeedbackItem, TeamFeedbackPage } from "../types";

type TeamFeedbackListCache =
  | TeamFeedbackItem[]
  | TeamFeedbackPage
  | InfiniteData<TeamFeedbackPage, number>;

type CachedFeedbackList = readonly [
  QueryKey,
  TeamFeedbackListCache | undefined,
];

const isTeamFeedbackPage = (value: unknown): value is TeamFeedbackPage =>
  Boolean(
    value &&
      typeof value === "object" &&
      "feedback" in value &&
      Array.isArray(value.feedback),
  );

const isInfiniteTeamFeedback = (
  value: unknown,
): value is InfiniteData<TeamFeedbackPage, number> =>
  Boolean(
    value &&
      typeof value === "object" &&
      "pages" in value &&
      Array.isArray(value.pages) &&
      value.pages.every(isTeamFeedbackPage),
  );

const updateFeedbackItem = (
  feedback: TeamFeedbackItem,
  feedbackId: string,
  readAt: string | null,
) => (feedback.id === feedbackId ? { ...feedback, readAt } : feedback);

const updateListCache = (
  data: TeamFeedbackListCache | undefined,
  feedbackId: string,
  readAt: string | null,
): TeamFeedbackListCache | undefined => {
  if (!data) return data;

  if (Array.isArray(data)) {
    return data.map((feedback) =>
      updateFeedbackItem(feedback, feedbackId, readAt),
    );
  }

  if (isInfiniteTeamFeedback(data)) {
    return {
      ...data,
      pages: data.pages.map((page) => ({
        ...page,
        feedback: page.feedback.map((feedback) =>
          updateFeedbackItem(feedback, feedbackId, readAt),
        ),
      })),
    };
  }

  if (isTeamFeedbackPage(data)) {
    return {
      ...data,
      feedback: data.feedback.map((feedback) =>
        updateFeedbackItem(feedback, feedbackId, readAt),
      ),
    };
  }

  return data;
};

export const useSetTeamFeedbackReadState = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async ({
      feedbackId,
      isRead,
    }: {
      feedbackId: string;
      isRead: boolean;
    }) => {
      const response = await setTeamFeedbackReadStateAction(
        feedbackId,
        isRead,
        workspaceSlug,
      );
      if (response.error?.message) throw new Error(response.error.message);
      if (!response.data)
        throw new Error("Feedback read state was not returned");
      return response.data;
    },
    onMutate: async ({ feedbackId, isRead }) => {
      const detailKey = feedbackKeys.detail(workspaceSlug, feedbackId);
      const listsKey = feedbackKeys.lists(workspaceSlug);

      await Promise.all([
        queryClient.cancelQueries({ queryKey: detailKey }),
        queryClient.cancelQueries({ queryKey: listsKey }),
      ]);

      const previousDetail =
        queryClient.getQueryData<TeamFeedbackItem>(detailKey);
      const previousLists = queryClient.getQueriesData<TeamFeedbackListCache>({
        queryKey: listsKey,
      }) as CachedFeedbackList[];
      const optimisticReadAt = isRead ? new Date().toISOString() : null;

      queryClient.setQueryData<TeamFeedbackItem>(detailKey, (feedback) =>
        feedback
          ? updateFeedbackItem(feedback, feedbackId, optimisticReadAt)
          : feedback,
      );
      queryClient.setQueriesData<TeamFeedbackListCache>(
        { queryKey: listsKey },
        (data) => updateListCache(data, feedbackId, optimisticReadAt),
      );

      return { previousDetail, previousLists };
    },
    onError: (error, { feedbackId }, context) => {
      if (context?.previousDetail) {
        queryClient.setQueryData(
          feedbackKeys.detail(workspaceSlug, feedbackId),
          context.previousDetail,
        );
      }
      context?.previousLists.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data);
      });

      toast.error("Failed to update feedback", {
        description: error.message || "The read state was not saved",
      });
    },
    onSuccess: ({ readAt }, { feedbackId }) => {
      queryClient.setQueryData<TeamFeedbackItem>(
        feedbackKeys.detail(workspaceSlug, feedbackId),
        (feedback) =>
          feedback
            ? updateFeedbackItem(feedback, feedbackId, readAt)
            : feedback,
      );
      queryClient.setQueriesData<TeamFeedbackListCache>(
        { queryKey: feedbackKeys.lists(workspaceSlug) },
        (data) => updateListCache(data, feedbackId, readAt),
      );
    },
    onSettled: (_data, _error, { feedbackId }) => {
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.detail(workspaceSlug, feedbackId),
      });
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.lists(workspaceSlug),
      });
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.teamSummaries(workspaceSlug),
      });
    },
  });
};
