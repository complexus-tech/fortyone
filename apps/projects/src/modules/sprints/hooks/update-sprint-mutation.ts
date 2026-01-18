import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useAnalytics } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import type { Sprint, UpdateSprint } from "../types";
import { updateSprintAction } from "../actions/update-sprint";

export const useUpdateSprintMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({
      sprintId,
      updates,
    }: {
      sprintId: string;
      updates: UpdateSprint;
    }) => updateSprintAction(sprintId, updates, workspaceSlug),
    onMutate: async ({ sprintId, updates }) => {
      await queryClient.cancelQueries({
        queryKey: sprintKeys.lists(workspaceSlug),
      });

      const previousSprints = queryClient.getQueryData<Sprint[]>(
        sprintKeys.lists(workspaceSlug),
      );

      if (previousSprints) {
        queryClient.setQueryData<Sprint[]>(
          sprintKeys.lists(workspaceSlug),
          previousSprints.map((s) =>
            s.id === sprintId
              ? {
                  ...s,
                  name: updates.name ?? s.name,
                  startDate: updates.startDate ?? s.startDate,
                  endDate: updates.endDate ?? s.endDate,
                  goal: updates.goal ?? s.goal,
                  objectiveId: updates.objectiveId ?? s.objectiveId,
                }
              : s,
          ),
        );
      }

      return { previousSprints };
    },
    onError: (error, variables, context) => {
      if (context?.previousSprints) {
        queryClient.setQueryData(
          sprintKeys.lists(workspaceSlug),
          context.previousSprints,
        );
      }
      toast.error("Failed to update sprint", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res, { sprintId, updates }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      // Track sprint update
      analytics.track("sprint_updated", {
        sprintId,
        ...updates,
      });
      queryClient.invalidateQueries({
        queryKey: sprintKeys.lists(workspaceSlug),
      });

      toast.success("Success", {
        description: "Sprint updated successfully",
      });
    },
  });

  return mutation;
};
