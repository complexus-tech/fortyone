import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { objectiveKeys } from "../constants";
import { createObjective } from "../actions/create-objective";
import type { Objective } from "../types";

export const useCreateObjectiveMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createObjective,
    onError: (error, variables) => {
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
    onMutate: (newObjective) => {
      const previousObjectives = queryClient.getQueryData<Objective[]>(
        objectiveKeys.list(),
      );

      const optimisticObjective: Objective = {
        ...newObjective,
        id: "optimistic",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        workspaceId: "optimistic",
        description: newObjective.description || "",
        leadUser: newObjective.leadUser || "",
        teamId: newObjective.teamId || "",
        startDate: newObjective.startDate || "",
        endDate: newObjective.endDate || "",
        stats: {
          total: 0,
          cancelled: 0,
          completed: 0,
          started: 0,
          unstarted: 0,
          backlog: 0,
        },
      };

      if (previousObjectives) {
        queryClient.setQueryData<Objective[]>(objectiveKeys.list(), [
          ...previousObjectives,
          optimisticObjective,
        ]);
      }
      return { previousObjectives };
    },
    onSuccess: () => {
      toast.success("Success", {
        description: "Objective created successfully",
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: objectiveKeys.list() });
    },
  });

  return mutation;
};
