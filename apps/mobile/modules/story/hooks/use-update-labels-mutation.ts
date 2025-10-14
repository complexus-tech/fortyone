import { useMutation, useQueryClient } from "@tanstack/react-query";
import { updateLabelsAction } from "../actions/update-labels";
import { storyKeys, labelKeys } from "@/constants/keys";
import type { DetailedStory } from "@/modules/stories/types";

export const useUpdateLabelsMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ storyId, labels }: { storyId: string; labels: string[] }) =>
      updateLabelsAction(storyId, labels),

    onMutate: ({ storyId, labels }) => {
      // Snapshot the previous value
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId)
      );

      // Update the story detail optimistically
      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          labels,
        });
      }

      // Get all active queries and update them
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (queryKey.toLowerCase().includes("stories") && query.isActive()) {
          if (queryKey.toLowerCase().includes("detail")) {
            updateDetailQuery(queryClient, query.queryKey, storyId, labels);
          } else {
            updateListQuery(queryClient, query.queryKey, storyId, labels);
          }
        }
      });

      return { previousStory };
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
      console.error("Failed to update labels:", error);
    },

    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      // Invalidate story and label queries
      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      queryClient.invalidateQueries({ queryKey: labelKeys.lists() });

      console.log("Labels updated successfully");
    },
  });

  return mutation;
};

// Helper function to update detail queries
const updateDetailQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  labels: string[]
) => {
  queryClient.setQueryData<DetailedStory>(queryKey, (old) => {
    if (!old || old.id !== storyId) return old;
    return { ...old, labels };
  });
};

// Helper function to update list queries
const updateListQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  labels: string[]
) => {
  queryClient.setQueryData(queryKey, (old: any) => {
    if (!old) return old;

    // Handle different list query structures
    if (old.stories) {
      // Grouped stories response
      return {
        ...old,
        stories: old.stories.map((story: any) =>
          story.id === storyId ? { ...story, labels } : story
        ),
      };
    } else if (Array.isArray(old)) {
      // Simple array of stories
      return old.map((story: any) =>
        story.id === storyId ? { ...story, labels } : story
      );
    }

    return old;
  });
};
