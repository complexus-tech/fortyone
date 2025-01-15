import { useMutation, useQueryClient } from "@tanstack/react-query";
import { updateTeamAction } from "@/lib/actions/teams/update-team";
import type { UpdateTeamInput } from "@/lib/actions/teams/update-team";
import type { Team } from "@/lib/queries/teams/get-team";
import { teamKeys } from "@/constants/keys";

export const useUpdateTeamMutation = (id: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: UpdateTeamInput) => updateTeamAction(id, input),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: teamKeys.detail(id) });
      const previousTeam = queryClient.getQueryData<Team>(teamKeys.detail(id));

      if (previousTeam) {
        queryClient.setQueryData<Team>(teamKeys.detail(id), {
          ...previousTeam,
          ...input,
        });
      }

      return { previousTeam };
    },
    onSuccess: (updatedTeam: Team) => {
      queryClient.setQueryData(teamKeys.detail(id), updatedTeam);
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });
};
