import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { ObjectiveStatus } from "../../types";
import { objectiveKeys } from "../../constants";
import { deleteObjectiveStatusAction } from "../../actions/statuses/delete";

export const useDeleteObjectiveStatusMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (statusId: string) => deleteObjectiveStatusAction(statusId),
    onMutate: (statusId) => {
      const previousStatuses = queryClient.getQueryData<ObjectiveStatus[]>(
        objectiveKeys.statuses(),
      );
      if (previousStatuses) {
        queryClient.setQueryData<ObjectiveStatus[]>(
          objectiveKeys.statuses(),
          previousStatuses.filter((status) => status.id !== statusId),
        );
      }
      return { previousStatuses };
    },
    onError: (error, variables, context) => {
      if (context?.previousStatuses) {
        queryClient.setQueryData<ObjectiveStatus[]>(
          objectiveKeys.statuses(),
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
        queryKey: objectiveKeys.statuses(),
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: objectiveKeys.statuses(),
      });
    },
  });

  return mutation;
};
