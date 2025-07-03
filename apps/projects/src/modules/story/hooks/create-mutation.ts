import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useAnalytics } from "@/hooks";
import { slugify } from "@/utils";
import type { Story, GroupedStoriesResponse } from "@/modules/stories/types";
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
                  title: story.title || "Untitled",
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
    onError: (error, story) => {
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
