import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "../constants";
import { bulkRestoreAction } from "../actions/bulk-restore-stories";

export const useBulkRestoreStoryMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: bulkRestoreAction,

    onMutate: (storyIds) => {
      // For restore operations, we rely on invalidation rather than optimistic updates
      // since we don't have the full story data available at this point
      return { storyIds };
    },

    onError: (error, storyIds) => {
      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
      toast.error("Failed to restore stories", {
        description:
          error.message || "An error occurred while restoring stories",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(storyIds);
          },
        },
      });
    },

    onSuccess: (res, storyIds) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
      toast.success("Success", {
        description: `${storyIds.length} stories restored`,
      });
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
    },
  });

  return mutation;
};
