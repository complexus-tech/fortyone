import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { statusKeys } from "@/constants/keys";
import type { State } from "@/types/states";
import { deleteStateAction } from "../../actions/states/delete";

export const useDeleteStateMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (stateId: string) => deleteStateAction(stateId),
    onMutate: (stateId) => {
      const previousStates = queryClient.getQueryData<State[]>(
        statusKeys.lists(),
      );
      if (previousStates) {
        queryClient.setQueryData<State[]>(
          statusKeys.lists(),
          previousStates.filter((state) => state.id !== stateId),
        );
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
      toast.error("Failed to delete state", {
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
