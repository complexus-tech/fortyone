import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import type { InfiniteData } from "@tanstack/react-query";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import { objectiveKeys } from "@/shared/objectives/keys";
import type { WipStorySnapshot } from "@/shared/story/wip-capacity";
import {
  affectsWipCapacity,
  captureWipStorySnapshots,
  isWipCapacityQuery,
  refreshWipCapacity,
} from "@/shared/story/wip-capacity";
import { storyKeys } from "../constants";
import type {
  DetailedStory,
  GroupedStoriesResponse,
  GroupStoriesResponse,
  Story,
} from "../types";
import { bulkUpdateAction } from "../actions/bulk-update-stories";
import { loadWipCapacity } from "../public/wip-capacity";
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
  wipSnapshots: WipStorySnapshot[];
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
      if (data && storyIds.includes(data.id)) {
        return { ...data, ...payload };
      }
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
        { storyIds: Array.from(new Set(storyIds)), updates: payload },
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

    onMutate: async ({ storyIds, payload }) => {
      const wipSnapshots = await captureWipStorySnapshots(
        queryClient,
        workspaceSlug,
        storyIds,
        payload,
      );
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

      return { previousQueryStates, wipSnapshots };
    },

    onError: (error, variables, context) => {
      restorePreviousQueryStates(queryClient, context);

      const itemFailure =
        error instanceof BulkStoryUpdateFailure ? error : null;
      const hasPersistedChanges = Boolean(
        itemFailure?.successfulStoryIds.length,
      );
      queryClient.invalidateQueries({
        queryKey: storyKeys.all(workspaceSlug),
        predicate: (query) =>
          !hasPersistedChanges || !isWipCapacityQuery(query, workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.list(workspaceSlug),
      });

      if (hasPersistedChanges && affectsWipCapacity(variables.payload)) {
        void refreshWipCapacity({
          queryClient,
          workspaceSlug,
          loadCapacity: loadWipCapacity,
          snapshots: context?.wipSnapshots,
          payload: variables.payload,
          successfulStoryIds: itemFailure?.successfulStoryIds,
        });
      }
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

    onSuccess: (_result, { storyIds, payload }, context) => {
      const uniqueStoryIds = Array.from(new Set(storyIds));
      analytics.track("stories_bulk_updated", {
        storyIds: uniqueStoryIds,
        count: uniqueStoryIds.length,
        ...payload,
      });

      queryClient.invalidateQueries({
        queryKey: storyKeys.all(workspaceSlug),
        predicate: (query) =>
          !affectsWipCapacity(payload) ||
          !isWipCapacityQuery(query, workspaceSlug),
      });
      if (affectsWipCapacity(payload)) {
        void refreshWipCapacity({
          queryClient,
          workspaceSlug,
          loadCapacity: loadWipCapacity,
          snapshots: context.wipSnapshots,
          payload,
          successfulStoryIds: uniqueStoryIds,
        });
      }
    },
  });

  return mutation;
};
