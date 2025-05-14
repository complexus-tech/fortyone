import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useAnalytics } from "@/hooks";
import { slugify } from "@/utils";
import type { Story } from "@/modules/stories/types";
import type { DetailedStory } from "../types";
import { duplicateStoryAction } from "../actions/duplicate-story";

export const useDuplicateStoryMutation = () => {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
    }: {
      storyId: string;
      story: Partial<DetailedStory>;
    }) => duplicateStoryAction(storyId),

    onMutate: ({ story }) => {
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
                          title: `${story.title} (Copy)`,
                          sequenceId: data.subStories.length + 1,
                          updatedAt: new Date().toISOString(),
                          createdAt: new Date().toISOString(),
                          labels: [],
                          subStories: [],
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
              (data: Story[] = []) => {
                return [
                  ...data,
                  {
                    ...story,
                    id: "123",
                    title: `${story.title} (Copy)`,
                    sequenceId: data.length + 1,
                    updatedAt: new Date().toISOString(),
                    createdAt: new Date().toISOString(),
                    subStories: [],
                    labels: [],
                  },
                ];
              },
            );
          }
        }
      });
    },

    onError: (error, variables) => {
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
            (data: Story[] = []) => {
              return data.filter((story) => story.id !== "123");
            },
          );
        }
      });

      toast.error("Failed to duplicate story", {
        description:
          error.message || "An error occurred while duplicating the story",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      const duplicatedStory = res.data!;

      analytics.track("story_created", {
        storyId: duplicatedStory.id,
        title: duplicatedStory.title,
        teamId: duplicatedStory.teamId,
        hasObjective: Boolean(duplicatedStory.objectiveId),
        hasSprint: Boolean(duplicatedStory.sprintId),
      });

      // Invalidate queries to get fresh data
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

      toast.success("Success", {
        description: "Story duplicated successfully",
        action: {
          label: "View story",
          onClick: () => {
            router.push(
              `/story/${duplicatedStory.id}/${slugify(duplicatedStory.title)}`,
            );
          },
        },
      });
    },
  });

  return mutation;
};
