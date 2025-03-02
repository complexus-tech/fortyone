import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { sprintKeys } from "@/constants/keys";
import type { NewSprint, Sprint } from "../types";
import { createSprintAction } from "../actions/create-sprint";

export const useCreateSprintMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (newSprint: NewSprint) => createSprintAction(newSprint),
    onMutate: async (newSprint) => {
      await queryClient.cancelQueries({
        queryKey: sprintKeys.team(newSprint.teamId),
      });
      await queryClient.cancelQueries({ queryKey: sprintKeys.lists() });

      const previousTeamSprints = queryClient.getQueryData<Sprint[]>(
        sprintKeys.team(newSprint.teamId),
      );
      const previousSprints = queryClient.getQueryData<Sprint[]>(
        sprintKeys.lists(),
      );

      const optimisticSprint: Sprint = {
        id: "temp-id",
        name: newSprint.name,
        goal: newSprint.goal || "",
        objectiveId: newSprint.objectiveId || "",
        teamId: newSprint.teamId,
        workspaceId: "",
        startDate: newSprint.startDate,
        endDate: newSprint.endDate,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        stats: {
          total: 0,
          cancelled: 0,
          completed: 0,
          started: 0,
          unstarted: 0,
          backlog: 0,
        },
      };

      if (previousTeamSprints) {
        queryClient.setQueryData<Sprint[]>(sprintKeys.team(newSprint.teamId), [
          optimisticSprint,
          ...previousTeamSprints,
        ]);
      }

      if (previousSprints) {
        queryClient.setQueryData<Sprint[]>(sprintKeys.lists(), [
          optimisticSprint,
          ...previousSprints,
        ]);
      }

      return { previousTeamSprints, previousSprints };
    },
    onError: (error, newSprint, context) => {
      if (context?.previousTeamSprints) {
        queryClient.setQueryData(
          sprintKeys.team(newSprint.teamId),
          context.previousTeamSprints,
        );
      }
      if (context?.previousSprints) {
        queryClient.setQueryData(sprintKeys.lists(), context.previousSprints);
      }

      toast.error("Failed to create sprint", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(newSprint);
          },
        },
      });
    },
    onSuccess: () => {
      toast.success("Success", {
        description: "Sprint created successfully",
      });
    },
    onSettled: (_, __, newSprint) => {
      queryClient.invalidateQueries({ queryKey: sprintKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: sprintKeys.team(newSprint.teamId),
      });
    },
  });

  return mutation;
};
