import type { InfiniteData } from "@tanstack/react-query";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { analyticsKeys } from "@/constants/keys";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import type { KeyResultListResponse } from "@/modules/key-results/types";
import { objectiveKeys } from "../constants";
import { updateKeyResult } from "../actions/update-key-result";
import type { KeyResult, KeyResultUpdate } from "../types";

type UpdateKeyResultVariables = {
  keyResultId: string;
  objectiveId: string;
  data: KeyResultUpdate;
  silent?: boolean;
};

type WorkspaceKeyResultsData = InfiniteData<KeyResultListResponse>;

export const useUpdateKeyResultMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: async ({ keyResultId, data }: UpdateKeyResultVariables) => {
      const response = await updateKeyResult(
        keyResultId,
        data.lead === null ? { ...data, clearLead: true } : data,
        workspaceSlug,
      );
      if (response.error?.message) throw new Error(response.error.message);
      return response;
    },

    onMutate: async ({ keyResultId, objectiveId, data }) => {
      const workspaceKeyResultsQuery = ["key-results", workspaceSlug] as const;

      await Promise.all([
        queryClient.cancelQueries({
          queryKey: objectiveKeys.keyResults(workspaceSlug, objectiveId),
        }),
        queryClient.cancelQueries({
          queryKey: workspaceKeyResultsQuery,
        }),
      ]);
      const previousKeyResults = queryClient.getQueryData<KeyResult[]>(
        objectiveKeys.keyResults(workspaceSlug, objectiveId),
      );
      const previousWorkspaceKeyResults =
        queryClient.getQueriesData<WorkspaceKeyResultsData>({
          queryKey: workspaceKeyResultsQuery,
        });

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
      queryClient.setQueriesData<WorkspaceKeyResultsData>(
        { queryKey: workspaceKeyResultsQuery },
        (old) =>
          old
            ? {
                ...old,
                pages: old.pages.map((page) => ({
                  ...page,
                  keyResults: page.keyResults.map((keyResult) =>
                    keyResult.id === keyResultId
                      ? {
                          ...keyResult,
                          ...data,
                          updatedAt: new Date().toISOString(),
                        }
                      : keyResult,
                  ),
                })),
              }
            : old,
      );

      return { previousKeyResults, previousWorkspaceKeyResults };
    },
    onError: (error, variables, context) => {
      if (context?.previousKeyResults) {
        queryClient.setQueryData<KeyResult[]>(
          objectiveKeys.keyResults(workspaceSlug, variables.objectiveId),
          context.previousKeyResults,
        );
      }
      for (const [queryKey, data] of context?.previousWorkspaceKeyResults ??
        []) {
        queryClient.setQueryData(queryKey, data);
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
    onSuccess: (_response, { objectiveId, keyResultId, data: updateData }) => {
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
      queryClient.invalidateQueries({
        queryKey: ["key-result-activities", workspaceSlug, keyResultId],
      });
      queryClient.invalidateQueries({
        queryKey: ["key-results", workspaceSlug],
      });
      queryClient.invalidateQueries({
        queryKey: analyticsKeys.all(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: ["strategy-map", workspaceSlug],
      });
    },
  });

  return mutation;
};
