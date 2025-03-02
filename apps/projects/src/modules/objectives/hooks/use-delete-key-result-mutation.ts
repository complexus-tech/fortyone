import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { objectiveKeys } from "../constants";
import { deleteKeyResult } from "../actions/delete-key-result";
import type { KeyResult } from "../types";

type DeleteKeyResultVariables = {
  keyResultId: string;
  objectiveId: string;
};

export const useDeleteKeyResultMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ keyResultId, objectiveId }: DeleteKeyResultVariables) =>
      deleteKeyResult(keyResultId, objectiveId),

    onMutate: async ({ keyResultId, objectiveId }) => {
      await queryClient.cancelQueries({
        queryKey: objectiveKeys.keyResults(objectiveId),
      });

      const previousKeyResults = queryClient.getQueryData<KeyResult[]>(
        objectiveKeys.keyResults(objectiveId),
      );

      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(objectiveId),
        (old = []) => old.filter((keyResult) => keyResult.id !== keyResultId),
      );
      toast.success("Success", {
        description: "Key result deleted successfully",
      });

      return { previousKeyResults };
    },
    onError: (error, variables, context) => {
      if (context?.previousKeyResults) {
        queryClient.setQueryData<KeyResult[]>(
          objectiveKeys.keyResults(variables.objectiveId),
          context.previousKeyResults,
        );
      }
      toast.error("Failed to delete key result", {
        description:
          error.message || "An error occurred while deleting the key result",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSettled: (_, __, { objectiveId }) => {
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.keyResults(objectiveId),
      });
    },
  });

  return mutation;
};
