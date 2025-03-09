import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import { deleteTeamAction } from "../actions/delete-team";

export const useDeleteTeamMutation = () => {
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();
  const toastId = "delete-team";

  const mutation = useMutation({
    mutationFn: (id: string) => deleteTeamAction(id),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: teamKeys.lists() });

      toast.loading("Deleting team...", {
        description: "Please wait while we delete the team",
        id: toastId,
      });
    },
    onError: (error, variables) => {
      toast.dismiss(toastId);
      toast.error("Error", {
        description: error.message || "Failed to delete team",
        id: toastId,
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (_, teamId) => {
      // Track team deletion
      analytics.track("team_deleted", {
        teamId,
      });

      toast.success("Success", {
        description: "Team deleted successfully",
        id: toastId,
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });

  return mutation;
};
