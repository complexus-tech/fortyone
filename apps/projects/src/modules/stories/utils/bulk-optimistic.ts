import type {
  InfiniteData,
  QueryClient,
  QueryKey,
} from "@tanstack/react-query";
import { hashKey } from "@tanstack/react-query";
import { statusKeys } from "@/constants/keys";
import type { State } from "@/types/states";
import { storyKeys } from "../constants";
import type {
  DetailedStory,
  GroupedStoriesResponse,
  GroupStoriesResponse,
  GroupStoryParams,
  Story,
} from "../types";
import {
  computeTargetKey,
  moveStoryBetweenGroups,
  parseGroupQueryKey,
  patchSelectedStory,
  updateStoryInGroups,
} from "./optimistic";

export type BulkStoryCacheSnapshot = Map<QueryKey, unknown>;

const isRecord = (data: unknown): data is Record<string, unknown> =>
  data !== null && typeof data === "object";

const isInfiniteStories = (
  data: unknown,
): data is InfiniteData<GroupStoriesResponse> =>
  isRecord(data) && Array.isArray(data.pages);

const isGroupedStories = (data: unknown): data is GroupedStoriesResponse =>
  isRecord(data) &&
  isRecord(data.meta) &&
  typeof data.meta.groupBy === "string" &&
  Array.isArray(data.groups) &&
  data.groups.every(
    (group: unknown) => isRecord(group) && Array.isArray(group.stories),
  );

const scopeKey = (params: Record<string, unknown>) =>
  hashKey([
    Object.fromEntries(
      Object.entries(params).filter(
        ([key, value]) =>
          !["groupKey", "page", "pageSize", "storiesPerGroup"].includes(key) &&
          value !== undefined &&
          value !== null &&
          (!Array.isArray(value) || value.length > 0),
      ),
    ),
  ]);

const targetMatchesFilters = (
  queryClient: QueryClient,
  workspaceSlug: string,
  params: Partial<GroupStoryParams>,
  target: string,
) => {
  if (params.groupBy === "priority") {
    return (
      (!params.priorities?.length || params.priorities.includes(target)) &&
      !params.excludedPriorities?.includes(target)
    );
  }
  if (params.groupBy === "assignee") {
    return (
      (!params.assigneeIds?.length || params.assigneeIds.includes(target)) &&
      !params.excludedAssigneeIds?.includes(target) &&
      (!params.hasNoAssignee || target === "null") &&
      (!params.hasAssignee || target !== "null")
    );
  }
  if (params.groupBy !== "status") return true;
  if (
    (params.statusIds?.length && !params.statusIds.includes(target)) ||
    params.excludedStatusIds?.includes(target)
  ) {
    return false;
  }
  if (!params.categories?.length) return true;

  const status = queryClient
    .getQueriesData<State[]>({ queryKey: statusKeys.lists(workspaceSlug) })
    .flatMap(([, statuses]) => statuses ?? [])
    .find(({ id }) => id === target);
  return Boolean(status && params.categories.includes(status.category));
};

export const restoreBulkStoryUpdate = (
  queryClient: QueryClient,
  snapshot: BulkStoryCacheSnapshot,
) => {
  snapshot.forEach((data, queryKey) => {
    if (data === undefined) {
      queryClient.removeQueries({ queryKey, exact: true });
    } else {
      queryClient.setQueryData(queryKey, data);
    }
  });
};

