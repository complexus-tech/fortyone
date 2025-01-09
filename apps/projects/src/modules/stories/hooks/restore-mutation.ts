import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "../constants";
import { bulkRestoreAction } from "../actions/bulk-restore-stories";

export const useBulkRestoreStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: bulkRestoreAction,
    onError: (error, storyIds) => {
      toast.error("Failed to restore stories", {
        description:
          error.message || "An error occurred while restoring stories",
        action: {
          label: "Retry",
          onClick: () => { mutation.mutate(storyIds); },
        },
      });
    },
    onSuccess: (_, storyIds) => {
      toast.success("Success", {
        description: `${storyIds.length} stories restored`,
      });
    },
    onSettled: (_, __, storyIds) => {
      storyIds.forEach((storyId) => {
        queryClient.invalidateQueries({ queryKey: storyKeys.detail(storyId) });
      });
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
    },
  });

  return mutation;
};
