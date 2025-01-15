import { useMutation, useQueryClient } from "@tanstack/react-query";
import { deleteTeamAction } from "@/lib/actions/teams/delete-team";
import { teamKeys } from "@/constants/keys";

export const useDeleteTeamMutation = (id: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => deleteTeamAction(id),
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: teamKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });
};
