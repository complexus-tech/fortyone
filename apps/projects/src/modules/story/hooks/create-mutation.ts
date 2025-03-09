import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import nProgress from "nprogress";
import { useRouter } from "next/navigation";
import { useAnalytics } from "@/hooks";
import { slugify } from "@/utils";
import { createStoryAction } from "../actions/create-story";

export const useCreateStoryMutation = () => {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: createStoryAction,
    onMutate: async (newStory) => {
      // Cancel any outgoing refetches to avoid overwriting our optimistic update
      await queryClient.cancelQueries();

      // Get the query cache to find active story queries
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      // Store previous values of active story-related queries for potential rollback
      const previousQueries = new Map();

      // Apply optimistic updates only to active story list queries
      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        // Only target active queries that contain "stories"
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          const previousData = queryClient.getQueryData(query.queryKey);
          if (previousData && Array.isArray(previousData)) {
            previousQueries.set(query.queryKey, previousData);
            queryClient.setQueryData(query.queryKey, [
              ...previousData,
              newStory,
            ]);
          }
        }
      });

      // Return the previous values for potential rollback
      return { previousQueries };
    },
    onError: (error, story, context) => {
      // Rollback to previous values if there's an error
      if (context?.previousQueries) {
        context.previousQueries.forEach((data, queryKey) => {
          queryClient.setQueryData(queryKey, data);
        });
      }

      toast.error(`Failed to create story: ${story.title}`, {
        description:
          error.message || "An error occurred while creating the story",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(story);
          },
        },
      });
    },
    onSuccess: (story) => {
      // Track story creation
      analytics.track("story_created", {
        storyId: story.id,
        title: story.title,
        teamId: story.teamId,
        hasObjective: Boolean(story.objectiveId),
        hasSprint: Boolean(story.sprintId),
      });

      // Invalidate all queries that contain "stories" in their query key
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (queryKey.toLowerCase().includes("stories")) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });

      toast.success("Success", {
        description: "Story created successfully",
        action: {
          label: "View story",
          onClick: () => {
            nProgress.start();
            router.push(`/story/${story.id}/${slugify(story.title)}`);
          },
        },
      });
    },
  });

  return mutation;
};
