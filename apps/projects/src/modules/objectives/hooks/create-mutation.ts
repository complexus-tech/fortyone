import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "@/lib/auth/client";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import { createObjective } from "../actions/create-objective";
import type { Objective, NewObjective } from "../types";
import { objectiveKeys } from "../constants";
import {
  optimisticallyCreateObjective,
  settleOptimisticObjective,
} from "./create-mutation-cache";

export const useCreateObjectiveMutation = () => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();
  const mutationKey = [...objectiveKeys.all(workspaceSlug), "create"] as const;

  const mutation = useMutation({
    mutationKey,
    mutationFn: async (newObjective: NewObjective) => {
      const response = await createObjective(newObjective, workspaceSlug);
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      if (!response.data?.objective.id) {
        throw new Error(
          "Objective creation did not return a created objective.",
        );
      }
      return response.data.objective;
    },

    onMutate: (newObjective) => {
      const optimisticObjective: Objective = {
        ...newObjective,
        id: `optimistic:${crypto.randomUUID()}`,
        sequenceId: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        workspaceId: "optimistic",
        isPrivate: newObjective.isPrivate ?? false,
        health: null,
        color: newObjective.color ?? "#4A90E2",
        forecastStartDate: null,
        forecastEndDate: null,
        scheduleStatus: "no_schedule",
        forecastDaysDelta: 0,
        forecastCauseStory: null,
        priority: newObjective.priority,
        statusId: newObjective.statusId,
        keyResultCount: newObjective.keyResults?.length ?? 0,
        description: newObjective.description || "",
        shortSummary: newObjective.shortSummary || null,
        leadUser: newObjective.leadUser || "",
        teamId: newObjective.teamId || "",
        startDate: newObjective.startDate || "",
        endDate: newObjective.endDate || "",
        createdBy: session?.user.id || "",
        stats: {
          total: 0,
          cancelled: 0,
          completed: 0,
          started: 0,
          unstarted: 0,
          backlog: 0,
        },
      };
      return optimisticallyCreateObjective(
        queryClient,
        workspaceSlug,
        optimisticObjective,
      );
    },
    onError: (error, variables, context) => {
      settleOptimisticObjective(queryClient, context);

      toast.error("Failed to create objective", {
        description:
          error.message || "An error occurred while creating the objective",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (objective, _variables, context) => {
      settleOptimisticObjective(queryClient, context, objective);
      analytics.track("objective_created", {
        name: objective.name,
        startDate: objective.startDate,
        priority: objective.priority,
      });
    },
    onSettled: () => {
      // Wait for sibling creates before refetching away their optimistic rows.
      if (queryClient.isMutating({ mutationKey }) === 1) {
        void queryClient.invalidateQueries({
          queryKey: objectiveKeys.list(workspaceSlug),
        });
      }
    },
  });

  return mutation;
};
