import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  QueryObserver,
} from "@tanstack/react-query";
import { toast } from "sonner";
import posthog from "posthog-js";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import { statusKeys } from "@/constants/keys";
import type { State } from "@/types/states";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getStory } from "@/shared/story/queries/get-story";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useCreateStoryMutation } from "@/modules/story/hooks/create-mutation";
import {
  captureWipStorySnapshots,
  refreshWipCapacity,
  wipCapacityToastId,
} from "@/shared/story/wip-capacity";
import { useBulkUpdateStoriesMutation } from "../hooks/update-mutation";
import { bulkUpdateAction } from "../actions/bulk-update-stories";
import { getGroupedStories } from "../queries/get-grouped-stories";
import { storyKeys } from "../constants";
import { loadWipCapacity } from "../public/wip-capacity";
import type { DetailedStory, GroupedStoriesResponse } from "../types";

jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(),
  useTerminology: jest.fn(),
  useWorkspacePath: jest.fn(),
}));
jest.mock("next/navigation", () => ({
  useParams: () => ({}),
  useRouter: () => ({ push: jest.fn() }),
}));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn(), warning: jest.fn() },
}));
jest.mock("posthog-js", () => ({
  __esModule: true,
  default: { captureException: jest.fn() },
}));
jest.mock("../queries/get-grouped-stories", () => ({
  getGroupedStories: jest.fn(),
}));
jest.mock("../actions/bulk-update-stories", () => ({
  bulkUpdateAction: jest.fn(),
}));
jest.mock("@/lib/queries/states/get-states", () => ({
  getStatuses: jest.fn(),
}));
jest.mock("@/shared/story/queries/get-story", () => ({ getStory: jest.fn() }));
jest.mock("@/modules/story/actions/update-story", () => ({
  updateStoryAction: jest.fn(),
}));
jest.mock("@/modules/story/actions/create-story", () => ({
  createStoryAction: jest.fn(),
}));

const WORKSPACE = "forty-one";
const captureExceptionSpy = jest.spyOn(posthog, "captureException");
const TEAM = "team-one";
const STATUS = "started";
const CAPACITY_PARAMS = { groupBy: "status", teamIds: [TEAM] } as const;
const capacityKey = [...storyKeys.grouped(WORKSPACE), CAPACITY_PARAMS];
const destination = {
  id: STATUS,
  teamId: TEAM,
  name: "In progress",
  wipLimit: 5,
} as State;
const story = (id = "story-one", overrides: Partial<DetailedStory> = {}) =>
  ({
    id,
    title: "Task",
    statusId: "backlog",
    teamId: TEAM,
    parentId: "",
    archivedAt: null,
    deletedAt: null,
    subStories: [],
    ...overrides,
  }) as DetailedStory;
const counts = (count: number): GroupedStoriesResponse => ({
  groups: [
    {
      key: STATUS,
      totalCount: count,
      loadedCount: 0,
      stories: [],
      hasMore: true,
      nextPage: 1,
    },
  ],
  meta: {
    groupBy: "status",
    totalGroups: 1,
    filters: {},
    orderBy: "created",
    orderDirection: "desc",
  },
});
const createClient = () => {
  const client = new QueryClient({
    defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
  });
  client.setQueryData(statusKeys.team(WORKSPACE, TEAM), [destination]);
  return client;
};
const wrapperFor = (client: QueryClient) =>
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  };
const observeCounts = (client: QueryClient) => {
  client.setQueryData(capacityKey, counts(9));
  const observer = new QueryObserver(client, {
    queryKey: capacityKey,
    queryFn: async () => counts(99),
    staleTime: Infinity,
  });
  return observer.subscribe(() => undefined);
};

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(useWorkspacePath).mockReturnValue({
    workspaceSlug: WORKSPACE,
    withWorkspace: (path: string) => path,
  } as ReturnType<typeof useWorkspacePath>);
  jest.mocked(useAnalytics).mockReturnValue({
    analytics: { track: jest.fn() },
  } as unknown as ReturnType<typeof useAnalytics>);
  jest.mocked(useTerminology).mockReturnValue({
    getTermDisplay: (_term, options) =>
      options?.variant === "plural" ? "tasks" : "task",
  } as ReturnType<typeof useTerminology>);
  jest.mocked(getGroupedStories).mockResolvedValue(counts(6));
  jest.mocked(updateStoryAction).mockResolvedValue({ data: null });
  jest.mocked(getStatuses).mockResolvedValue([destination]);
});

