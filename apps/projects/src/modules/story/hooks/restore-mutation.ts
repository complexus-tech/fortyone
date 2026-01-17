import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { InfiniteData } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
} from "@/modules/stories/types";
import type { DetailedStory } from "../types";
import { restoreStoryAction } from "../actions/restore-story";

const updateInfiniteQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  story: DetailedStory,
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: InfiniteData<GroupStoriesResponse> | undefined) => {
      if (!data?.pages) return data;
      return {
        ...data,
        pages: data.pages.map((page, index) =>
          index === 0
            ? {
                ...page,
                stories: [...page.stories, story],
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
      return {
        ...data,
        groups: data.groups.map((group, index) =>
          index === 0
            ? {
                ...group,
                stories: [...group.stories, story],
                totalCount: group.totalCount + 1,
                loadedCount: group.loadedCount + 1,
              }
            : group,
        ),
      };
    },
  );
};

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

export const useRestoreStoryMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: restoreStoryAction,

    onMutate: (storyId) => {
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(workspaceSlug, storyId),
      );
      if (previousStory) {
        const restoredStory = {
          ...previousStory,
          deletedAt: null,
        };
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, storyId),
          restoredStory,
        );

        const queryCache = queryClient.getQueryCache();
        const queries = queryCache.getAll();

        queries.forEach((query) => {
          const queryKey = JSON.stringify(query.queryKey);
          if (
            queryKey.toLowerCase().includes("stories") &&
            !queryKey.toLowerCase().includes("detail") &&
            query.isActive()
          ) {
            updateListQuery(queryClient, query.queryKey, restoredStory);
          }
        });
      }

      return { previousStory };
    },

    onError: (error, storyId, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, storyId),
          context.previousStory,
        );
      }

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

      toast.error("Failed to restore story", {
        description:
          error.message || "An error occurred while restoring the story",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(storyId);
          },
        },
      });
    },

    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
    },
  });

  return mutation;
};
