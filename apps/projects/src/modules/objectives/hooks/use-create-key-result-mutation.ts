import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { objectiveKeys } from "../constants";
import { createKeyResult } from "../actions/create-key-result";
import type { KeyResult, NewObjectiveKeyResult } from "../types";

export const useCreateKeyResultMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createKeyResult,
    onError: (error, variables) => {
      toast.error("Failed to create key result", {
        description:
          error.message || "An error occurred while creating the key result",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onMutate: async (newKeyResult: NewObjectiveKeyResult) => {
      await queryClient.cancelQueries({
        queryKey: objectiveKeys.keyResults(newKeyResult.objectiveId),
      });

      const previousKeyResults = queryClient.getQueryData<KeyResult[]>(
        objectiveKeys.keyResults(newKeyResult.objectiveId),
      );

      const optimisticKeyResult: KeyResult = {
        ...newKeyResult,
        id: "optimistic",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(newKeyResult.objectiveId),
        (old = []) => [...old, optimisticKeyResult],
      );

      return { previousKeyResults };
    },
    onSuccess: () => {
      toast.success("Success", {
        description: "Key result created successfully",
      });
    },
    onSettled: (_, __, newKeyResult) => {
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.keyResults(newKeyResult.objectiveId),
      });
    },
  });

  return mutation;
};
