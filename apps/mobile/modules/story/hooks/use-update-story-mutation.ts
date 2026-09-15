import { useQueryClient } from "@tanstack/react-query";
import { useSessionMutation } from "@/lib/use-session-mutation";
import { toast } from "sonner-native";
import { updateStoryAction } from "../actions/update-story";
import { homeKeys, searchKeys, storyKeys } from "@/constants/keys";
import type { DetailedStory } from "@/modules/stories/types";
import {
  optimisticallyUpdateStory,
  restoreStoryCache,
} from "@/modules/stories/utils/cache";

export const useUpdateStoryMutation = () => {
  const client = useQueryClient();
  const queryKeys = [storyKeys.all, searchKeys.all];
  const overviewKey = homeKeys.overview();
  return useSessionMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: Partial<DetailedStory>;
    }) => updateStoryAction(storyId, payload),
    onMutate: async ({ storyId, payload }) => {
      const snapshots = await Promise.all(
        queryKeys.map((queryKey) =>
          optimisticallyUpdateStory(client, queryKey, storyId, payload),
        ),
      );
      return snapshots.flat();
    },
    onError: (error, _variables, snapshots) => {
      if (snapshots) restoreStoryCache(client, snapshots);
      toast.error("Failed to update task", { description: error.message });
    },
    onSettled: () =>
      Promise.all(
        [...queryKeys, overviewKey].map((queryKey) =>
          client.invalidateQueries({ queryKey }),
        ),
      ),
  });
};
