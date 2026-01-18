import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { objectiveKeys } from "../constants";
import { deleteKeyResult } from "../actions/delete-key-result";
import type { KeyResult } from "../types";

type DeleteKeyResultVariables = {
  keyResultId: string;
  objectiveId: string;
};

export const useDeleteKeyResultMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: ({ keyResultId }: DeleteKeyResultVariables) =>
      deleteKeyResult(keyResultId, workspaceSlug),

    onMutate: async ({ keyResultId, objectiveId }) => {
      await queryClient.cancelQueries({
        queryKey: objectiveKeys.keyResults(workspaceSlug, objectiveId),
      });

      const previousKeyResults = queryClient.getQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, objectiveId),
      );

      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, objectiveId),
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
          objectiveKeys.keyResults(workspaceSlug, variables.objectiveId),
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
    onSuccess: (res, { objectiveId }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

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
