import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { InfiniteData } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
  GroupStoryParams,
  Story,
} from "@/modules/stories/types";
import type { SearchResponse } from "@/modules/search/types";
import {
  computeTargetKey,
  moveStoryBetweenGroups,
  parseGroupQueryKey,
} from "@/modules/stories/utils/optimistic";
import type { DetailedStory } from "../types";
import { updateStoryAction } from "../actions/update-story";

export const useUpdateStoryMutation = () => {
  const queryClient = useQueryClient();
  const { analytics } = useAnalytics();

  const mutation = useMutation({
    mutationFn: ({
      storyId,
      payload,
    }: {
      storyId: string;
      payload: Partial<DetailedStory>;
    }) => updateStoryAction(storyId, payload),

    onMutate: ({ storyId, payload }) => {
      queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId),
      );

      const activeQueries = queryClient.getQueryCache().getAll();

      activeQueries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.cancelQueries({ queryKey: query.queryKey });
          if (queryKey.toLowerCase().includes("detail")) {
            updateDetailQuery(queryClient, query.queryKey, storyId, payload);
          } else {
            updateListQuery(queryClient, query.queryKey, storyId, payload);
          }
        }
      });

      updateSearchResults(queryClient, storyId, payload);

      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          ...payload,
        });
        return { previousStory };
      }
    },

    onError: (error, variables, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(variables.storyId),
          context.previousStory,
        );
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      toast.error("Failed to update story", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: (res, { storyId, payload }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      analytics.track("story_updated", {
        storyId,
        ...payload,
      });

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
    },
  });

  return mutation;
};

const updateDetailQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  const parentStory = queryClient.getQueryData<DetailedStory>(queryKey);
  if (parentStory?.subStories) {
    queryClient.setQueryData<DetailedStory>(queryKey, {
      ...parentStory,
      subStories: parentStory.subStories.map((subStory) =>
        subStory.id === storyId ? { ...subStory, ...payload } : subStory,
      ),
    });
  }
};

const updateListQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  const queryData = queryClient.getQueryData(queryKey);
  const isInfiniteQuery =
    queryData && typeof queryData === "object" && "pages" in queryData;

  if (isInfiniteQuery) {
    updateInfiniteQuery(queryClient, queryKey, storyId, payload);
  } else {
    updateGroupedQuery(queryClient, queryKey, storyId, payload);
  }
};

const updateInfiniteQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  // Store the real story object once we encounter it
  let movedStory: Story | undefined;

  const patchInfiniteQuery = (
    key: readonly unknown[],
    updater: (
      data: InfiniteData<GroupStoriesResponse> | undefined,
    ) => InfiniteData<GroupStoriesResponse> | undefined,
  ) =>
    queryClient.setQueryData<InfiniteData<GroupStoriesResponse>>(key, updater);

  // 1️⃣ patch current query (remove/update)
  patchInfiniteQuery(queryKey, (data) => {
    if (!data?.pages) return data;
    const { groupKey, params } = parseGroupQueryKey(queryKey);
    const target = computeTargetKey(params.groupBy ?? "none", payload);

    if (!target || target === groupKey) {
      return {
        ...data,
        pages: data.pages.map((p) => ({
          ...p,
          stories: p.stories.map((s) => {
            if (s.id === storyId) {
              movedStory = { ...s, ...payload };
              return movedStory;
            }
            return s;
          }),
        })),
      };
    }

    // moved: filter out and capture
    return {
      ...data,
      pages: data.pages.map((p) => ({
        ...p,
        stories: p.stories.filter((s) => {
          if (s.id === storyId) {
            movedStory = { ...s, ...payload };
            return false;
          }
          return true;
        }),
      })),
    };
  });

  const { params: currentParams } = parseGroupQueryKey(queryKey);
  const targetKeyValue = computeTargetKey(
    currentParams.groupBy ?? "none",
    payload,
  );

  if (!targetKeyValue) return;

  const targetParams: Partial<GroupStoryParams> = {
    ...currentParams,
    groupKey: targetKeyValue,
  };

  const targetQueryKey = [
    "stories",
    "group",
    targetKeyValue,
    targetParams,
  ] as const;

  patchInfiniteQuery(targetQueryKey, (data) => {
    if (!movedStory) return data;

    // If no pages yet, create minimal first page
    if (!data?.pages || data.pages.length === 0) {
      return {
        pages: [
          {
            groupKey: targetKeyValue,
            stories: [movedStory],
            pagination: {
              page: 1,
              pageSize: 1,
              hasMore: false,
              nextPage: 2,
            },
            filters: {},
          },
        ],
        pageParams: [1],
      } as InfiniteData<GroupStoriesResponse>;
    }

    const firstPage = data.pages[0];

    // Avoid duplicates
    if (firstPage.stories.some((s) => s.id === storyId)) return data;

    return {
      ...data,
      pages: [
        { ...firstPage, stories: [movedStory, ...firstPage.stories] },
        ...data.pages.slice(1),
      ],
    };
  });
};

const updateGroupedQuery = (
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  queryClient.setQueryData<GroupedStoriesResponse>(queryKey, (data) => {
    if (!data) return data;

    const target = computeTargetKey(data.meta.groupBy, payload);
    return {
      ...data,
      groups: moveStoryBetweenGroups(data.groups, storyId, target, payload),
    };
  });
};

const updateSearchResults = (
  queryClient: ReturnType<typeof useQueryClient>,
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  queryClient
    .getQueriesData<SearchResponse>({ queryKey: ["search"] })
    .forEach(([queryKey, data]) => {
      if (data?.stories) {
        queryClient.setQueryData<SearchResponse>(queryKey, {
          ...data,
          stories: data.stories.map((story) =>
            story.id === storyId ? { ...story, ...payload } : story,
          ),
        });
      }
    });
};
