import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import type { Story } from "@/modules/stories/types";
import type { DetailedStory } from "../types";
import { updateStoryAction } from "../actions/update-story";

export const useUpdateStoryMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: Partial<DetailedStory>;
    }) => updateStoryAction(storyId, payload),
    onError: (_, variables) => {
      toast.error("Failed to update story", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => { mutation.mutate(variables); },
        },
      });
    },
    onMutate: ({ storyId, payload }) => {
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId),
      );
      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          ...payload,
        });
      }

      const activeQueries = queryClient
        .getQueryCache()
        .getAll()
        .filter((query) => query.isActive);

      activeQueries.forEach((query) => {
        if (
          query.queryKey.includes("stories") &&
          query.queryKey.includes("list")
        ) {
          queryClient.setQueryData<Story[]>(query.queryKey, (stories) =>
            stories?.map((story) =>
              story.id === storyId ? { ...story, ...payload } : story,
            ),
          );
        }
      });
    },
    onSettled: (storyId) => {
      queryClient.invalidateQueries({ queryKey: storyKeys.detail(storyId!) });
      queryClient.invalidateQueries({ queryKey: storyKeys.lists() });
      queryClient.invalidateQueries({ queryKey: storyKeys.teams() });
    },
  });

  return mutation;
};