it("waits for single persistence and refreshes counts even when the moved task is outside loaded count results", async () => {
  const client = createClient();
  client.setQueryData(storyKeys.list(WORKSPACE, "filtered"), [story()]);
  const unsubscribe = observeCounts(client);
  let persist: (() => void) | undefined;
  jest.mocked(updateStoryAction).mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        persist = () => {
          resolve({ data: null });
        };
      }),
  );
  const { result } = renderHook(() => useUpdateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  act(() => {
    result.current.mutate({
      storyId: "story-one",
      payload: { statusId: STATUS },
    });
  });
  await waitFor(() => {
    expect(updateStoryAction).toHaveBeenCalled();
  });
  expect(getGroupedStories).not.toHaveBeenCalled();
  expect(toast.warning).not.toHaveBeenCalled();
  act(() => persist?.());
  await waitFor(() => {
    expect(toast.warning).toHaveBeenCalledTimes(1);
  });
  expect(toast.warning).toHaveBeenCalledWith(
    "In progress is over capacity: 6/5",
    expect.objectContaining({
      id: wipCapacityToastId(WORKSPACE, STATUS),
      description: "Reduce by 1 to get within the limit.",
    }),
  );
  expect(getGroupedStories).toHaveBeenCalledWith(
    { workspaceSlug: WORKSPACE },
    { groupBy: "status", teamIds: [TEAM] },
  );
  expect(
    client.getQueryData<GroupedStoriesResponse>(capacityKey)?.groups[0]
      ?.totalCount,
  ).toBe(6);
  expect(result.current.isSuccess).toBe(true);
  unsubscribe();
});

