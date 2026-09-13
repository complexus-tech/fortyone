import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner-native";
import { updateLabelsAction } from "../actions/update-labels";
import { storyKeys } from "@/constants/keys";
import {
  optimisticallyUpdateStory,
  restoreStoryCache,
} from "@/modules/stories/utils/cache";

export const useUpdateLabelsMutation = () => {
  const client = useQueryClient();
  const queryKey = storyKeys.all;
  return useMutation({
    mutationFn: ({ storyId, labels }: { storyId: string; labels: string[] }) =>
      updateLabelsAction(storyId, labels),
    onMutate: ({ storyId, labels }) =>
      optimisticallyUpdateStory(client, queryKey, storyId, { labels }),
    onError: (error, _variables, snapshots) => {
      if (snapshots) restoreStoryCache(client, snapshots);
      toast.error("Failed to update labels", { description: error.message });
    },
    onSettled: () => client.invalidateQueries({ queryKey }),
  });
};
