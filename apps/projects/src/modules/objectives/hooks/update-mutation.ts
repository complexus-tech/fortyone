import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { objectiveKeys } from "../constants";
import { updateObjective } from "../actions/update-objective";
import type { Objective, ObjectiveUpdate } from "../types";

export type UpdateObjectiveVariables = {
  objectiveId: string;
  data: ObjectiveUpdate;
};

export const useUpdateObjectiveMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ objectiveId, data }: UpdateObjectiveVariables) =>
      updateObjective(objectiveId, data),

    onMutate: ({ objectiveId, data }) => {
      const prevObjective = queryClient.getQueryData<Objective>(
        objectiveKeys.objective(objectiveId),
      );

      const prevTeamObjectives = queryClient.getQueryData<Objective[]>(
        objectiveKeys.list(),
      );

      if (prevObjective) {
        queryClient.setQueryData<Objective>(
          objectiveKeys.objective(objectiveId),
          {
            ...prevObjective,
            name: data.name ?? prevObjective.name,
            description: data.description ?? prevObjective.description,
            leadUser: data.leadUser ?? prevObjective.leadUser,
            startDate: data.startDate ?? prevObjective.startDate,
            endDate: data.endDate ?? prevObjective.endDate,
            statusId: data.statusId ?? prevObjective.statusId,
            priority: data.priority ?? prevObjective.priority,
          },
        );
      }

      if (prevTeamObjectives) {
        queryClient.setQueryData<Objective[]>(
          objectiveKeys.list(),
          prevTeamObjectives.map((objective) =>
            objective.id === objectiveId
              ? {
                  ...objective,
                  leadUser: data.leadUser ?? objective.leadUser,
                  startDate: data.startDate ?? objective.startDate,
                  endDate: data.endDate ?? objective.endDate,
                  statusId: data.statusId ?? objective.statusId,
                  priority: data.priority ?? objective.priority,
                  health: data.health ?? objective.health,
                }
              : objective,
          ),
        );
      }
      return { prevObjective };
    },
    onError: (error, variables) => {
      toast.error("Failed to update objective", {
        description:
          error.message || "An error occurred while updating the objective",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (_, { objectiveId }) => {
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.objective(objectiveId),
      });
      queryClient.invalidateQueries({ queryKey: objectiveKeys.list() });
    },
  });

  return mutation;
};
