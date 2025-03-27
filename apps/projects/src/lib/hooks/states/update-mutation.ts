import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import { statusKeys } from "@/constants/keys";
import type { State } from "@/types/states";
import type { UpdateState } from "../../actions/states/update";
import { updateStateAction } from "../../actions/states/update";

export const useUpdateStateMutation = () => {
  const queryClient = useQueryClient();
  const { teamId } = useParams<{ teamId: string }>();

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
        const updatedStates = previousStates.map((state) => {
          // If we're setting a new default status and this status is currently the default
          if (
            newState.payload.isDefault === true &&
            state.isDefault &&
            state.id !== newState.stateId
          ) {
            // Set the previous default status to not default
            return { ...state, isDefault: false };
          }

          // Update the target status with new payload
          if (state.id === newState.stateId) {
            return { ...state, ...newState.payload };
          }

          // Return other statuses unchanged
          return state;
        });
        queryClient.setQueryData<State[]>(statusKeys.lists(), updatedStates);
        if (teamId) {
          queryClient.setQueryData<State[]>(
            statusKeys.team(teamId),
            updatedStates.filter((state) => state.teamId === teamId),
          );
        }
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
