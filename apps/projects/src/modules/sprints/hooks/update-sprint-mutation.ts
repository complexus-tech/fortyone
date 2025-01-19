import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { sprintKeys } from "@/constants/keys";
import type { Sprint, UpdateSprint } from "../types";
import { updateSprintAction } from "../actions/update-sprint";

export const useUpdateSprintMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      sprintId,
      updates,
    }: {
      sprintId: string;
      updates: UpdateSprint;
    }) => updateSprintAction(sprintId, updates),
    onMutate: async ({ sprintId, updates }) => {
      await queryClient.cancelQueries({ queryKey: sprintKeys.lists() });

      const previousSprints = queryClient.getQueryData<Sprint[]>(
        sprintKeys.lists(),
      );

      if (previousSprints) {
        queryClient.setQueryData<Sprint[]>(
          sprintKeys.lists(),
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
    onError: (_, variables) => {
      toast.error("Failed to update sprint", {
        description: "Your changes were not saved",
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
        description: "Sprint updated successfully",
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: sprintKeys.lists() });
    },
  });

  return mutation;
};