it.each([
  { name: "same status", value: story("story-one", { statusId: STATUS }) },
  { name: "sub-task", value: story("story-one", { parentId: "parent" }) },
  {
    name: "archived task",
    value: story("story-one", { archivedAt: "2026-10-04" }),
  },
  {
    name: "deleted task",
    value: story("story-one", { deletedAt: "2026-10-04" }),
  },
])("refreshes active counts without nudging for a $name", async ({ value }) => {
  const client = createClient();
  client.setQueryData(storyKeys.detail(WORKSPACE, value.id), value);
  const unsubscribe = observeCounts(client);
  const { result } = renderHook(() => useUpdateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  await act(async () =>
    result.current.mutateAsync({
      storyId: value.id,
      payload: { statusId: STATUS },
    }),
  );
  await waitFor(() => {
    expect(getGroupedStories).toHaveBeenCalled();
  });
  expect(toast.warning).not.toHaveBeenCalled();
  unsubscribe();
});

it("does not warn at the limit or after a failed save", async () => {
  const client = createClient();
  client.setQueryData(storyKeys.detail(WORKSPACE, "story-one"), story());
  jest.mocked(getGroupedStories).mockResolvedValueOnce(counts(5));
  const { result } = renderHook(() => useUpdateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  await act(async () =>
    result.current.mutateAsync({
      storyId: "story-one",
      payload: { statusId: STATUS },
    }),
  );
  await waitFor(() => {
    expect(getGroupedStories).toHaveBeenCalledTimes(1);
  });
  expect(toast.warning).not.toHaveBeenCalled();
  jest
    .mocked(updateStoryAction)
    .mockResolvedValueOnce({ error: { message: "Save rejected" }, data: null });
  act(() => {
    result.current.mutate({
      storyId: "story-one",
      payload: { statusId: STATUS },
    });
  });
  await waitFor(() => {
    expect(result.current.isError).toBe(true);
  });
  expect(getGroupedStories).toHaveBeenCalledTimes(1);
  expect(toast.warning).not.toHaveBeenCalled();
});

it("deduplicates bulk IDs and warns only for confirmed successes in a partial response", async () => {
  const client = createClient();
  client.setQueryData(storyKeys.list(WORKSPACE, "filtered"), [
    story("one"),
    story("two"),
    story("unchanged", { statusId: STATUS }),
  ]);
  jest.mocked(bulkUpdateAction).mockResolvedValueOnce({
    data: {
      totalCount: 3,
      succeededCount: 2,
      failedCount: 1,
      partial: true,
      items: [
        { storyId: "one", success: true },
        { storyId: "unchanged", success: true },
        { storyId: "two", success: false, error: "Cannot update task" },
      ],
    },
  });
  const { result } = renderHook(() => useBulkUpdateStoriesMutation(), {
    wrapper: wrapperFor(client),
  });
  act(() => {
    result.current.mutate({
      storyIds: ["one", "one", "two", "unchanged"],
      payload: { statusId: STATUS },
    });
  });
  await waitFor(() => {
    expect(toast.warning).toHaveBeenCalledTimes(1);
  });
  expect(result.current.isError).toBe(true);
  expect(toast.error).toHaveBeenCalledTimes(1);
  expect(bulkUpdateAction).toHaveBeenCalledWith(
    { storyIds: ["one", "two", "unchanged"], updates: { statusId: STATUS } },
    WORKSPACE,
  );
  expect(toast.warning).toHaveBeenCalledWith(
    "In progress is over capacity: 6/5",
    expect.any(Object),
  );
});

it("does not warn when only an unchanged task succeeds in a partial bulk response", async () => {
  const client = createClient();
  client.setQueryData(storyKeys.list(WORKSPACE, "filtered"), [
    story("one", { statusId: STATUS }),
    story("two"),
  ]);
  const unsubscribe = observeCounts(client);
  jest.mocked(bulkUpdateAction).mockResolvedValueOnce({
    data: {
      totalCount: 2,
      succeededCount: 1,
      failedCount: 1,
      partial: true,
      items: [
        { storyId: "one", success: true },
        { storyId: "two", success: false },
      ],
    },
  });
  const { result } = renderHook(() => useBulkUpdateStoriesMutation(), {
    wrapper: wrapperFor(client),
  });
  act(() => {
    result.current.mutate({
      storyIds: ["one", "two"],
      payload: { statusId: STATUS },
    });
  });
  await waitFor(() => {
    expect(getGroupedStories).toHaveBeenCalled();
  });
  expect(toast.warning).not.toHaveBeenCalled();
  unsubscribe();
});

it("updates the creation success toast to the capacity warning and preserves its View action", async () => {
  const client = createClient();
  jest
    .mocked(createStoryAction)
    .mockResolvedValueOnce({ data: story("new", { statusId: STATUS }) });
  const { result } = renderHook(() => useCreateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  await act(async () =>
    result.current.mutateAsync({
      title: "Task",
      teamId: TEAM,
    }),
  );
  await waitFor(() => {
    expect(toast.warning).toHaveBeenCalledTimes(1);
  });
  const successOptions = jest.mocked(toast.success).mock.calls[0]?.[1];
  const warningOptions = jest.mocked(toast.warning).mock.calls[0]?.[1];
  expect(successOptions?.id).toBe(warningOptions?.id);
  expect(warningOptions?.action).toBe(successOptions?.action);
  expect(result.current.isSuccess).toBe(true);
});

it("uses the persisted destination team and server total for a successful bulk move", async () => {
  const client = createClient();
  const destinationTeam = "team-two";
  client.setQueryData(statusKeys.team(WORKSPACE, destinationTeam), [
    { ...destination, id: "other-started", teamId: destinationTeam },
  ]);
  client.setQueryData(storyKeys.list(WORKSPACE, "filtered"), [
    story("one"),
    story("two"),
  ]);
  jest.mocked(bulkUpdateAction).mockResolvedValueOnce({
    data: {
      totalCount: 2,
      succeededCount: 2,
      failedCount: 0,
      partial: false,
      items: [
        { storyId: "one", success: true },
        { storyId: "two", success: true },
      ],
    },
  });
  jest.mocked(getGroupedStories).mockResolvedValueOnce({
    ...counts(7),
    groups: [{ ...counts(7).groups[0], key: "other-started" }],
  });
  const { result } = renderHook(() => useBulkUpdateStoriesMutation(), {
    wrapper: wrapperFor(client),
  });
  await act(async () =>
    result.current.mutateAsync({
      storyIds: ["one", "one", "two"],
      payload: { statusId: "other-started", teamId: destinationTeam },
    }),
  );
  await waitFor(() => {
    expect(toast.warning).toHaveBeenCalledTimes(1);
  });
  expect(getGroupedStories).toHaveBeenCalledWith(
    { workspaceSlug: WORKSPACE },
    { groupBy: "status", teamIds: [destinationTeam] },
  );
  expect(toast.warning).toHaveBeenCalledWith(
    "In progress is over capacity: 7/5",
    expect.any(Object),
  );
  expect(result.current.isSuccess).toBe(true);
});

it("keeps a persisted save successful if fresh counts fail", async () => {
  const client = createClient();
  client.setQueryData(storyKeys.detail(WORKSPACE, "story-one"), story());
  jest
    .mocked(getGroupedStories)
    .mockRejectedValueOnce(new Error("Count service unavailable"));
  const { result } = renderHook(() => useUpdateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  await act(async () =>
    result.current.mutateAsync({
      storyId: "story-one",
      payload: { statusId: STATUS },
    }),
  );
  await waitFor(() => {
    expect(captureExceptionSpy).toHaveBeenCalled();
  });
  expect(result.current.isSuccess).toBe(true);
  expect(toast.error).not.toHaveBeenCalled();
  expect(toast.warning).not.toHaveBeenCalled();
});

it("resolves previous status when both the status metadata and story cache are absent", async () => {
  const client = new QueryClient();
  jest.mocked(getStory).mockResolvedValueOnce(story());
  const snapshots = await captureWipStorySnapshots(
    client,
    WORKSPACE,
    ["story-one", "story-one"],
    { statusId: STATUS },
  );
  expect(getStory).toHaveBeenCalledTimes(1);
  await refreshWipCapacity({
    queryClient: client,
    workspaceSlug: WORKSPACE,
    loadCapacity: loadWipCapacity,
    snapshots,
    payload: { statusId: STATUS },
  });
  expect(getStatuses).toHaveBeenCalledTimes(1);
  expect(toast.warning).toHaveBeenCalledTimes(1);
});

it("uses a newly configured team limit over older workspace status data", async () => {
  const client = new QueryClient();
  client.setQueryData(
    statusKeys.lists(WORKSPACE),
    [{ ...destination, wipLimit: null }],
    { updatedAt: 1 },
  );
  client.setQueryData(statusKeys.team(WORKSPACE, TEAM), [destination], {
    updatedAt: 2,
  });
  client.setQueryData(storyKeys.detail(WORKSPACE, "story-one"), story());
  const { result } = renderHook(() => useUpdateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  await act(async () =>
    result.current.mutateAsync({
      storyId: "story-one",
      payload: { statusId: STATUS },
    }),
  );
  await waitFor(() => {
    expect(toast.warning).toHaveBeenCalledTimes(1);
  });
  expect(getStatuses).not.toHaveBeenCalled();
});

it("uses a newer list status over a stale inactive detail when deciding whether a move entered the status", async () => {
  const client = createClient();
  client.setQueryData(
    storyKeys.detail(WORKSPACE, "story-one"),
    story("story-one", { statusId: STATUS }),
    { updatedAt: 1 },
  );
  client.setQueryData(storyKeys.list(WORKSPACE, "filtered"), [story()], {
    updatedAt: 2,
  });
  const snapshots = await captureWipStorySnapshots(
    client,
    WORKSPACE,
    ["story-one"],
    { statusId: STATUS },
  );
  expect(snapshots[0]?.statusId).toBe("backlog");
  await refreshWipCapacity({
    queryClient: client,
    workspaceSlug: WORKSPACE,
    loadCapacity: loadWipCapacity,
    snapshots,
    payload: { statusId: STATUS },
  });
  expect(toast.warning).toHaveBeenCalledTimes(1);
});

it("captures page records and recognizes nested sub-tasks before optimism", async () => {
  const client = createClient();
  client.setQueryData(storyKeys.groupStories(WORKSPACE, "backlog", {}), {
    pages: [{ stories: [story("parent", { subStories: [story("child")] })] }],
  });
  const snapshots = await captureWipStorySnapshots(
    client,
    WORKSPACE,
    ["parent", "child"],
    { statusId: STATUS },
  );
  expect(snapshots).toEqual(
    expect.arrayContaining([
      expect.objectContaining({ id: "parent", parentId: null }),
      expect.objectContaining({ id: "child", parentId: "parent" }),
    ]),
  );
  expect(getStory).not.toHaveBeenCalled();
});

it("keeps the latest server count and emits one warning when refreshes complete out of order", async () => {
  const client = createClient();
  let finishFirst: ((value: GroupedStoriesResponse) => void) | undefined;
  jest.mocked(getGroupedStories).mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        finishFirst = resolve;
      }),
  );
  const options = {
    queryClient: client,
    workspaceSlug: WORKSPACE,
    loadCapacity: loadWipCapacity,
    snapshots: [
      {
        id: "one",
        statusId: "backlog",
        teamId: TEAM,
        parentId: null,
        archivedAt: null,
        deletedAt: null,
      },
    ],
    payload: { statusId: STATUS },
  };
  const first = refreshWipCapacity(options);
  await waitFor(() => {
    expect(getGroupedStories).toHaveBeenCalledTimes(1);
  });
  const second = refreshWipCapacity(options);
  await second;
  finishFirst?.(counts(10));
  await first;
  expect(toast.warning).toHaveBeenCalledTimes(1);
  expect(toast.warning).toHaveBeenCalledWith(
    "In progress is over capacity: 6/5",
    expect.any(Object),
  );
  expect(
    client.getQueryData<GroupedStoriesResponse>(capacityKey)?.groups[0]
      ?.totalCount,
  ).toBe(6);
});
