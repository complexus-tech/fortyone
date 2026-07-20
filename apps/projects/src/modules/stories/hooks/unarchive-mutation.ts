import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { storyKeys } from "../constants";
import { bulkUnarchiveAction } from "../actions/bulk-unarchive-stories";

export const useBulkUnarchiveStoryMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });

  const mutation = useMutation({
    mutationFn: (storyIds: string[]) =>
      bulkUnarchiveAction(storyIds, workspaceSlug),
    onError: (error, storyIds) => {
      toast.error(`Failed to unarchive ${storyTermPlural}`, {
        description:
          error.message ||
          `An error occurred while unarchiving ${storyTermPlural}`,
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
        description: `${storyIds.length} ${getTermDisplay("storyTerm", {
          variant: storyIds.length === 1 ? "singular" : "plural",
        })} unarchived`,
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
    },
  });

  return mutation;
};
