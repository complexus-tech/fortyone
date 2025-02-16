import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { teamKeys } from "@/constants/keys";
import type { UpdateTeamInput } from "../types";
import { updateTeam } from "../actions";

export const useUpdateTeamMutation = (teamId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: UpdateTeamInput) => updateTeam(teamId, input),
    onSuccess: () => {
      toast.success("Success", {
        description: "Team updated successfully",
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });
};
