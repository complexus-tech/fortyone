import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { deleteObjective } from "../actions/delete-objective";
import { updateObjective } from "../actions/update-objective";
import { objectiveKeys } from "../constants";
import type { ObjectiveUpdate } from "../types";

type BulkUpdateObjectivesVariables = {
  objectiveIds: string[];
  data: ObjectiveUpdate;
};

const getBulkError = (responses: { error?: { message?: string } | null }[]) =>
  responses.find(({ error }) => error?.message)?.error?.message;

export const useBulkUpdateObjectivesMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async ({
      objectiveIds,
      data,
    }: BulkUpdateObjectivesVariables) => {
      const responses = await Promise.all(
        objectiveIds.map((objectiveId) =>
          updateObjective(objectiveId, data, workspaceSlug),
        ),
      );
      const error = getBulkError(responses);
      if (error) throw new Error(error);
      return objectiveIds;
    },
    onError: (error) => {
      toast.error("Failed to update objectives", {
        description: error.message || "Your changes were not saved",
      });
    },
    onSuccess: (objectiveIds) => {
      objectiveIds.forEach((objectiveId) => {
        void queryClient.invalidateQueries({
          queryKey: objectiveKeys.objective(workspaceSlug, objectiveId),
        });
      });
      void queryClient.invalidateQueries({
        queryKey: objectiveKeys.list(workspaceSlug),
      });
      toast.success("Objectives updated");
    },
  });
};

export const useBulkDeleteObjectivesMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async (objectiveIds: string[]) => {
      const responses = await Promise.all(
        objectiveIds.map((objectiveId) =>
          deleteObjective(objectiveId, workspaceSlug),
        ),
      );
      const error = getBulkError(responses);
      if (error) throw new Error(error);
      return objectiveIds;
    },
    onError: (error) => {
      toast.error("Failed to delete objectives", {
        description: error.message || "The objectives were not deleted",
      });
    },
    onSuccess: (objectiveIds) => {
      objectiveIds.forEach((objectiveId) => {
        queryClient.removeQueries({
          queryKey: objectiveKeys.objective(workspaceSlug, objectiveId),
        });
      });
      void queryClient.invalidateQueries({
        queryKey: objectiveKeys.list(workspaceSlug),
      });
      toast.success("Objectives deleted");
    },
  });
};
