import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { objectiveKeys } from "../constants";
import { updateObjective } from "../actions/update-objective";
import type { ObjectiveUpdate } from "../types";

export type UpdateObjectiveVariables = {
  objectiveId: string;
  data: ObjectiveUpdate;
};

export const useUpdateObjectiveMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ objectiveId, data }: UpdateObjectiveVariables) =>
      updateObjective(objectiveId, data),
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
    onSuccess: () => {
      toast.success("Success", {
        description: "Objective updated successfully",
      });
    },
    onSettled: (_, __, { objectiveId }) => {
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.objective(objectiveId),
      });
      queryClient.invalidateQueries({ queryKey: objectiveKeys.list() });
    },
  });

  return mutation;
};
