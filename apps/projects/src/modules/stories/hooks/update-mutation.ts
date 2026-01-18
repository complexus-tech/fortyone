import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import type { InfiniteData } from "@tanstack/react-query";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import type { DetailedStory } from "@/modules/story/types";
import { storyKeys } from "../constants";
import type { GroupedStoriesResponse, GroupStoriesResponse } from "../types";
import { bulkUpdateAction } from "../actions/bulk-update-stories";

const updateDetailQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyIds: string[],
  payload: Partial<DetailedStory>,
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: DetailedStory | undefined) => {
      if (data?.subStories) {
        return {
          ...data,
          subStories: data.subStories.map((story) =>
            storyIds.includes(story.id) ? { ...story, ...payload } : story,
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
  payload: Partial<DetailedStory>,
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: InfiniteData<GroupStoriesResponse> | undefined) => {
      if (!data?.pages) return data;
      return {
        ...data,
        pages: data.pages.map((page) => ({
          ...page,
          stories: page.stories.map((story) =>
            storyIds.includes(story.id) ? { ...story, ...payload } : story,
          ),
        })),
      };
    },
  );
};

const updateGroupedQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyIds: string[],
  payload: Partial<DetailedStory>,
) => {
  queryClient.setQueriesData(
    { queryKey },
    (data: GroupedStoriesResponse | undefined) => {
      if (!data) return data;
      return {
        ...data,
        groups: data.groups.map((group) => ({
          ...group,
          stories: group.stories.map((story) =>
            storyIds.includes(story.id) ? { ...story, ...payload } : story,
          ),
        })),
      };
    },
  );
};

const updateListQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyIds: string[],
  payload: Partial<DetailedStory>,
) => {
  const queryData = queryClient.getQueryData(queryKey);
  const isInfiniteQuery =
    queryData && typeof queryData === "object" && "pages" in queryData;

  if (isInfiniteQuery) {
    updateInfiniteQuery(queryClient, queryKey, storyIds, payload);
  } else {
    updateGroupedQuery(queryClient, queryKey, storyIds, payload);
  }
};

export const useBulkUpdateStoriesMutation = () => {
  const queryClient = useQueryClient();
  const { storyId } = useParams<{ storyId?: string }>();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({
      storyIds,
      payload,
    }: {
      storyIds: string[];
      payload: Partial<DetailedStory>;
    }) => bulkUpdateAction({ storyIds, updates: payload }, workspaceSlug),

    onMutate: ({ storyIds, payload }) => {
      const previousQueryStates = new Map<string, unknown>();
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.cancelQueries({ queryKey: query.queryKey });

          const previousData = queryClient.getQueryData(query.queryKey);
          previousQueryStates.set(queryKey, previousData);

          if (queryKey.toLowerCase().includes("detail")) {
            updateDetailQuery(queryClient, query.queryKey, storyIds, payload);
          } else {
            updateListQuery(queryClient, query.queryKey, storyIds, payload);
          }
        }
      });

      if (storyId) {
        const parentStory = queryClient.getQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, storyId),
        );
        if (parentStory) {
          const previousParentData = queryClient.getQueryData(
            storyKeys.detail(workspaceSlug, storyId),
          );
          previousQueryStates.set(
            JSON.stringify(storyKeys.detail(workspaceSlug, storyId)),
            previousParentData,
          );

          updateDetailQuery(
            queryClient,
            storyKeys.detail(workspaceSlug, storyId),
            storyIds,
            payload,
          );
        }
      }

      return { previousQueryStates };
    },

    onError: (error, variables, context) => {
      if (context?.previousQueryStates) {
        context.previousQueryStates.forEach((data, queryKey) => {
          try {
            const parsedQueryKey = JSON.parse(queryKey);
            queryClient.setQueryData(parsedQueryKey, data);
          } catch {
            // Skip invalid query keys
          }
        });
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });

      toast.error("Failed to update stories", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (res, { storyIds, payload }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      analytics.track("stories_bulk_updated", {
        storyIds,
        count: storyIds.length,
        ...payload,
      });

      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
    },
  });

  return mutation;
};
