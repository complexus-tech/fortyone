import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { statusKeys } from "@/constants/keys";
import type { State } from "@/types/states";
import type { UpdateState } from "../../actions/states/update";
import { updateStateAction } from "../../actions/states/update";

export const useUpdateStateMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      stateId,
      payload,
    }: {
      stateId: string;
      payload: UpdateState;
    }) => updateStateAction(stateId, payload),

    onMutate: (newState) => {
      const previousStates = queryClient.getQueryData<State[]>(
        statusKeys.lists(),
      );
      if (previousStates) {
        const updatedStates = previousStates.map((state) =>
          state.id === newState.stateId
            ? { ...state, ...newState.payload }
            : state,
        );
        queryClient.setQueryData<State[]>(statusKeys.lists(), updatedStates);
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
      toast.error("Failed to update link", {
        description: error.message || "Your changes were not saved",
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

      queryClient.invalidateQueries({
        queryKey: statusKeys.lists(),
      });
    },
  });

  return mutation;
};
