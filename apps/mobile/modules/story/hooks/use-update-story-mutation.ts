import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner-native";
import { updateStoryAction } from "../actions/update-story";
import { storyKeys } from "@/constants/keys";
import type { DetailedStory } from "@/modules/stories/types";
import {
  optimisticallyUpdateStory,
  restoreStoryCache,
} from "@/modules/stories/utils/cache";

export const useUpdateStoryMutation = () => {
  const client = useQueryClient();
  const queryKey = storyKeys.all;
  return useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: Partial<DetailedStory>;
    }) => updateStoryAction(storyId, payload),
    onMutate: ({ storyId, payload }) =>
      optimisticallyUpdateStory(client, queryKey, storyId, payload),
    onError: (error, _variables, snapshots) => {
      if (snapshots) restoreStoryCache(client, snapshots);
      toast.error("Failed to update task", { description: error.message });
    },
    onSettled: () => client.invalidateQueries({ queryKey }),
  });
};
