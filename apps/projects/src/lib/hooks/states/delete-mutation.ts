import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import { statusKeys } from "@/constants/keys";
import type { State } from "@/types/states";
import { deleteStateAction } from "../../actions/states/delete";

export const useDeleteStateMutation = () => {
  const queryClient = useQueryClient();
  const { teamId } = useParams<{ teamId: string }>();

  const mutation = useMutation({
    mutationFn: (stateId: string) => deleteStateAction(stateId),
    onMutate: (stateId) => {
      const previousStates = queryClient.getQueryData<State[]>(
        statusKeys.team(teamId),
      );
      if (previousStates) {
        queryClient.setQueryData<State[]>(
          statusKeys.team(teamId),
          previousStates.filter((state) => state.id !== stateId),
        );
      }
      return { previousStates };
    },
    onError: (error, variables, context) => {
      if (context?.previousStates) {
        queryClient.setQueryData<State[]>(
          statusKeys.team(teamId),
          context.previousStates,
        );
      }
      toast.error("Failed to delete state", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
      queryClient.invalidateQueries({
        queryKey: statusKeys.team(teamId),
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
      queryClient.invalidateQueries({
        queryKey: statusKeys.team(teamId),
      });
    },
  });

  return mutation;
};
