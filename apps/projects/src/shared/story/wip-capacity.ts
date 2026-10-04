import type { Query, QueryClient } from "@tanstack/react-query";
import { hashKey } from "@tanstack/react-query";
import type { ExternalToast } from "sonner";
import { toast } from "sonner";
import posthog from "posthog-js";
import { statusKeys } from "@/constants/keys";
import { getStatuses } from "@/lib/queries/states/get-states";
import type { State } from "@/types/states";
import { getStory } from "./queries/get-story";
import { storyKeys } from "./cache-keys";
import type { StoryUpdate } from "./types";

export type WipCapacityParams = { groupBy: "status"; teamIds?: string[] };
export type WipCapacityResult = {
  groups: { key: string; totalCount: number }[];
};
export type LoadWipCapacity = (
  workspaceSlug: string,
  params: WipCapacityParams,
) => Promise<WipCapacityResult>;

export type WipStorySnapshot = {
  id: string;
  statusId: string;
  teamId?: string;
  parentId: string | null;
  archivedAt: string | null;
  deletedAt: string | null;
};

type CapacityRequest = {
  token: symbol;
  promise: Promise<WipCapacityResult>;
};

type CapacityCoordinator = {
  requests: Map<string, CapacityRequest>;
  notices: Map<string, number>;
};

const coordinators = new WeakMap<QueryClient, CapacityCoordinator>();

const getCoordinator = (queryClient: QueryClient) => {
  let coordinator = coordinators.get(queryClient);
  if (!coordinator) {
    coordinator = { requests: new Map(), notices: new Map() };
    coordinators.set(queryClient, coordinator);
  }
  return coordinator;
};

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value && typeof value === "object");

const queriesByFreshness = (
  queryClient: QueryClient,
  queryKey: readonly unknown[],
) =>
  queryClient
    .getQueryCache()
    .findAll({ queryKey })
    .sort(
      (left, right) =>
        right.state.dataUpdatedAt - left.state.dataUpdatedAt ||
        Number(right.isActive()) - Number(left.isActive()) ||
        right.queryKey.length - left.queryKey.length,
    );

const collectStories = (
  data: unknown,
  wantedIds: Set<string>,
  snapshots: Map<string, WipStorySnapshot>,
  parentId: string | null = null,
) => {
  if (Array.isArray(data)) {
    for (const item of data)
      collectStories(item, wantedIds, snapshots, parentId);
    return;
  }
  if (!isRecord(data)) return;

  if (
    typeof data.id === "string" &&
    wantedIds.has(data.id) &&
    !snapshots.has(data.id) &&
    typeof data.statusId === "string"
  ) {
    snapshots.set(data.id, {
      id: data.id,
      statusId: data.statusId,
      teamId: typeof data.teamId === "string" ? data.teamId : undefined,
      parentId:
        typeof data.parentId === "string" && data.parentId
          ? data.parentId
          : parentId,
      archivedAt: typeof data.archivedAt === "string" ? data.archivedAt : null,
      deletedAt: typeof data.deletedAt === "string" ? data.deletedAt : null,
    });
  }

  for (const key of ["data", "groups", "pages", "stories"]) {
    if (data[key]) collectStories(data[key], wantedIds, snapshots, parentId);
  }
  if (data.subStories) {
    collectStories(
      data.subStories,
      wantedIds,
      snapshots,
      typeof data.id === "string" ? data.id : parentId,
    );
  }
};

const getCachedStatus = (
  queryClient: QueryClient,
  workspaceSlug: string,
  statusId: string,
) => {
  const queries = queriesByFreshness(
    queryClient,
    statusKeys.lists(workspaceSlug),
  );
  for (const query of queries) {
    const statuses = queryClient.getQueryData<State[]>(query.queryKey);
    const status = statuses?.find((item) => item.id === statusId);
    if (status) return status;
  }
};

