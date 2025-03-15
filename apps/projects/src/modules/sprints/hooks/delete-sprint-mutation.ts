import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import type { Sprint } from "../types";
import { deleteSprintAction } from "../actions/delete-sprint";

export const useDeleteSprintMutation = () => {
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: deleteSprintAction,
    onMutate: async (sprintId) => {
      await queryClient.cancelQueries({ queryKey: sprintKeys.lists() });
      const previousSprints = queryClient.getQueryData<Sprint[]>(
        sprintKeys.lists(),
      );

      if (previousSprints) {
        queryClient.setQueryData<Sprint[]>(
          sprintKeys.lists(),
          previousSprints.filter((s) => s.id !== sprintId),
        );
      }

      return { previousSprints };
    },
    onError: (error, sprintId, context) => {
      if (context?.previousSprints) {
        queryClient.setQueryData(sprintKeys.lists(), context.previousSprints);
      }
      toast.error("Failed to delete sprint", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(sprintId);
          },
        },
      });
    },
    onSuccess: (res, sprintId) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: sprintKeys.lists() });
      analytics.track("sprint_deleted", {
        sprintId,
      });

      toast.success("Success", {
        description: "Sprint deleted successfully",
      });
    },
  });

  return mutation;
};
