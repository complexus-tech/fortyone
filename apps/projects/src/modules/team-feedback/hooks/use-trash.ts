import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { feedbackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import {
  restoreTeamFeedbackAction,
  trashTeamFeedbackAction,
} from "../actions/trash";

const getMutationError = (
  response: Awaited<ReturnType<typeof trashTeamFeedbackAction>>,
  fallback: string,
) => {
  if (response.error?.message) {
    throw new Error(response.error.message);
  }
  if (!("data" in response)) {
    throw new Error(fallback);
  }
  return response;
};

export const useRestoreTeamFeedback = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async (feedbackId: string) =>
      getMutationError(
        await restoreTeamFeedbackAction(feedbackId, workspaceSlug),
        "Feedback could not be restored",
      ),
    onError: (error) => {
      toast.error("Failed to restore feedback", {
        description: error.message,
      });
    },
    onSuccess: () => {
      toast.success("Feedback restored");
    },
    onSettled: () => {
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.lists(workspaceSlug),
      });
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.teamSummaries(workspaceSlug),
      });
    },
  });
};

export const useTrashTeamFeedback = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async (feedbackId: string) =>
      getMutationError(
        await trashTeamFeedbackAction(feedbackId, workspaceSlug),
        "Feedback could not be moved to trash",
      ),
    onError: (error) => {
      toast.error("Failed to move feedback to trash", {
        description: error.message,
      });
    },
    onSuccess: () => {
      toast.success("Feedback moved to trash", {
        description: "It can be restored for the next 30 days.",
      });
    },
    onSettled: () => {
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.lists(workspaceSlug),
      });
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.teamSummaries(workspaceSlug),
      });
    },
  });
};
