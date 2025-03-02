import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "next-auth/react";
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
  const { data: session } = useSession();
  const mutation = useMutation({
    mutationFn: ({
      keyResultId,
      objectiveId,
      data,
    }: UpdateKeyResultVariables) =>
      updateKeyResult(keyResultId, objectiveId, data),

    onMutate: async ({ keyResultId, objectiveId, data }) => {
      await queryClient.cancelQueries({
        queryKey: objectiveKeys.keyResults(objectiveId),
      });
      toast.success("Success", {
        description: "Key result updated successfully",
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
                  lastUpdatedBy: session?.user?.id || "",
                }
              : keyResult,
          ),
      );

      return { previousKeyResults };
    },
    onError: (error, variables, context) => {
      if (context?.previousKeyResults) {
        queryClient.setQueryData<KeyResult[]>(
          objectiveKeys.keyResults(variables.objectiveId),
          context.previousKeyResults,
        );
      }
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
    onSettled: (_, __, { objectiveId }) => {
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.keyResults(objectiveId),
      });
    },
  });

  return mutation;
};
