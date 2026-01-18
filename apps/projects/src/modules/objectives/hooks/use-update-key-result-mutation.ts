import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useAnalytics } from "@/hooks";
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
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({ keyResultId, data }: UpdateKeyResultVariables) =>
      updateKeyResult(keyResultId, data, workspaceSlug),

    onMutate: async ({ keyResultId, objectiveId, data }) => {
      await queryClient.cancelQueries({
        queryKey: objectiveKeys.keyResults(workspaceSlug, objectiveId),
      });
      toast.success("Success", {
        description: "Key result updated successfully",
      });
      const previousKeyResults = queryClient.getQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, objectiveId),
      );

      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, objectiveId),
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
    onError: (error, variables, context) => {
      if (context?.previousKeyResults) {
        queryClient.setQueryData<KeyResult[]>(
          objectiveKeys.keyResults(workspaceSlug, variables.objectiveId),
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
    onSuccess: (res, { objectiveId, keyResultId, data: updateData }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      analytics.track("key_result_updated", {
        keyResultId,
        objectiveId,
        ...updateData,
      });

      queryClient.invalidateQueries({
        queryKey: objectiveKeys.keyResults(workspaceSlug, objectiveId),
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.activitiesInfinite(workspaceSlug, objectiveId),
      });
    },
  });

  return mutation;
};
