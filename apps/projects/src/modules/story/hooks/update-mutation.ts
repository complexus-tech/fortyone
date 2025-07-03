import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { GroupedStoriesResponse } from "@/modules/stories/types";
import type { SearchResponse } from "@/modules/search/types";
import type { DetailedStory } from "../types";
import { updateStoryAction } from "../actions/update-story";

export const useUpdateStoryMutation = () => {
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: Partial<DetailedStory>;
    }) => updateStoryAction(storyId, payload),

    onMutate: ({ storyId, payload }) => {
      queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId),
      );

      const activeQueries = queryClient.getQueryCache().getAll();

      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.cancelQueries({ queryKey: query.queryKey });
          if (queryKey.toLowerCase().includes("detail")) {
            // Handle sub stories (flat array) - try to update sub stories if they exist
            const parentStory = queryClient.getQueryData<DetailedStory>(
              query.queryKey,
            );
            if (parentStory?.subStories) {
              queryClient.setQueryData<DetailedStory>(query.queryKey, {
                ...parentStory,
                subStories: parentStory.subStories.map((subStory) =>
                  subStory.id === storyId
                    ? { ...subStory, ...payload }
                    : subStory,
                ),
              });
            }
          } else {
            // Handle grouped stories (main story lists)
            queryClient.setQueryData<GroupedStoriesResponse>(
              query.queryKey,
              (data) => {
                if (!data) return data;

                return {
                  ...data,
                  groups: data.groups.map((group) => ({
                    ...group,
                    stories: group.stories.map((story) =>
                      story.id === storyId ? { ...story, ...payload } : story,
                    ),
                  })),
                };
              },
            );
          }
        }
      });

      // Update search results if any exist
      queryClient
        .getQueriesData<SearchResponse>({ queryKey: ["search"] })
        .forEach(([queryKey, data]) => {
          if (data?.stories) {
            queryClient.setQueryData<SearchResponse>(queryKey, {
              ...data,
              stories: data.stories.map((story) =>
                story.id === storyId ? { ...story, ...payload } : story,
              ),
            });
          }
        });

      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          ...payload,
        });
        return { previousStory };
      }
    },
    onError: (error, variables, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(variables.storyId),
          context.previousStory,
        );
      }
      // invalidate all stories
      queryClient.invalidateQueries({ queryKey: storyKeys.all });

      toast.error("Failed to update story", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res, { storyId, payload }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      analytics.track("story_updated", {
        storyId,
        ...payload,
      });
      const activeQueries = queryClient.getQueryCache().getAll();
      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });
    },
  });

  return mutation;
};
