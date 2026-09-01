import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import type { InfiniteData } from "@tanstack/react-query";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import { objectiveKeys } from "@/shared/objectives/keys";
import { storyKeys } from "../constants";
import type {
  DetailedStory,
  GroupedStoriesResponse,
  GroupStoriesResponse,
  Story,
} from "../types";
import { bulkUpdateAction } from "../actions/bulk-update-stories";
import {
  assertBulkStoryUpdateSucceeded,
  BulkStoryUpdateFailure,
} from "./bulk-update-result";

type BulkUpdateVariables = {
  storyIds: string[];
  payload: Partial<DetailedStory>;
};

type BulkUpdateContext = {
  previousQueryStates: Map<string, unknown>;
};

const restorePreviousQueryStates = (
  queryClient: ReturnType<typeof useQueryClient>,
  context?: BulkUpdateContext,
) => {
  context?.previousQueryStates.forEach((data, queryKey) => {
    try {
      queryClient.setQueryData(JSON.parse(queryKey), data);
    } catch {
      // Query keys are JSON-serializable in this cache. Ignore stale entries
      // if a custom key violates that contract instead of masking the mutation.
    }
  });
};

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
        pages: data.pages.map((page) => {
          if (!Array.isArray(page.stories)) return page;
          return {
            ...page,
            stories: page.stories.map((story) =>
              storyIds.includes(story.id) ? { ...story, ...payload } : story,
            ),
          };
        }),
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
      if (!data || !Array.isArray(data.groups)) return data;
      return {
        ...data,
        groups: data.groups.map((group) => {
          if (!Array.isArray(group.stories)) return group;
          return {
            ...group,
            stories: group.stories.map((story) =>
              storyIds.includes(story.id) ? { ...story, ...payload } : story,
            ),
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
  payload: Partial<DetailedStory>,
) => {
  const queryData = queryClient.getQueryData(queryKey);
  if (Array.isArray(queryData)) {
    queryClient.setQueryData<Story[]>(queryKey, (data) => {
      if (!Array.isArray(data)) return data;
      return data.map((story) =>
        storyIds.includes(story.id) ? { ...story, ...payload } : story,
      );
    });
    return;
  }

  const isInfiniteQuery = Boolean(
    queryData && typeof queryData === "object" && "pages" in queryData,
  );

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
  const { getTermDisplay } = useTerminology();

  const mutation = useMutation({
    mutationFn: async ({ storyIds, payload }: BulkUpdateVariables) => {
      const response = await bulkUpdateAction(
        { storyIds, updates: payload },
        workspaceSlug,
      );

      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      if (!response.data) {
        throw new Error("The bulk update returned no result");
      }

      return assertBulkStoryUpdateSucceeded(response.data);
    },

    onMutate: ({ storyIds, payload }) => {
      const previousQueryStates = new Map<string, unknown>();
      const queryCache = queryClient.getQueryCache();
      const queries = queryCache.getAll();

      queries.forEach((query) => {
        const queryKey = JSON.stringify(query.queryKey);
        if (query.isActive() && queryKey.toLowerCase().includes("stories")) {
          queryClient.cancelQueries({ queryKey: query.queryKey });

          const previousData = queryClient.getQueryData(query.queryKey);
          if (!previousQueryStates.has(queryKey)) {
            previousQueryStates.set(queryKey, previousData);
          }

          if (queryKey.toLowerCase().includes("detail")) {
            updateDetailQuery(queryClient, query.queryKey, storyIds, payload);
          } else {
            updateListQuery(queryClient, query.queryKey, storyIds, payload);
          }
        }
      });

      if (storyId) {
        const parentStoryKey = storyKeys.detail(workspaceSlug, storyId);
        const serializedParentStoryKey = JSON.stringify(parentStoryKey);
        const parentStory =
          queryClient.getQueryData<DetailedStory>(parentStoryKey);
        if (parentStory && !previousQueryStates.has(serializedParentStoryKey)) {
          const previousParentData = queryClient.getQueryData(parentStoryKey);
          previousQueryStates.set(serializedParentStoryKey, previousParentData);

          updateDetailQuery(queryClient, parentStoryKey, storyIds, payload);
        }
      }

      return { previousQueryStates };
    },

    onError: (error, variables, context) => {
      restorePreviousQueryStates(queryClient, context);

      queryClient.invalidateQueries({ queryKey: storyKeys.all(workspaceSlug) });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.list(workspaceSlug),
      });

      const itemFailure =
        error instanceof BulkStoryUpdateFailure ? error : null;
      const retryStoryIds =
        itemFailure &&
        itemFailure.failedStoryIds.length === itemFailure.failedCount
          ? itemFailure.failedStoryIds
          : variables.storyIds;
      const failureTitle = itemFailure
        ? `Failed to update ${itemFailure.failedCount} of ${itemFailure.totalCount} ${getTermDisplay(
            "storyTerm",
            {
              variant: itemFailure.totalCount === 1 ? "singular" : "plural",
            },
          )}`
        : `Failed to update ${getTermDisplay("storyTerm", { variant: "plural" })}`;

      toast.error(failureTitle, {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate({
              ...variables,
              storyIds: retryStoryIds,
            });
          },
        },
      });
    },

    onSuccess: (_result, { storyIds, payload }) => {
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
