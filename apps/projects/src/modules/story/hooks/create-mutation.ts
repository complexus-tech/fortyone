import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import nProgress from "nprogress";
import { useRouter } from "next/navigation";
import { useAnalytics } from "@/hooks";
import { slugify } from "@/utils";
import type { Story } from "@/modules/stories/types";
import { storyKeys } from "@/modules/stories/constants";
import { createStoryAction } from "../actions/create-story";
import type { DetailedStory } from "../types";

export const useCreateStoryMutation = () => {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: createStoryAction,

    onMutate: (story) => {
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (queryKey.toLowerCase().includes("stories") && query.isActive()) {
          if (queryKey.toLowerCase().includes("detail")) {
            // add a sub story if the story has a parent id
            if (story.parentId) {
              queryClient.setQueriesData(
                { queryKey: query.queryKey },
                (data: DetailedStory | undefined) => {
                  if (data?.id === story.parentId && data?.subStories) {
                    return {
                      ...data,
                      subStories: [
                        ...data.subStories,
                        {
                          ...story,
                          id: "123",
                          sequenceId: data.subStories.length + 1,
                          updatedAt: new Date().toISOString(),
                          createdAt: new Date().toISOString(),
                          labels: [],
                        },
                      ],
                    };
                  }
                },
              );
            }
          } else {
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
    onSuccess: (res, story) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      const createdStory = res.data!;

      // Track story creation
      analytics.track("story_created", {
        storyId: createdStory.id,
        title: createdStory.title,
        teamId: createdStory.teamId,
        hasObjective: Boolean(createdStory.objectiveId),
        hasSprint: Boolean(createdStory.sprintId),
      });

      // Invalidate all queries that contain "stories" in their query key
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail") &&
          query.isActive()
        ) {
          queryClient.invalidateQueries({ queryKey: query.queryKey });
        }
      });

      if (story.parentId) {
        queryClient.invalidateQueries({
          queryKey: storyKeys.detail(story.parentId),
        });
      }

      toast.success("Success", {
        description: "Story created successfully",
        action: {
          label: "View story",
          onClick: () => {
            nProgress.start();
            router.push(
              `/story/${createdStory.id}/${slugify(createdStory.title)}`,
            );
          },
        },
      });
    },
  });

  return mutation;
};
