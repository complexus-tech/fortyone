import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { feedbackKeys } from "@/constants/keys";
import { storyKeys } from "@/modules/stories/constants";
import { planTeamFeedbackAction } from "../actions/plan";
import type { PlanTeamFeedbackInput } from "../types";

export const usePlanTeamFeedback = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm", { capitalize: true });

  return useMutation({
    mutationFn: ({
      feedbackId,
      payload,
    }: {
      feedbackId: string;
      payload: PlanTeamFeedbackInput;
    }) => planTeamFeedbackAction(feedbackId, payload, workspaceSlug),
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error("Failed to plan feedback", {
          description: response.error.message,
        });
        return;
      }
      toast.success(
        response.data?.created ? `${storyTerm} created` : `${storyTerm} linked`,
      );
    },
    onSettled: (response, _error, { feedbackId }) => {
      queryClient.invalidateQueries({
        queryKey: feedbackKeys.detail(workspaceSlug, feedbackId),
      });
      queryClient.invalidateQueries({
        queryKey: feedbackKeys.lists(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.all(workspaceSlug),
      });
      if (response?.data?.storyId) {
        queryClient.invalidateQueries({
          queryKey: feedbackKeys.storyLinks(
            workspaceSlug,
            response.data.storyId,
          ),
        });
      }
    },
  });
};
