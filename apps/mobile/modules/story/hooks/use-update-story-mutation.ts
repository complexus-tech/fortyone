import { useMutation, useQueryClient } from "@tanstack/react-query";
import { updateStoryAction } from "../actions/update-story";
import { storyKeys } from "@/constants/keys";
import type { DetailedStory } from "@/modules/stories/types";

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

    onMutate: ({ storyId, payload }) => {
      // Cancel outgoing refetches
      queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });

      // Snapshot the previous value
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId)
      );

      // Get all active queries
      const activeQueries = queryClient.getQueryCache().getAll();

      // Update all story-related queries optimistically
      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.cancelQueries({ queryKey: query.queryKey });

          if (queryKey.toLowerCase().includes("detail")) {
            // Update detail queries
            updateDetailQuery(queryClient, query.queryKey, storyId, payload);
          } else {
            // Update list queries
            updateListQuery(queryClient, query.queryKey, storyId, payload);
          }
        }
      });

      // Update the specific story detail
      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          ...payload,
        });
        return { previousStory };
      }
    },

    onError: (error, variables, context) => {
      // Rollback on error
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(variables.storyId),
          context.previousStory
        );
      }

      // Invalidate all story queries
      queryClient.invalidateQueries({ queryKey: storyKeys.all });

      // TODO: Add toast error notification
      console.error("Failed to update story:", error);
    },

    onSuccess: (res, { storyId, payload }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      // TODO: Add analytics tracking
      console.log("Story updated successfully:", { storyId, payload });

      // Invalidate all story queries to ensure consistency
      queryClient.invalidateQueries({ queryKey: storyKeys.all });
    },
  });

  return mutation;
};

// Helper function to update detail queries
const updateDetailQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>
) => {
  queryClient.setQueryData<DetailedStory>(queryKey, (old) => {
    if (!old || old.id !== storyId) return old;
    return { ...old, ...payload };
  });
};

// Helper function to update list queries
const updateListQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>
) => {
  queryClient.setQueryData(queryKey, (old: any) => {
    if (!old) return old;

    // Handle different list query structures
    if (old.stories) {
      // Grouped stories response
      return {
        ...old,
        stories: old.stories.map((story: any) =>
          story.id === storyId ? { ...story, ...payload } : story
        ),
      };
    } else if (Array.isArray(old)) {
      // Simple array of stories
      return old.map((story: any) =>
        story.id === storyId ? { ...story, ...payload } : story
      );
    }

    return old;
  });
};
