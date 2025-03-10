import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import nProgress from "nprogress";
import { useRouter } from "next/navigation";
import { useAnalytics } from "@/hooks";
import { slugify } from "@/utils";
import type { Story } from "@/modules/stories/types";
import { createStoryAction } from "../actions/create-story";

export const useCreateStoryMutation = () => {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: createStoryAction,

    onMutate: (story) => {
      // Invalidate all queries that contain "stories" in their query key
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail")
        ) {
          queryClient.setQueriesData(
            { queryKey: query.queryKey },
            (data: Story[]) => {
              return [
                ...data,
                {
                  ...story,
                  id: "123",
                  sequenceId: data.length + 1,
                  updatedAt: new Date().toISOString(),
                  createdAt: new Date().toISOString(),
                  labels: [],
                },
              ];
            },
          );
        }
      });
    },
    onError: (error, story) => {
      // remove all stories with id 123
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail")
        ) {
          queryClient.setQueriesData(
            { queryKey: query.queryKey },
            (data: Story[]) => {
              return data.filter((story) => story.id !== "123");
            },
          );
        }
      });

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
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail")
        ) {
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
