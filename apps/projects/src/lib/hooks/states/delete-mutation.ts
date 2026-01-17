import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import { useWorkspacePath } from "@/hooks";
import { statusKeys } from "@/constants/keys";
import type { State } from "@/types/states";
import { deleteStateAction } from "../../actions/states/delete";

export const useDeleteStateMutation = () => {
  const queryClient = useQueryClient();
  const { teamId } = useParams<{ teamId: string }>();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (stateId: string) => deleteStateAction(stateId, workspaceSlug),
    onMutate: (stateId) => {
      const previousStates = queryClient.getQueryData<State[]>(
        statusKeys.team(workspaceSlug, teamId),
      );
      if (previousStates) {
        queryClient.setQueryData<State[]>(
          statusKeys.team(workspaceSlug, teamId),
          previousStates.filter((state) => state.id !== stateId),
        );
      }
      return { previousStates };
    },
    onError: (error, variables, context) => {
      if (context?.previousStates) {
        queryClient.setQueryData<State[]>(
          statusKeys.team(workspaceSlug, teamId),
          context.previousStates,
        );
      }
      toast.error("Failed to delete state", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
      queryClient.invalidateQueries({
        queryKey: statusKeys.team(workspaceSlug, teamId),
      });
      queryClient.invalidateQueries({
        queryKey: statusKeys.lists(workspaceSlug),
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: statusKeys.lists(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: statusKeys.team(workspaceSlug, teamId),
      });
    },
  });

  return mutation;
};
