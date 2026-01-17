import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { InfiniteData } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks";
import type { DetailedStory } from "@/modules/story/types";
import { storyKeys } from "../constants";
import { bulkArchiveAction } from "../actions/bulk-archive-stories";
import type { GroupedStoriesResponse, GroupStoriesResponse } from "../types";

const updateDetailQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyIds: string[],
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: DetailedStory | undefined) => {
      if (data?.subStories) {
        return {
          ...data,
          subStories: data.subStories.filter(
            (story) => !storyIds.includes(story.id),
          ),
        };
      }
      return data;
    },
  );
};

const updateInfiniteQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyIds: string[],
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: InfiniteData<GroupStoriesResponse> | undefined) => {
      if (!data?.pages) return data;
      return {
        ...data,
        pages: data.pages.map((page) => ({
          ...page,
          stories: page.stories.filter((story) => !storyIds.includes(story.id)),
        })),
      };
    },
  );
};

const updateGroupedQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyIds: string[],
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: GroupedStoriesResponse | undefined) => {
      if (!data) return data;
      return {
        ...data,
        groups: data.groups.map((group) => {
          const archivedCount = group.stories.filter((story) =>
            storyIds.includes(story.id),
          ).length;

          return {
            ...group,
            stories: group.stories.filter(
              (story) => !storyIds.includes(story.id),
            ),
            totalCount: Math.max(0, group.totalCount - archivedCount),
            loadedCount: Math.max(0, group.loadedCount - archivedCount),
          };
        }),
      };
    },
  );
};

const updateListQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyIds: string[],
) => {
  const queryData = queryClient.getQueryData(queryKey);
  const isInfiniteQuery =
    queryData && typeof queryData === "object" && "pages" in queryData;

  if (isInfiniteQuery) {
    updateInfiniteQuery(queryClient, queryKey, storyIds);
  } else {
    updateGroupedQuery(queryClient, queryKey, storyIds);
  }
};

export const useBulkArchiveStoryMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (storyIds: string[]) =>
      bulkArchiveAction(storyIds, workspaceSlug),

    onMutate: (storyIds) => {
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (queryKey.toLowerCase().includes("stories") && query.isActive()) {
          if (queryKey.toLowerCase().includes("detail")) {
            updateDetailQuery(queryClient, query.queryKey, storyIds);
          } else {
            updateListQuery(queryClient, query.queryKey, storyIds);
          }
        }
      });

      return storyIds;
    },

    onError: (error, storyIds) => {
      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });

      toast.error("Failed to archive stories", {
        description:
          error.message || "An error occurred while archiving the stories",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(storyIds);
          },
        },
      });
    },

    onSuccess: (res, storyIds) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });

      toast.success("Stories archived successfully", {
        description: `${storyIds.length} stor${
          storyIds.length === 1 ? "y" : "ies"
        } archived`,
      });
    },
  });

  return mutation;
};
