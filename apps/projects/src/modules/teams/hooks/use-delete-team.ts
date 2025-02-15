import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { teamKeys } from "@/constants/keys";
import { deleteTeam } from "../actions";

export const useDeleteTeamMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteTeam,
    onSuccess: () => {
      toast.success("Success", {
        description: "Team deleted successfully",
      });
      queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
    },
  });
};
