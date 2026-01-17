import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { useAnalytics } from "@/hooks";
import { objectiveKeys } from "../constants";
import { createKeyResult } from "../actions/create-key-result";
import type { KeyResult, NewObjectiveKeyResult } from "../types";

export const useCreateKeyResultMutation = () => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: (newKeyResult: NewObjectiveKeyResult) =>
      createKeyResult(newKeyResult, workspaceSlug),

    onMutate: async (newKeyResult: NewObjectiveKeyResult) => {
      await queryClient.cancelQueries({
        queryKey: objectiveKeys.keyResults(workspaceSlug, newKeyResult.objectiveId),
      });

      const previousKeyResults = queryClient.getQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, newKeyResult.objectiveId),
      );

      const optimisticKeyResult: KeyResult = {
        ...newKeyResult,
        id: "optimistic",
        createdBy: session?.user?.id || "",
        lead: newKeyResult.lead || null,
        contributors: newKeyResult.contributors || [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, newKeyResult.objectiveId),
        (old = []) => [optimisticKeyResult, ...old],
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
    onSuccess: (res, newKeyResult) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      analytics.track("key_result_created", {
        ...newKeyResult,
      });

      toast.success("Success", {
        id: "key-result-created",
        description: "Key result created successfully",
      });

      queryClient.invalidateQueries({
        queryKey: objectiveKeys.keyResults(workspaceSlug, newKeyResult.objectiveId),
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.activitiesInfinite(workspaceSlug, newKeyResult.objectiveId),
      });
    },
  });

  return mutation;
};
