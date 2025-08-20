import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { InfiniteData } from "@tanstack/react-query";
import type { DetailedStory } from "@/modules/story/types";
import { storyKeys } from "../constants";
import { bulkDeleteAction } from "../actions/bulk-delete-stories";
import type { GroupedStoriesResponse, GroupStoriesResponse } from "../types";
import { useBulkRestoreStoryMutation } from "./restore-mutation";

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
          const deletedCount = group.stories.filter((story) =>
            storyIds.includes(story.id),
          ).length;

          return {
            ...group,
            stories: group.stories.filter(
              (story) => !storyIds.includes(story.id),
            ),
            totalCount: Math.max(0, group.totalCount - deletedCount),
            loadedCount: Math.max(0, group.loadedCount - deletedCount),
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

export const useBulkDeleteStoryMutation = () => {
  const queryClient = useQueryClient();
  const { mutateAsync } = useBulkRestoreStoryMutation();

  const mutation = useMutation({
    mutationFn: bulkDeleteAction,

    onMutate: ({ storyIds }) => {
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

    onError: (error, payload) => {
      queryClient.invalidateQueries({ queryKey: storyKeys.all });

      toast.error("Failed to delete stories", {
        description:
          error.message || "An error occurred while deleting the stories",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(payload);
          },
        },
      });
    },

    onSuccess: (res, { storyIds, hardDelete }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });

      toast.info("You want to undo this action?", {
        description: `${storyIds.length} stor${
          storyIds.length === 1 ? "y" : "ies"
        } deleted`,
        cancel: hardDelete
          ? undefined
          : {
              label: "Undo",
              onClick: () => {
                mutateAsync(storyIds);
              },
            },
      });
    },
  });

  return mutation;
};
