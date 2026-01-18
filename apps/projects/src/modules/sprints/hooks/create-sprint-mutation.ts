import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useAnalytics } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import type { NewSprint, Sprint } from "../types";
import { createSprintAction } from "../actions/create-sprint";

export const useCreateSprintMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: (newSprint: NewSprint) =>
      createSprintAction(newSprint, workspaceSlug),
    onMutate: async (newSprint) => {
      await queryClient.cancelQueries({
        queryKey: sprintKeys.team(workspaceSlug, newSprint.teamId),
      });
      await queryClient.cancelQueries({
        queryKey: sprintKeys.lists(workspaceSlug),
      });

      const previousTeamSprints = queryClient.getQueryData<Sprint[]>(
        sprintKeys.team(workspaceSlug, newSprint.teamId),
      );
      const previousSprints = queryClient.getQueryData<Sprint[]>(
        sprintKeys.lists(workspaceSlug),
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
        queryClient.setQueryData<Sprint[]>(
          sprintKeys.team(workspaceSlug, newSprint.teamId),
          [optimisticSprint, ...previousTeamSprints],
        );
      }

      if (previousSprints) {
        queryClient.setQueryData<Sprint[]>(sprintKeys.lists(workspaceSlug), [
          optimisticSprint,
          ...previousSprints,
        ]);
      }

      return { previousTeamSprints, previousSprints };
    },
    onError: (error, newSprint, context) => {
      if (context?.previousTeamSprints) {
        queryClient.setQueryData(
          sprintKeys.team(workspaceSlug, newSprint.teamId),
          context.previousTeamSprints,
        );
      }
      if (context?.previousSprints) {
        queryClient.setQueryData(
          sprintKeys.lists(workspaceSlug),
          context.previousSprints,
        );
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
    onSuccess: (res, newSprint) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      const sprint = res.data!;

      // Track sprint creation
      analytics.track("sprint_created", {
        sprintId: sprint.id,
        name: sprint.name,
        teamId: sprint.teamId,
        hasObjective: Boolean(sprint.objectiveId),
        duration: {
          startDate: sprint.startDate,
          endDate: sprint.endDate,
        },
      });

      queryClient.invalidateQueries({
        queryKey: sprintKeys.lists(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: sprintKeys.team(workspaceSlug, newSprint.teamId),
      });

      toast.success("Success", {
        description: "Sprint created successfully",
      });
    },
  });

  return mutation;
};
