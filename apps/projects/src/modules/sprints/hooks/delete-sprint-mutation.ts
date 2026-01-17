import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useAnalytics } from "@/hooks";
import { sprintKeys } from "@/constants/keys";
import type { Sprint } from "../types";
import { deleteSprintAction } from "../actions/delete-sprint";

export const useDeleteSprintMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: (sprintId: string) => deleteSprintAction(sprintId, workspaceSlug),
    onMutate: async (sprintId) => {
      await queryClient.cancelQueries({
        queryKey: sprintKeys.lists(workspaceSlug),
      });
      const previousSprints = queryClient.getQueryData<Sprint[]>(
        sprintKeys.lists(workspaceSlug),
      );

      if (previousSprints) {
        queryClient.setQueryData<Sprint[]>(
          sprintKeys.lists(workspaceSlug),
          previousSprints.filter((s) => s.id !== sprintId),
        );
      }

      return { previousSprints };
    },
    onError: (error, sprintId, context) => {
      if (context?.previousSprints) {
        queryClient.setQueryData(
          sprintKeys.lists(workspaceSlug),
          context.previousSprints,
        );
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

      queryClient.invalidateQueries({
        queryKey: sprintKeys.lists(workspaceSlug),
      });
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
