import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { teamKeys } from "@/constants/keys";
import type { Team, UpdateTeamInput } from "../types";
import { updateTeam } from "../actions";

export const useUpdateTeamMutation = (teamId: string) => {
  const queryClient = useQueryClient();
  const toastId = "update-team";

  const mutation = useMutation({
    mutationFn: (input: UpdateTeamInput) => updateTeam(teamId, input),
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
    onSuccess: () => {
      toast.success("Success", {
        description: "Team updated successfully",
        id: toastId,
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
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
  });

  return mutation;
};
