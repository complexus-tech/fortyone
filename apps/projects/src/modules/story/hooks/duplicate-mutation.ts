import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import type { InfiniteData } from "@tanstack/react-query";
import { useAnalytics } from "@/hooks";
import { slugify } from "@/utils";
import type {
  GroupedStoriesResponse,
  Story,
  GroupStoriesResponse,
} from "@/modules/stories/types";
import { duplicateStoryAction } from "../actions/duplicate-story";
import type { DetailedStory } from "../types";

const updateDetailQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  story: DetailedStory,
) => {
  if (!story.parentId) return;

  queryClient.setQueriesData(
    { queryKey },
    (data: DetailedStory | undefined) => {
      if (data?.id === story.parentId) {
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
      return data;
    },
  );
};

const updateInfiniteQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  story: DetailedStory,
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: InfiniteData<GroupStoriesResponse> | undefined) => {
      if (!data?.pages) return data;

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
          data.pages.reduce(
            (max: number, page) =>
              Math.max(max, ...page.stories.map((s) => s.sequenceId)),
            0,
          ) + 1,
        priority: story.priority,
        startDate: story.startDate || null,
        endDate: story.endDate || null,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        labels: [],
        subStories: [],
      };

      return {
        ...data,
        pages: data.pages.map((page, index) =>
          index === 0
            ? {
                ...page,
                stories: [...page.stories, newStory],
              }
            : page,
        ),
      };
    },
  );
};

const updateGroupedQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  story: DetailedStory,
) => {
  queryClient.setQueriesData(
    { queryKey },
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
              Math.max(max, ...group.stories.map((s) => s.sequenceId)),
            0,
          ) + 1,
        priority: story.priority,
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
};

// Helper function to update list queries (grouped or infinite)
const updateListQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  story: DetailedStory,
) => {
  const queryData = queryClient.getQueryData(queryKey);
  const isInfiniteQuery =
    queryData && typeof queryData === "object" && "pages" in queryData;

  if (isInfiniteQuery) {
    updateInfiniteQuery(queryClient, queryKey, story);
  } else {
    updateGroupedQuery(queryClient, queryKey, story);
  }
};

// Helper function to remove optimistic story on error
const removeOptimisticStory = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
) => {
  const queryData = queryClient.getQueryData(queryKey);
  const isInfiniteQuery =
    queryData && typeof queryData === "object" && "pages" in queryData;

  if (isInfiniteQuery) {
    queryClient.setQueriesData(
      { queryKey },
      (data: InfiniteData<GroupStoriesResponse> | undefined) => {
        if (!data?.pages) return data;
        return {
          ...data,
          pages: data.pages.map((page) => ({
            ...page,
            stories: page.stories.filter((story) => story.id !== "123"),
          })),
        };
      },
    );
  } else {
    queryClient.setQueriesData(
      { queryKey },
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
};

export const useDuplicateStoryMutation = () => {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({ storyId }: { storyId: string; story: DetailedStory }) =>
      duplicateStoryAction(storyId),

    onMutate: ({ story }) => {
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (queryKey.toLowerCase().includes("stories") && query.isActive()) {
          if (queryKey.toLowerCase().includes("detail")) {
            updateDetailQuery(queryClient, query.queryKey, story);
          } else {
            updateListQuery(queryClient, query.queryKey, story);
          }
        }
      });
    },

    onError: (error, variables) => {
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (
          queryKey.toLowerCase().includes("stories") &&
          !queryKey.toLowerCase().includes("detail")
        ) {
          removeOptimisticStory(queryClient, query.queryKey);
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
