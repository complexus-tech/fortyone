import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import type { Team, UpdateTeamInput } from "../types";
import { updateTeamAction } from "../actions/update-team";

export const useUpdateTeamMutation = (teamId: string) => {
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();
  const toastId = "update-team";

  const mutation = useMutation({
    mutationFn: (input: UpdateTeamInput) => updateTeamAction(teamId, input),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: teamKeys.detail(teamId) });
      const previousTeam = queryClient.getQueryData<Team>(
        teamKeys.detail(teamId),
      );
      queryClient.setQueryData(
        teamKeys.detail(teamId),
        (old: Team | undefined) => (old ? { ...old, ...input } : undefined),
      );
      toast.loading("Updating team...", {
        description: "Please wait while we update the team",
        id: toastId,
      });

      return { previousTeam };
    },
    onError: (error, variables, context) => {
      toast.error("Error", {
        description: error.message || "Failed to update team",
        id: toastId,
        action: {
          label: "Retry",
          onClick: () => {
            toast.dismiss(toastId);
            mutation.mutate(variables);
          },
        },
      });
      queryClient.setQueryData(teamKeys.detail(teamId), context?.previousTeam);
    },
    onSuccess: (res, input) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      // Track team update
      analytics.track("team_updated", {
        teamId,
        ...input,
      });

      toast.success("Success", {
        description: "Team updated successfully",
        id: toastId,
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });

  return mutation;
};
