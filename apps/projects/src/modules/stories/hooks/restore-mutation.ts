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
          onClick: () => {
            mutation.mutate(storyIds);
          },
        },
      });
    },
    onSuccess: (_, storyIds) => {
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail")
        ) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });
      toast.success("Success", {
        description: `${storyIds.length} stories restored`,
      });
    },
    onSettled: (_, __, storyIds) => {
      storyIds.forEach((storyId) => {
        queryClient.invalidateQueries({ queryKey: storyKeys.detail(storyId) });
      });
    },
  });

  return mutation;
};
