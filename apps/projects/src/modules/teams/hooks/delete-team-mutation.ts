import { useMutation, useQueryClient } from "@tanstack/react-query";
import { teamKeys } from "@/constants/keys";
import type { Team } from "@/modules/teams/types";
import { deleteTeamAction } from "../actions/delete-team";

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