/** Capture before optimistic updates, including records outside the detail cache. */
export const captureWipStorySnapshots = async (
  queryClient: QueryClient,
  workspaceSlug: string,
  storyIds: string[],
  payload: StoryUpdate,
): Promise<WipStorySnapshot[]> => {
  if (!payload.statusId) return [];
  const destination = getCachedStatus(
    queryClient,
    workspaceSlug,
    payload.statusId,
  );
  if (destination && !destination.wipLimit) return [];

  const wantedIds = new Set(storyIds);
  const snapshots = new Map<string, WipStorySnapshot>();
  for (const query of queriesByFreshness(
    queryClient,
    storyKeys.all(workspaceSlug),
  )) {
    collectStories(query.state.data, wantedIds, snapshots);
  }

  // An uncached record must have a known previous status to avoid same-status
  // warnings. A failed lookup skips its optional warning, never the save.
  if (!destination || destination.wipLimit) {
    await Promise.all(
      Array.from(wantedIds)
        .filter((id) => !snapshots.has(id))
        .map(async (id) => {
          try {
            const story = await getStory(id, { workspaceSlug });
            collectStories(story, wantedIds, snapshots);
          } catch (error) {
            posthog.captureException(error, { source: "wip_previous_status" });
          }
        }),
    );
  }
  return Array.from(snapshots.values());
};

export const affectsWipCapacity = (payload: StoryUpdate) =>
  ["statusId", "teamId", "parentId", "archivedAt", "deletedAt"].some(
    (field) => field in payload,
  );

/** Only the filter-free query used by WorkflowCountsProvider qualifies. */
export const isWipCapacityQuery = (
  query: Pick<Query, "queryKey">,
  workspaceSlug: string,
) => {
  const [resource, workspace, kind, params] = query.queryKey;
  return (
    resource === "stories" &&
    workspace === workspaceSlug &&
    kind === "grouped" &&
    query.queryKey.length === 4 &&
    isRecord(params) &&
    params.groupBy === "status" &&
    Object.keys(params).every((key) => key === "groupBy" || key === "teamIds")
  );
};

export const wipCapacityToastId = (workspaceSlug: string, statusId: string) =>
  `wip-capacity:${workspaceSlug}:${statusId}`;

const fetchFreshCapacity = async (
  queryClient: QueryClient,
  workspaceSlug: string,
  params: WipCapacityParams,
  loadCapacity: LoadWipCapacity,
) => {
  const queryKey = [...storyKeys.grouped(workspaceSlug), params];
  const coordinator = getCoordinator(queryClient);
  const requestKey = hashKey(queryKey);
  const token = Symbol(requestKey);
  const request: CapacityRequest = {
    token,
    promise: (async () => {
      await queryClient.cancelQueries({ queryKey, exact: true });
      const data = await loadCapacity(workspaceSlug, params);
      if (coordinator.requests.get(requestKey)?.token === token) {
        queryClient.setQueryData(queryKey, data);
      }
      return data;
    })(),
  };
  coordinator.requests.set(requestKey, request);

  const waitForLatest = (
    current: CapacityRequest,
  ): Promise<WipCapacityResult> =>
    current.promise.then(
      (data) => {
        const latest = coordinator.requests.get(requestKey);
        if (!latest || latest === current) return data;
        return waitForLatest(latest);
      },
      (error: unknown) => {
        const latest = coordinator.requests.get(requestKey);
        if (!latest || latest === current) throw error;
        return waitForLatest(latest);
      },
    );
  return waitForLatest(request);
};

type WipFeedbackOptions = {
  queryClient: QueryClient;
  workspaceSlug: string;
  loadCapacity: LoadWipCapacity;
  snapshots?: WipStorySnapshot[];
  payload?: StoryUpdate;
  successfulStoryIds?: string[];
  createdStory?: WipStorySnapshot;
  toastOptions?: ExternalToast;
};

