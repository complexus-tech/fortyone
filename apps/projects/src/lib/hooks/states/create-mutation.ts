import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { statusKeys } from "@/constants/keys";
import type { State } from "@/types/states";
import type { NewState } from "../../actions/states/create";
import { createStateAction } from "../../actions/states/create";

export const useCreateStateMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (newState: NewState) => createStateAction(newState),

    onMutate: (newState) => {
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

    onError: (_, variables, context) => {
      if (context?.previousStates) {
        queryClient.setQueryData<State[]>(
          statusKeys.lists(),
          context.previousStates,
        );
      }
      toast.error("Failed to create state", {
        description: "Your changes were not saved",
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
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: statusKeys.lists(),
      });
    },
  });

  return mutation;
};
