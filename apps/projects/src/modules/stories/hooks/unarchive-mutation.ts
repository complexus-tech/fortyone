import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "../constants";
import { bulkUnarchiveAction } from "../actions/bulk-unarchive-stories";

export const useBulkUnarchiveStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: bulkUnarchiveAction,
    onError: (error, storyIds) => {
      toast.error("Failed to unarchive stories", {
        description:
          error.message || "An error occurred while unarchiving stories",
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
      toast.success("Success", {
        description: `${storyIds.length} stories unarchived`,
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: storyKeys.all });
    },
  });

  return mutation;
};
