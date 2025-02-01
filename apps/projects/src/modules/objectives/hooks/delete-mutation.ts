import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { objectiveKeys } from "../constants";
import { deleteObjective } from "../actions/delete-objective";
import type { Objective } from "../types";

export const useDeleteObjectiveMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: deleteObjective,
    onError: (error, objectiveId) => {
      toast.error("Failed to delete objective", {
        description:
          error.message || "An error occurred while deleting the objective",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(objectiveId);
          },
        },
      });
    },
    onMutate: async (objectiveId) => {
      await queryClient.cancelQueries({ queryKey: objectiveKeys.list() });
      await queryClient.cancelQueries({
        queryKey: objectiveKeys.objective(objectiveId),
      });

      const previousObjectives = queryClient.getQueryData<Objective[]>(
        objectiveKeys.list(),
      );

      if (previousObjectives) {
        queryClient.setQueryData<Objective[]>(
          objectiveKeys.list(),
          previousObjectives.filter(
            (objective) => objective.id !== objectiveId,
          ),
        );
      }
      return { previousObjectives };
    },
    onSuccess: () => {
      toast.success("Success", {
        description: "Objective deleted successfully",
      });
    },
    onSettled: (_, __, objectiveId) => {
      queryClient.removeQueries({
        queryKey: objectiveKeys.objective(objectiveId),
      });
      queryClient.invalidateQueries({ queryKey: objectiveKeys.list() });
    },
  });

  return mutation;
};
