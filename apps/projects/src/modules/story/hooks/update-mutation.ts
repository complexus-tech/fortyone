import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { InfiniteData } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type {
  GroupedStoriesResponse,
  GroupStoriesResponse,
  GroupStoryParams,
} from "@/modules/stories/types";
import type { SearchResponse } from "@/modules/search/types";
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
  // Helper to update a single infinite query in place
  const patchInfiniteQuery = (
    key: readonly unknown[],
    updater: (
      data: InfiniteData<GroupStoriesResponse> | undefined,
    ) => InfiniteData<GroupStoriesResponse> | undefined,
  ) => {
    queryClient.setQueryData<InfiniteData<GroupStoriesResponse>>(key, updater);
  };

  // Update the current query (remove or edit)
  patchInfiniteQuery(queryKey, (data) => {
    if (!data?.pages) return data;

    // Find groupKey from queryKey structure ["stories","group",groupKey,{...params}]
    const currentGroupKey =
      (Array.isArray(queryKey) && queryKey.length > 2
        ? (queryKey[2] as string)
        : undefined) ?? "";

    // Determine grouping field of this list from params object
    const paramsObject =
      Array.isArray(queryKey) && queryKey.length > 3
        ? (queryKey[3] as Partial<GroupStoryParams>)
        : {};

    const groupBy = paramsObject.groupBy ?? "none";

    // Compute the story's new key after update
    let newKey: string | undefined;
    switch (groupBy) {
      case "status":
        newKey = payload.statusId;
        break;
      case "priority":
        newKey = payload.priority as string | undefined;
        break;
      case "assignee":
        newKey = payload.assigneeId ?? undefined;
        break;
      default:
        break;
    }

    // If story stays in the same group, just patch the fields
    if (!newKey || newKey === currentGroupKey) {
      return {
        ...data,
        pages: data.pages.map((page) => ({
          ...page,
          stories: page.stories.map((story) =>
            story.id === storyId ? { ...story, ...payload } : story,
          ),
        })),
      };
    }

    // Otherwise, remove story from this query
    return {
      ...data,
      pages: data.pages.map((page) => ({
        ...page,
        stories: page.stories.filter((s) => s.id !== storyId),
      })),
    };
  });

  // If story moved to a different group, insert it in the target group's cache (if present)
  const currentParams =
    Array.isArray(queryKey) && queryKey.length > 3
      ? (queryKey[3] as Partial<GroupStoryParams>)
      : ({} as Partial<GroupStoryParams>);

  const groupByField = currentParams.groupBy ?? "none";

  let targetKeyValue: string | undefined;
  switch (groupByField) {
    case "status":
      targetKeyValue = payload.statusId;
      break;
    case "priority":
      targetKeyValue = payload.priority as string | undefined;
      break;
    case "assignee":
      targetKeyValue = payload.assigneeId ?? undefined;
      break;
    default:
      break;
  }

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
    if (!data?.pages) return data;

    const firstPage = data.pages[0];

    // Avoid duplicates
    const exists = firstPage.stories.some((s) => s.id === storyId);

    if (exists) return data;

    // Prepend story to first page
    const updatedStory = {
      ...(firstPage.stories[0] ?? {}),
      id: storyId,
      ...payload,
    } as (typeof firstPage.stories)[number];

    return {
      ...data,
      pages: [
        {
          ...firstPage,
          stories: [updatedStory, ...firstPage.stories],
        },
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

    // Determine the grouping key used by this response
    const groupBy = data.meta.groupBy;

    let movedStory: DetailedStory | null = null;

    // 1. Remove the story from its current group and collect it
    const groupsWithoutStory = data.groups.map((group) => {
      const remainingStories = group.stories.filter((story) => {
        const isMatch = story.id === storyId;
        if (isMatch) {
          movedStory = { ...story, ...payload } as DetailedStory;
        }
        return !isMatch;
      });

      return {
        ...group,
        stories: remainingStories,
      };
    });

    // 2. Figure out the target group key based on the payload
    let targetKey: string | undefined;
    switch (groupBy) {
      case "status":
        targetKey = payload.statusId;
        break;
      case "priority":
        targetKey = (payload.priority as string | undefined) ?? undefined;
        break;
      case "assignee":
        targetKey = payload.assigneeId ?? undefined;
        break;
      default:
        break;
    }

    // 3. Insert the story into the target group (front of the list)
    const finalGroups = groupsWithoutStory.map((group) => {
      if (targetKey === group.key) {
        return {
          ...group,
          stories: [movedStory!, ...group.stories],
        };
      }
      return group;
    });

    return {
      ...data,
      groups: finalGroups,
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
