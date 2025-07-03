import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useAnalytics } from "@/hooks";
import { slugify } from "@/utils";
import type { Story, GroupedStoriesResponse } from "@/modules/stories/types";
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
            // Handle sub stories (flat array) - add a sub story if the story has a parent id
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
            // Handle grouped stories (main story lists)
            queryClient.setQueriesData(
              { queryKey: query.queryKey },
              (data: GroupedStoriesResponse | undefined) => {
                if (!data) return data;

                const newStory: Story = {
                  id: "123",
                  title: `${story.title || "Untitled"} (Copy)`,
                  description: story.description || "",
                  statusId: story.statusId || "",
                  sprintId: story.sprintId || null,
                  objectiveId: story.objectiveId || null,
                  keyResultId: story.keyResultId || null,
                  teamId: story.teamId || "",
                  workspaceId: story.workspaceId || "",
                  assigneeId: story.assigneeId || null,
                  reporterId: story.reporterId || "",
                  epicId: story.epicId || null,
                  sequenceId:
                    data.groups.reduce(
                      (max, group) =>
                        Math.max(
                          max,
                          ...group.stories.map((s) => s.sequenceId),
                        ),
                      0,
                    ) + 1,
                  priority: story.priority || "No Priority",
                  startDate: story.startDate || null,
                  endDate: story.endDate || null,
                  createdAt: new Date().toISOString(),
                  updatedAt: new Date().toISOString(),
                  labels: [],
                  subStories: [],
                };

                // Add to the first group (or create a new group if no groups exist)
                if (data.groups.length === 0) {
                  return {
                    ...data,
                    groups: [
                      {
                        key: "default",
                        totalCount: 1,
                        stories: [newStory],
                        loadedCount: 1,
                        hasMore: false,
                        nextPage: 1,
                      },
                    ],
                    meta: {
                      ...data.meta,
                      totalGroups: 1,
                    },
                  };
                }

                return {
                  ...data,
                  groups: data.groups.map((group, index) =>
                    index === 0
                      ? {
                          ...group,
                          stories: [...group.stories, newStory],
                          totalCount: group.totalCount + 1,
                          loadedCount: group.loadedCount + 1,
                        }
                      : group,
                  ),
                };
              },
            );
          }
        }
      });
    },

    onError: (error, variables) => {
      // Remove all stories with id 123
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
            (data: GroupedStoriesResponse | undefined) => {
              if (!data) return data;

              return {
                ...data,
                groups: data.groups.map((group) => ({
                  ...group,
                  stories: group.stories.filter((story) => story.id !== "123"),
                  totalCount: Math.max(0, group.totalCount - 1),
                  loadedCount: Math.max(0, group.loadedCount - 1),
                })),
              };
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
