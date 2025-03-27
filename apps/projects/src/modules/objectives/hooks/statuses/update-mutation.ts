import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { objectiveKeys } from "../../constants";
import type { UpdateObjectiveStatus } from "../../actions/statuses/update";
import { updateObjectiveStatusAction } from "../../actions/statuses/update";
import type { ObjectiveStatus } from "../../types";

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
      const previousStatuses = queryClient.getQueryData<ObjectiveStatus[]>(
        objectiveKeys.statuses(),
      );
      if (previousStatuses) {
        const updatedStatuses = previousStatuses.map((status) => {
          // If we're setting a new default status and this status is currently the default
          if (
            newStatus.payload.isDefault === true &&
            status.isDefault &&
            status.id !== newStatus.statusId
          ) {
            // Set the previous default status to not default
            return { ...status, isDefault: false };
          }

          // Update the target status with new payload
          if (status.id === newStatus.statusId) {
            return { ...status, ...newStatus.payload };
          }

          // Return other statuses unchanged
          return status;
        });

        queryClient.setQueryData<ObjectiveStatus[]>(
          objectiveKeys.statuses(),
          updatedStatuses,
        );
      }

      return { previousStatuses };
    },
    onError: (_, variables, context) => {
      if (context?.previousStatuses) {
        queryClient.setQueryData<ObjectiveStatus[]>(
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
