import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "next-auth/react";
import { useAnalytics } from "@/hooks";
import { objectiveKeys } from "../constants";
import { createKeyResult } from "../actions/create-key-result";
import type { KeyResult, NewObjectiveKeyResult } from "../types";

export const useCreateKeyResultMutation = () => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: createKeyResult,

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
        createdBy: session?.user?.id || "",
        lastUpdatedBy: session?.user?.id || "",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      queryClient.setQueryData<KeyResult[]>(
        objectiveKeys.keyResults(newKeyResult.objectiveId),
        (old = []) => [...old, optimisticKeyResult],
      );
      toast.success("Success", {
        description: "Key result created successfully",
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
    onSettled: (_, error, newKeyResult) => {
      // Track key result creation
      if (!error) {
        analytics.track("key_result_created", {
          ...newKeyResult,
        });
      }

      queryClient.invalidateQueries({
        queryKey: objectiveKeys.keyResults(newKeyResult.objectiveId),
      });
    },
  });

  return mutation;
};
