import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { State } from "@/types/states";
import { objectiveKeys } from "../../constants";
import type { UpdateObjectiveStatus } from "../../actions/statuses/update";
import { updateObjectiveStatusAction } from "../../actions/statuses/update";

export const useUpdateObjectiveStatusMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      statusId,
      payload,
    }: {
      statusId: string;
      payload: UpdateObjectiveStatus;
    }) => updateObjectiveStatusAction(statusId, payload),

    onMutate: (newStatus) => {
      const previousStatuses = queryClient.getQueryData<State[]>(
        objectiveKeys.statuses(),
      );
      if (previousStatuses) {
        const updatedStatuses = previousStatuses.map((status) =>
          status.id === newStatus.statusId
            ? { ...status, ...newStatus.payload }
            : status,
        );
        queryClient.setQueryData<State[]>(
          objectiveKeys.statuses(),
          updatedStatuses,
        );
      }

      return { previousStatuses };
    },
    onError: (_, variables, context) => {
      if (context?.previousStatuses) {
        queryClient.setQueryData<State[]>(
          objectiveKeys.statuses(),
          context.previousStatuses,
        );
      }
      toast.error("Failed to update status", {
        description: "Your changes were not saved",
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
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.statuses(),
      });
    },
  });

  return mutation;
};
