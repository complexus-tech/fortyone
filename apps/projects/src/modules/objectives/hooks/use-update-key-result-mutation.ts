import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { objectiveKeys } from "../constants";
import { updateKeyResult } from "../actions/update-key-result";
import type { KeyResult, KeyResultUpdate } from "../types";

type UpdateKeyResultVariables = {
  keyResultId: string;
  objectiveId: string;
  data: KeyResultUpdate;
};

export const useUpdateKeyResultMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      keyResultId,
      objectiveId,
      data,
    }: UpdateKeyResultVariables) =>
      updateKeyResult(keyResultId, objectiveId, data),
    onError: (error, variables) => {
      toast.error("Failed to update key result", {
        description:
          error.message || "An error occurred while updating the key result",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onMutate: async ({ keyResultId, objectiveId, data }) => {
      await queryClient.cancelQueries({
        queryKey: objectiveKeys.keyResults(objectiveId),
      });

      const previousKeyResults = queryClient.getQueryData<KeyResult[]>(
        objectiveKeys.keyResults(objectiveId),
      );

      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(objectiveId),
        (old = []) =>
          old.map((keyResult) =>
            keyResult.id === keyResultId
              ? {
                  ...keyResult,
                  ...data,
                  updatedAt: new Date().toISOString(),
                }
              : keyResult,
          ),
      );

      return { previousKeyResults };
    },
    onSuccess: () => {
      toast.success("Success", {
        description: "Key result updated successfully",
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
