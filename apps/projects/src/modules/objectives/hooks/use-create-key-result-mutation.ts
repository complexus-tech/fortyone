import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { analyticsKeys, keyResultKeys } from "@/constants/keys";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { objectiveKeys } from "../constants";
import { createKeyResult } from "../actions/create-key-result";
import type { KeyResult, NewObjectiveKeyResult } from "../types";

export const useCreateKeyResultMutation = () => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: async (newKeyResult: NewObjectiveKeyResult) => {
      const response = await createKeyResult(newKeyResult, workspaceSlug);
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      if (!response.data?.id) {
        throw new Error(
          "Key result creation did not return a created key result.",
        );
      }
      return response.data;
    },

    onMutate: async (newKeyResult: NewObjectiveKeyResult) => {
      await queryClient.cancelQueries({
        queryKey: objectiveKeys.keyResults(
          workspaceSlug,
          newKeyResult.objectiveId,
        ),
      });

      const previousKeyResults = queryClient.getQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, newKeyResult.objectiveId),
      );

      const optimisticKeyResultId = `optimistic:${crypto.randomUUID()}`;
      const optimisticKeyResult: KeyResult = {
        ...newKeyResult,
        id: optimisticKeyResultId,
        sequenceId: 0,
        createdBy: session?.user.id ?? "",
        lead: newKeyResult.lead || null,
        contributors: newKeyResult.contributors || [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, newKeyResult.objectiveId),
        (old = []) => [optimisticKeyResult, ...old],
      );

      return { optimisticKeyResultId, previousKeyResults };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, variables.objectiveId),
        (current) => {
          if (!context) return current ?? [];
          const keyResults = current ?? context.previousKeyResults;
          if (!keyResults) return [];

          return keyResults.filter(
            ({ id }) => id !== context.optimisticKeyResultId,
          );
        },
      );
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
    onSuccess: (keyResult, newKeyResult, context) => {
      analytics.track("key_result_created", {
        keyResultId: keyResult.id,
        objectiveId: keyResult.objectiveId,
        measurementType: keyResult.measurementType,
      });

      toast.success("Success", {
        id: "key-result-created",
        description: "Key result created successfully",
      });

      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, newKeyResult.objectiveId),
        (current = []) =>
          current.map((item) =>
            item.id === context.optimisticKeyResultId ? keyResult : item,
          ),
      );

      queryClient.invalidateQueries({
        queryKey: objectiveKeys.keyResults(
          workspaceSlug,
          newKeyResult.objectiveId,
        ),
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.list(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.activitiesInfinite(
          workspaceSlug,
          newKeyResult.objectiveId,
        ),
      });
      queryClient.invalidateQueries({
        queryKey: keyResultKeys.all(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: analyticsKeys.all(workspaceSlug),
      });
    },
  });

  return mutation;
};