export const applyBulkStoryUpdate = ({
  queryClient,
  workspaceSlug,
  storyIds,
  payload,
  parentStoryId,
}: {
  queryClient: QueryClient;
  workspaceSlug: string;
  storyIds: string[];
  payload: Partial<DetailedStory>;
  parentStoryId?: string;
}): BulkStoryCacheSnapshot => {
  const selectedIds = new Set(storyIds);
  const queries = queryClient
    .getQueryCache()
    .findAll({ queryKey: storyKeys.all(workspaceSlug) });
  // A source can write an inactive destination before its own turn in the loop.
  // Capture every existing value before writing any optimistic changes.
  const originalData = new Map(
    queries.map((query) => [hashKey(query.queryKey), query.state.data]),
  );
  const originalKeys = new Map(
    queries.map((query) => [hashKey(query.queryKey), query.queryKey]),
  );
  const previousQueryStates: BulkStoryCacheSnapshot = new Map();
  const activeQueries = queries.filter((query) => query.isActive());
  const activeInfiniteQueries = activeQueries.filter(
    (query) =>
      query.queryKey[2] === "group" && isInfiniteStories(query.state.data),
  );
  const patchStories = (stories: Story[]) =>
    stories.map((story) => patchSelectedStory(story, selectedIds, payload));
  const write = (queryKey: QueryKey, data: unknown) => {
    if (data === queryClient.getQueryData(queryKey)) return;
    const keyHash = hashKey(queryKey);
    const snapshotKey = originalKeys.get(keyHash) ?? queryKey;
    originalKeys.set(keyHash, snapshotKey);
    if (!previousQueryStates.has(snapshotKey)) {
      previousQueryStates.set(snapshotKey, originalData.get(keyHash));
    }
    queryClient.setQueryData(queryKey, data);
  };

  const updateInfinite = (
    queryKey: QueryKey,
    data: InfiniteData<GroupStoriesResponse>,
  ) => {
    const isGroupQuery = queryKey[2] === "group";
    const { groupKey, params } = parseGroupQueryKey(queryKey);
    const target = isGroupQuery
      ? computeTargetKey(params.groupBy ?? "none", payload)
      : undefined;
    const shouldMove = Boolean(target && target !== groupKey);
    const moved = new Map<string, Story>();
    const pages = data.pages.map((page) => {
      if (!Array.isArray(page.stories)) return page;
      const stories = patchStories(page.stories).filter((story) => {
        if (!shouldMove || !selectedIds.has(story.id)) return true;
        moved.set(story.id, story);
        return false;
      });
      return { ...page, stories };
    });
    write(queryKey, { ...data, pages });
    if (
      !target ||
      moved.size === 0 ||
      !targetMatchesFilters(queryClient, workspaceSlug, params, target)
    ) {
      return;
    }

    const targetQueryKey = storyKeys.groupStories(workspaceSlug, target, {
      ...params,
      groupKey: target,
    });
    const destination =
      queryClient.getQueryData<InfiniteData<GroupStoriesResponse>>(
        targetQueryKey,
      );
    const destinationIds = new Set(
      destination?.pages.flatMap((page) => page.stories.map(({ id }) => id)),
    );
    const arrivingStories = Array.from(moved.values()).filter(
      ({ id }) => !destinationIds.has(id),
    );
    if (destination?.pages.length) {
      write(targetQueryKey, {
        ...destination,
        pages: destination.pages.map((page, index) => ({
          ...page,
          stories: [
            ...(index === 0 ? arrivingStories : []),
            ...patchStories(page.stories),
          ],
        })),
      });
    } else {
      write(targetQueryKey, {
        pages: [
          {
            ...data.pages[0],
            groupKey: target,
            stories: arrivingStories,
            pagination: {
              page: 1,
              pageSize: arrivingStories.length,
              hasMore: false,
              nextPage: 2,
            },
          },
        ],
        pageParams: [1],
      });
    }
  };

  const updateGrouped = (queryKey: QueryKey, data: GroupedStoriesResponse) => {
    const groupBy = data.meta.groupBy;
    const target = computeTargetKey(groupBy, payload);
    const groupedParams = {
      ...data.meta.filters,
      groupBy,
      orderBy: data.meta.orderBy,
      orderDirection: data.meta.orderDirection,
    };
    const matchingPages = activeInfiniteQueries.filter(({ queryKey: key }) => {
      const { params } = parseGroupQueryKey(key);
      return scopeKey(params) === scopeKey(groupedParams);
    });
    const canEnterTarget =
      !target ||
      targetMatchesFilters(
        queryClient,
        workspaceSlug,
        groupedParams as Partial<GroupStoryParams>,
        target,
      );
    let groups = data.groups.map((group) => {
      if (!target || group.key === target) return group;
      const missingStories = new Map<string, Story>();
      const existingIds = new Set(group.stories.map(({ id }) => id));
      for (const query of matchingPages) {
        if (parseGroupQueryKey(query.queryKey).groupKey !== group.key) continue;
        const original = originalData.get(hashKey(query.queryKey));
        if (!isInfiniteStories(original)) continue;
        for (const page of original.pages) {
          for (const story of page.stories) {
            if (selectedIds.has(story.id) && !existingIds.has(story.id)) {
              missingStories.set(story.id, story);
            }
          }
        }
      }
      if (missingStories.size === 0) return group;
      // These stories were loaded after the grouped response's first page.
      // Include them for the move without counting them twice in totalCount.
      return {
        ...group,
        loadedCount: group.loadedCount + missingStories.size,
        stories: [...group.stories, ...missingStories.values()],
      };
    });
    for (const storyId of selectedIds) {
      groups = canEnterTarget
        ? updateStoryInGroups(groups, storyId, groupBy, payload)
        : moveStoryBetweenGroups(groups, storyId, undefined, payload);
    }
    write(queryKey, { ...data, groups });
  };

  for (const query of queries) {
    const { queryKey } = query;
    const isParentDetail =
      parentStoryId &&
      queryKey[2] === "detail" &&
      queryKey[3] === parentStoryId;
    if (!query.isActive() && !isParentDetail) continue;
    const data = queryClient.getQueryData(queryKey);
    if (queryKey[2] === "detail") {
      if (isRecord(data) && typeof data.id === "string") {
        write(
          queryKey,
          patchSelectedStory(data as DetailedStory, selectedIds, payload),
        );
      }
    } else if (Array.isArray(data)) {
      write(queryKey, patchStories(data as Story[]));
    } else if (isInfiniteStories(data)) {
      updateInfinite(queryKey, data);
    } else if (isGroupedStories(data)) {
      updateGrouped(queryKey, data);
    }
  }
  return previousQueryStates;
};
