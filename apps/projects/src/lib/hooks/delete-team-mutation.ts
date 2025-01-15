import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deleteTeamAction } from "@/lib/actions/teams/delete-team";
import { teamKeys } from "@/constants/keys";
import type { Team } from "@/lib/queries/teams/get-team";

export const useDeleteTeamMutation = (id: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => deleteTeamAction(id),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: teamKeys.lists() });
      const previousTeams = queryClient.getQueryData<Team[]>(teamKeys.lists());

      if (previousTeams) {
        queryClient.setQueryData<Team[]>(
          teamKeys.lists(),
          previousTeams.filter((team) => team.id !== id),
        );
      }

      return { previousTeams };
    },
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: teamKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });
};
