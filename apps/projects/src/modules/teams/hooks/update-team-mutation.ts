import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { teamKeys } from "@/constants/keys";
import type { UpdateTeamInput } from "../actions/update-team";
import { updateTeamAction } from "../actions/update-team";
import type { Team } from "../types";

export const useUpdateTeamMutation = (id: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (input: UpdateTeamInput) =>
      updateTeamAction(id, input, workspaceSlug),
    onMutate: async (input) => {
      await queryClient.cancelQueries({
        queryKey: teamKeys.detail(workspaceSlug, id),
      });
      const previousTeam = queryClient.getQueryData<Team>(
        teamKeys.detail(workspaceSlug, id),
      );

      if (previousTeam) {
        queryClient.setQueryData<Team>(teamKeys.detail(workspaceSlug, id), {
          ...previousTeam,
          ...input,
        });
      }

      return { previousTeam };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(
        teamKeys.detail(workspaceSlug, id),
        context?.previousTeam,
      );
      toast.error("Error", {
        description: error.message || "Failed to update team",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: teamKeys.detail(workspaceSlug, id),
      });
      queryClient.invalidateQueries({
        queryKey: teamKeys.lists(workspaceSlug),
      });
    },
  });

  return mutation;
};
