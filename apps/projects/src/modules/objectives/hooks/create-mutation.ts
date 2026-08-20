import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "@/lib/auth/client";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import { createObjective } from "../actions/create-objective";
import type { Objective, NewObjective } from "../types";

export const useCreateObjectiveMutation = () => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
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
        id: "optimistic",
        sequenceId: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        workspaceId: "optimistic",
        isPrivate: false,
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
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("objectives") &&
          query.isActive() &&
          queryKey.toLowerCase().includes("list")
        ) {
          queryClient.cancelQueries({ queryKey: query.queryKey });
          queryClient.setQueryData<Objective[]>(
            query.queryKey,
            (prev: Objective[] = []) => {
              return [...prev, optimisticObjective];
            },
          );
        }
      });
    },
    onError: (error, variables) => {
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("objectives") &&
          query.isActive() &&
          queryKey.toLowerCase().includes("list")
        ) {
          queryClient.setQueryData<Objective[]>(
            query.queryKey,
            (prev: Objective[] = []) => {
              return prev.filter((objective) => objective.id !== "optimistic");
            },
          );
        }
      });

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
    onSuccess: (objective) => {
      analytics.track("objective_created", {
        name: objective.name,
        startDate: objective.startDate,
        priority: objective.priority,
      });
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("objectives") &&
          query.isActive() &&
          queryKey.toLowerCase().includes("list")
        ) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });
    },
  });

  return mutation;
};
