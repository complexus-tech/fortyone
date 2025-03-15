import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { statusKeys } from "@/constants/keys";
import type { State } from "@/types/states";
import type { NewState } from "../../actions/states/create";
import { createStateAction } from "../../actions/states/create";

export const useCreateStateMutation = () => {
  const queryClient = useQueryClient();
  const toastId = "create-state";

  const mutation = useMutation({
    mutationFn: (newState: NewState) => createStateAction(newState),

    onMutate: (newState) => {
      toast.loading("Please wait...", {
        id: toastId,
        description: "Creating state...",
      });
      const optimisticState: State = {
        ...newState,
        id: "optimistic",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        color: "#000000",
        orderIndex: 50,
        workspaceId: "optimistic",
      };

      const previousStates = queryClient.getQueryData<State[]>(
        statusKeys.lists(),
      );
      if (previousStates) {
        queryClient.setQueryData<State[]>(statusKeys.lists(), [
          ...previousStates,
          optimisticState,
        ]);
      }

      return { previousStates };
    },

    onError: (error, variables, context) => {
      if (context?.previousStates) {
        queryClient.setQueryData<State[]>(
          statusKeys.lists(),
          context.previousStates,
        );
      }
      toast.error("Failed to create state", {
        description: error.message || "Your changes were not saved",
        id: toastId,
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
      queryClient.invalidateQueries({
        queryKey: statusKeys.lists(),
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      toast.success("State created", {
        id: toastId,
        description: "Your state has been created",
      });
      queryClient.invalidateQueries({
        queryKey: statusKeys.lists(),
      });
    },
  });

  return mutation;
};