/** Run after persistence; capacity failures must not turn a saved change into an error. */
export const refreshWipCapacity = async ({
  queryClient,
  workspaceSlug,
  loadCapacity,
  snapshots = [],
  payload = {},
  successfulStoryIds,
  createdStory,
  toastOptions,
}: WipFeedbackOptions) => {
  const successfulIds = successfulStoryIds ? new Set(successfulStoryIds) : null;
  const entrants = createdStory
    ? [createdStory]
    : snapshots
        .filter(
          (story) =>
            (!successfulIds || successfulIds.has(story.id)) &&
            payload.statusId &&
            (story.statusId !== payload.statusId ||
              (payload.teamId && story.teamId !== payload.teamId)),
        )
        .map((story) => ({ ...story, ...payload }));
  const entrant = entrants.find(
    (story) => !story.parentId && !story.archivedAt && !story.deletedAt,
  );
  const statusId = entrant?.statusId;
  const noticeKey = statusId
    ? wipCapacityToastId(workspaceSlug, statusId)
    : null;
  const coordinator = getCoordinator(queryClient);
  const noticeVersion = noticeKey
    ? (coordinator.notices.get(noticeKey) ?? 0) + 1
    : 0;
  if (noticeKey) coordinator.notices.set(noticeKey, noticeVersion);

  try {
    await queryClient.invalidateQueries({
      queryKey: storyKeys.grouped(workspaceSlug),
      predicate: (query) => isWipCapacityQuery(query, workspaceSlug),
      refetchType: "none",
    });

    let status = statusId
      ? getCachedStatus(queryClient, workspaceSlug, statusId)
      : undefined;
    if (statusId && !status) {
      try {
        const statuses = await getStatuses({ workspaceSlug });
        status = statuses.find((item) => item.id === statusId);
        queryClient.setQueryData(statusKeys.lists(workspaceSlug), statuses);
      } catch (error) {
        posthog.captureException(error, { source: "wip_status_limits" });
      }
    }

    const paramsByKey = new Map<string, WipCapacityParams>();
    for (const query of queryClient.getQueryCache().findAll({
      queryKey: storyKeys.grouped(workspaceSlug),
      type: "active",
    })) {
      if (!isWipCapacityQuery(query, workspaceSlug)) continue;
      const params = query.queryKey[3] as WipCapacityParams;
      paramsByKey.set(hashKey([params]), params);
    }
    const destinationParams: WipCapacityParams | null = status?.wipLimit
      ? { groupBy: "status", teamIds: [status.teamId] }
      : null;
    if (destinationParams) {
      paramsByKey.set(hashKey([destinationParams]), destinationParams);
    }

    const results = await Promise.allSettled(
      Array.from(paramsByKey.values(), async (params) => ({
        key: hashKey([params]),
        data: await fetchFreshCapacity(
          queryClient,
          workspaceSlug,
          params,
          loadCapacity,
        ),
      })),
    );
    let destinationCounts: WipCapacityResult | undefined;
    for (const result of results) {
      if (result.status === "rejected") {
        posthog.captureException(result.reason, {
          source: "wip_capacity_refresh",
        });
      } else if (
        destinationParams &&
        result.value.key === hashKey([destinationParams])
      ) {
        destinationCounts = result.value.data;
      }
    }
    if (
      !status?.wipLimit ||
      !noticeKey ||
      coordinator.notices.get(noticeKey) !== noticeVersion ||
      !destinationCounts
    ) {
      return;
    }
    const activeCount =
      destinationCounts.groups.find((group) => group.key === status.id)
        ?.totalCount ?? 0;
    if (activeCount <= status.wipLimit) return;

    toast.warning(
      `${status.name} is over capacity: ${activeCount}/${status.wipLimit}`,
      {
        description: `Reduce by ${activeCount - status.wipLimit} to get within the limit.`,
        ...toastOptions,
        id: noticeKey,
      },
    );
  } catch (error) {
    posthog.captureException(error, { source: "wip_capacity_refresh" });
  }
};
