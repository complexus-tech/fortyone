import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useAnalytics } from "@/hooks";
import type { SearchResponse } from "@/modules/search/types";
import { objectiveKeys } from "../constants";
import { updateObjective } from "../actions/update-objective";
import type { Objective, ObjectiveUpdate } from "../types";

export type UpdateObjectiveVariables = {
  objectiveId: string;
  data: ObjectiveUpdate;
};

export const useUpdateObjectiveMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({ objectiveId, data }: UpdateObjectiveVariables) =>
      updateObjective(objectiveId, data, workspaceSlug),

    onMutate: ({ objectiveId, data }) => {
      const prevObjective = queryClient.getQueryData<Objective>(
        objectiveKeys.objective(workspaceSlug, objectiveId),
      );

      const prevTeamObjectives = queryClient.getQueryData<Objective[]>(
        objectiveKeys.list(workspaceSlug),
      );

      if (prevObjective) {
        queryClient.setQueryData<Objective>(
          objectiveKeys.objective(workspaceSlug, objectiveId),
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
          objectiveKeys.list(workspaceSlug),
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

      // Update search results if any exist
      queryClient
        .getQueriesData<SearchResponse>({ queryKey: ["search"] })
        .forEach(([queryKey, searchData]) => {
          if (searchData?.objectives) {
            queryClient.setQueryData<SearchResponse>(queryKey, {
              ...searchData,
              objectives: searchData.objectives.map((objective) =>
                objective.id === objectiveId
                  ? {
                      ...objective,
                      name: data.name ?? objective.name,
                      description: data.description ?? objective.description,
                      leadUser: data.leadUser ?? objective.leadUser,
                      startDate: data.startDate ?? objective.startDate,
                      endDate: data.endDate ?? objective.endDate,
                      statusId: data.statusId ?? objective.statusId,
                      priority: data.priority ?? objective.priority,
                      health: data.health ?? objective.health,
                    }
                  : objective,
              ),
            });
          }
        });

      return { prevObjective, prevTeamObjectives };
    },
    onError: (error, variables, context) => {
      if (context) {
        queryClient.setQueryData(
          objectiveKeys.objective(workspaceSlug, variables.objectiveId),
          context.prevObjective,
        );
        queryClient.setQueryData(
          objectiveKeys.list(workspaceSlug),
          context.prevTeamObjectives,
        );
      }
      toast.error("Failed to update objective", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res, { objectiveId, data }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      analytics.track("objective_updated", {
        objectiveId,
        ...data,
      });

      queryClient.invalidateQueries({
        queryKey: objectiveKeys.objective(workspaceSlug, objectiveId),
      });
      queryClient.invalidateQueries({ queryKey: objectiveKeys.list(workspaceSlug) });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.activitiesInfinite(workspaceSlug, objectiveId),
      });
    },
  });

  return mutation;
};
