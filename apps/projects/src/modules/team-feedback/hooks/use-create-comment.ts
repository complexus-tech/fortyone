import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { feedbackKeys } from "@/constants/keys";
import { createTeamFeedbackCommentAction } from "../actions/comment";

export const useCreateTeamFeedbackComment = (feedbackId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (body: string) =>
      createTeamFeedbackCommentAction(feedbackId, { body }, workspaceSlug),
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error("Failed to add comment", {
          description: response.error.message,
        });
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: feedbackKeys.detail(workspaceSlug, feedbackId),
      });
      queryClient.invalidateQueries({
        queryKey: feedbackKeys.lists(workspaceSlug),
      });
    },
  });
};
