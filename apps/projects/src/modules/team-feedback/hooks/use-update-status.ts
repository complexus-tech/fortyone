import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { feedbackKeys } from "@/constants/keys";
import { updateTeamFeedbackStatusAction } from "../actions/update-status";
import type { TeamFeedbackItem, UpdateTeamFeedbackStatusInput } from "../types";

export const useUpdateTeamFeedbackStatus = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: ({
      feedbackId,
      payload,
    }: {
      feedbackId: string;
      payload: UpdateTeamFeedbackStatusInput;
    }) => updateTeamFeedbackStatusAction(feedbackId, payload, workspaceSlug),
    onMutate: async ({ feedbackId, payload }) => {
      const detailKey = feedbackKeys.detail(workspaceSlug, feedbackId);
      await queryClient.cancelQueries({ queryKey: detailKey });
      const previous = queryClient.getQueryData<TeamFeedbackItem>(detailKey);
      if (previous) {
        queryClient.setQueryData<TeamFeedbackItem>(detailKey, {
          ...previous,
          status: payload.status,
          roadmapSummary:
            payload.roadmapSummary ?? previous.roadmapSummary ?? null,
          updatedAt: new Date().toISOString(),
        });
      }
      return { previous };
    },
    onError: (error, { feedbackId }, context) => {
      if (context?.previous) {
        queryClient.setQueryData(
          feedbackKeys.detail(workspaceSlug, feedbackId),
          context.previous,
        );
      }
      toast.error("Failed to update feedback", {
        description: error.message || "Your changes were not saved",
      });
    },
    onSuccess: (response, { feedbackId }, context) => {
      if (response.error?.message) {
        if (context.previous) {
          queryClient.setQueryData(
            feedbackKeys.detail(workspaceSlug, feedbackId),
            context.previous,
          );
        }
        toast.error("Failed to update feedback", {
          description: response.error.message,
        });
        return;
      }
      toast.success("Feedback updated");
    },
    onSettled: (_response, _error, { feedbackId }) => {
      queryClient.invalidateQueries({
        queryKey: feedbackKeys.detail(workspaceSlug, feedbackId),
      });
      queryClient.invalidateQueries({
        queryKey: feedbackKeys.lists(workspaceSlug),
      });
    },
  });
};
