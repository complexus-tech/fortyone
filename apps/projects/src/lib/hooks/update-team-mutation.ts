import { useMutation, useQueryClient } from "@tanstack/react-query";
import { updateTeamAction } from "@/lib/actions/teams/update-team";
import type { UpdateTeamInput } from "@/lib/actions/teams/update-team";
import type { Team } from "@/lib/queries/teams/get-team";
import { teamKeys } from "@/constants/keys";

export const useUpdateTeamMutation = (id: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: UpdateTeamInput) => updateTeamAction(id, input),
    onSuccess: (updatedTeam: Team) => {
      queryClient.setQueryData(teamKeys.detail(id), updatedTeam);
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });
};
