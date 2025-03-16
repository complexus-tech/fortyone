import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams, useRouter } from "next/navigation";
import { useAnalytics } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import { deleteTeamAction } from "../actions/delete-team";

export const useDeleteTeamMutation = () => {
  const { teamId: teamIdParam } = useParams<{ teamId?: string }>();
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();
  const toastId = "delete-team";
  const router = useRouter();

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
    onSuccess: (res, teamId) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      // Track team deletion
      analytics.track("team_deleted", {
        teamId,
      });
      toast.success("Success", {
        description: "Team deleted successfully",
        id: toastId,
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
      if (teamIdParam) {
        // If the team is deleted from the team page, redirect to the my work page
        router.push("/my-work");
      }
    },
  });

  return mutation;
};
