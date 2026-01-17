import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import type { ObjectiveStatus } from "../../types";
import { objectiveKeys } from "../../constants";
import { deleteObjectiveStatusAction } from "../../actions/statuses/delete";

export const useDeleteObjectiveStatusMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (statusId: string) => deleteObjectiveStatusAction(statusId, workspaceSlug),
    onMutate: (statusId) => {
      const previousStatuses = queryClient.getQueryData<ObjectiveStatus[]>(
        objectiveKeys.statuses(workspaceSlug),
      );
      if (previousStatuses) {
        queryClient.setQueryData<ObjectiveStatus[]>(
          objectiveKeys.statuses(workspaceSlug),
          previousStatuses.filter((status) => status.id !== statusId),
        );
      }
      return { previousStatuses };
    },
    onError: (error, variables, context) => {
      if (context?.previousStatuses) {
        queryClient.setQueryData<ObjectiveStatus[]>(
          objectiveKeys.statuses(workspaceSlug),
          context.previousStatuses,
        );
      }
      toast.error("Failed to delete status", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.statuses(workspaceSlug),
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: objectiveKeys.statuses(workspaceSlug),
      });
    },
  });

  return mutation;
};
