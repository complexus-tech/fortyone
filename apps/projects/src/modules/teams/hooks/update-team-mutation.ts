import { useMutation, useQueryClient } from "@tanstack/react-query";
import { updateTeamAction } from "@/modules/teams/update-team";
import type { UpdateTeamInput } from "@/modules/teams/update-team";
import type { Team } from "@/modules/teams/types";
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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: teamKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });
};
