import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { State } from "@/types/states";
import { objectiveKeys } from "../../constants";
import type { NewObjectiveStatus } from "../../actions/statuses/create";
import { createObjectiveStatusAction } from "../../actions/statuses/create";

export const useCreateObjectiveStatusMutation = () => {
  const queryClient = useQueryClient();
  const toastId = "create-objective-status";

  const mutation = useMutation({
    mutationFn: (newStatus: NewObjectiveStatus) =>
      createObjectiveStatusAction(newStatus),

    onMutate: (newStatus) => {
      toast.loading("Please wait...", {
        id: toastId,
        description: "Creating status...",
      });
      const optimisticStatus: State = {
        ...newStatus,
        id: "optimistic",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        color: newStatus.color || "#000000",
        orderIndex: 50,
        workspaceId: "optimistic",
        teamId: "workspace",
      };

      const previousStatuses = queryClient.getQueryData<State[]>(
        objectiveKeys.statuses(),
      );
      if (previousStatuses) {
        queryClient.setQueryData<State[]>(objectiveKeys.statuses(), [
          ...previousStatuses,
          optimisticStatus,
        ]);
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
      toast.error("Failed to create status", {
        description: "Your changes were not saved",
        id: toastId,
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
      toast.success("Status created", {
        id: toastId,
        description: "Your status has been created",
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.statuses(),
      });
    },
  });

  return mutation;
};
