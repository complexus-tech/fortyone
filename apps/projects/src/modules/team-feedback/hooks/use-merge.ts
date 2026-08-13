import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { feedbackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { mergeTeamFeedbackAction } from "../actions/merge";

export const useMergeTeamFeedback = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async ({
      sourceItemId,
      targetItemId,
    }: {
      sourceItemId: string;
      targetItemId: string;
    }) => {
      const response = await mergeTeamFeedbackAction(
        sourceItemId,
        targetItemId,
        workspaceSlug,
      );
      if (!response.data) {
        throw new Error(
          response.error?.message ?? "Feedback could not be merged",
        );
      }
      return response.data;
    },
    onError: (error) => {
      toast.error("Failed to merge feedback", {
        description: error.message,
      });
    },
    onSuccess: () => {
      toast.success("Feedback merged", {
        description:
          "Followers and linked work now point to the canonical request.",
      });
    },
    onSettled: () => {
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.lists(workspaceSlug),
      });
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.details(workspaceSlug),
      });
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.updates(workspaceSlug),
      });
    },
  });
};
