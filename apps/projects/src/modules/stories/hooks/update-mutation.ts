import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import { useAnalytics } from "@/hooks";
import type { DetailedStory } from "@/modules/story/types";
import { storyKeys } from "../constants";
import type { Story } from "../types";
import { bulkUpdateAction } from "../actions/bulk-update-stories";

export const useBulkUpdateStoriesMutation = () => {
  const queryClient = useQueryClient();
  const { storyId } = useParams<{ storyId?: string }>();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({
      storyIds,
      payload,
    }: {
      storyIds: string[];
      payload: Partial<DetailedStory>;
    }) => bulkUpdateAction({ storyIds, updates: payload }),

    onMutate: ({ storyIds, payload }) => {
      // Cancel all story-related queries
      const activeQueries = queryClient.getQueryCache().getAll();
      // Store previous states for rollback
      const previousQueryStates = new Map<string, unknown>();

      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.cancelQueries({ queryKey: query.queryKey });
          const previousData = queryClient.getQueryData(query.queryKey);
          previousQueryStates.set(queryKey, previousData);

          if (queryKey.toLowerCase().includes("detail")) {
            // Update detailed stories and sub-stories
            const parentStory = queryClient.getQueryData<DetailedStory>(
              query.queryKey,
            );
            if (parentStory) {
              // If this story itself is being updated
              if (storyIds.includes(parentStory.id)) {
                queryClient.setQueryData<DetailedStory>(query.queryKey, {
                  ...parentStory,
                  ...payload,
                });
              }
              // Update sub-stories if any are being updated
              else if (parentStory.subStories.length > 0) {
                queryClient.setQueryData<DetailedStory>(query.queryKey, {
                  ...parentStory,
                  subStories: parentStory.subStories.map((subStory) =>
                    storyIds.includes(subStory.id)
                      ? { ...subStory, ...payload }
                      : subStory,
                  ),
                });
              }
            }
          } else {
            // Update story lists
            queryClient.setQueryData<Story[]>(query.queryKey, (stories) =>
              stories?.map((story) =>
                storyIds.includes(story.id) ? { ...story, ...payload } : story,
              ),
            );
          }
        }
      });

      return { previousQueryStates };
    },

    onError: (error, variables, context) => {
      // Rollback all optimistic updates
      if (context?.previousQueryStates) {
        context.previousQueryStates.forEach((data, queryKey) => {
          try {
            const parsedQueryKey = JSON.parse(queryKey);
            queryClient.setQueryData(parsedQueryKey, data);
          } catch {
            // Skip invalid query keys
          }
        });
      }

      // Invalidate all stories queries as fallback
      queryClient.invalidateQueries({ queryKey: storyKeys.all });

      toast.error("Failed to update stories", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (res, { storyIds, payload }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      analytics.track("stories_bulk_updated", {
        storyIds,
        count: storyIds.length,
        ...payload,
      });

      // Invalidate all story-related queries
      const activeQueries = queryClient.getQueryCache().getAll();
      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });

      // Also invalidate current story detail if viewing one
      if (storyId) {
        queryClient.invalidateQueries({
          queryKey: storyKeys.detail(storyId),
        });
      }
    },
  });

  return mutation;
};
