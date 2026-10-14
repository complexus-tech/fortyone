import { randomUUID } from "node:crypto";
import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import type { StoryActivity } from "@/modules/stories/types";
import { refreshWipCapacity } from "@/shared/story/wip-capacity";
import { updateChecklistAction } from "../actions/checklist";
import { getStoryChecklist } from "../queries/get-checklist";
import { getStoryActivities } from "../queries/get-activities";
import type { ChecklistCommand, StoryChecklistItem } from "../types/checklist";
import { useChecklistMutation, useStoryChecklist } from "./checklist";
import { useStoryActivitiesInfinite } from "./story-activities";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "viewer" } } }),
}));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn() },
}));
jest.mock("../actions/checklist", () => ({ updateChecklistAction: jest.fn() }));
jest.mock("../queries/get-checklist", () => ({ getStoryChecklist: jest.fn() }));
jest.mock("../queries/get-activities", () => ({
  getStoryActivities: jest.fn(),
}));
jest.mock("@/shared/story/wip-capacity", () => ({
  refreshWipCapacity: jest.fn(),
}));
jest.mock("@/modules/stories/public/wip-capacity", () => ({
  loadWipCapacity: jest.fn(),
}));

const item: StoryChecklistItem = {
  id: "first",
  storyId: "story",
  title: "Write tests",
  completed: false,
  position: 0,
  convertedAt: null,
  convertedStory: null,
  createdAt: "2026-10-13T12:00:00Z",
  updatedAt: "2026-10-13T12:00:00Z",
};
const key = storyKeys.checklist("acme", "story");
const randomUUIDDescriptor = Object.getOwnPropertyDescriptor(
  globalThis.crypto,
  "randomUUID",
);

beforeAll(() => {
  Object.defineProperty(globalThis.crypto, "randomUUID", {
    configurable: true,
    value: randomUUID,
  });
});

afterAll(() => {
  if (randomUUIDDescriptor) {
    Object.defineProperty(
      globalThis.crypto,
      "randomUUID",
      randomUUIDDescriptor,
    );
  } else {
    Reflect.deleteProperty(globalThis.crypto, "randomUUID");
  }
});

const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
};

const setup = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  client.setQueryData(key, [
    item,
    { ...item, id: "second", title: "Review tests", position: 1 },
  ]);
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  const hook = renderHook(() => useChecklistMutation("story"), { wrapper });
  return { ...hook, client, wrapper };
};

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(getStoryChecklist).mockReset();
  jest.mocked(getStoryActivities).mockReset();
});

it.each(["success", "lost response"])(
  "refreshes persisted activity without realtime after a checklist save (%s)",
  async (outcome) => {
    const { client, result, wrapper } = setup();
    const activityKey = storyKeys.activitiesInfinite("acme", "story");
    const emptyPage = {
      activities: [],
      pagination: { page: 1, pageSize: 20, hasMore: false, nextPage: 2 },
    };
    client.setQueryData(activityKey, { pages: [emptyPage], pageParams: [1] });
    const activity: StoryActivity = {
      id: "completion",
      storyId: "story",
      userId: "viewer",
      user: {
        id: "viewer",
        username: "joseph",
        fullName: "Joseph",
        avatarUrl: null,
        isActive: true,
        isSystem: false,
      },
      type: "update",
      field: "checklist_item",
      currentValue: item.title,
      oldValue: { id: item.id, title: item.title, completed: false },
      newValue: { id: item.id, title: item.title, completed: true },
      reason: "checklist_item_completed",
      createdAt: "2026-10-14T10:30:00Z",
    };
    jest.mocked(getStoryActivities).mockResolvedValue({
      ...emptyPage,
      activities: [activity],
    });
    const feed = renderHook(() => useStoryActivitiesInfinite("story"), {
      wrapper,
    });
    const save = deferred<StoryChecklistItem>();
    jest.mocked(updateChecklistAction).mockReturnValue(save.promise);

    act(() => {
      result.current.mutate({
        type: "update",
        itemId: item.id,
        updates: { completed: true },
      });
    });
    await waitFor(() => {
      expect(updateChecklistAction).toHaveBeenCalledTimes(1);
    });
    expect(getStoryActivities).not.toHaveBeenCalled();
    expect(feed.result.current.data?.pages[0].activities).toEqual([]);

    act(() => {
      if (outcome === "success") {
        save.resolve({ ...item, completed: true });
      } else {
        save.reject(new Error("Connection lost after saving"));
      }
    });
    await waitFor(() => {
      expect(feed.result.current.data?.pages[0].activities).toEqual([activity]);
    });
    expect(getStoryActivities).toHaveBeenCalledTimes(1);
  },
);

it.each<ChecklistCommand>([
  { type: "create", title: "New task" },
  { type: "update", itemId: item.id, updates: { title: "Renamed task" } },
  { type: "update", itemId: item.id, updates: { completed: true } },
  { type: "delete", itemId: item.id },
  { type: "clear" },
  { type: "reorder", itemIds: ["second", "first"] },
])("invalidates only the owning activity feed after $type", async (command) => {
  const { client, result } = setup();
  const activityKey = storyKeys.activitiesInfinite("acme", "story");
  const otherStoryKey = storyKeys.activitiesInfinite("acme", "other");
  const otherWorkspaceKey = storyKeys.activitiesInfinite("other", "story");
  for (const queryKey of [activityKey, otherStoryKey, otherWorkspaceKey]) {
    client.setQueryData(queryKey, { pages: [], pageParams: [] });
  }
  jest.mocked(updateChecklistAction).mockResolvedValue(null);

  await act(async () => {
    await result.current.mutateAsync(command);
  });

  expect(client.getQueryState(activityKey)?.isInvalidated).toBe(true);
  expect(client.getQueryState(otherStoryKey)?.isInvalidated).toBe(false);
  expect(client.getQueryState(otherWorkspaceKey)?.isInvalidated).toBe(false);
});

it("appends a new item before saving and reconciles by identity without removing matching titles", async () => {
  const { client, result } = setup();
  const existingItems = [
    { ...item, position: 9 },
    { ...item, id: "second", title: "Review tests", position: 1 },
  ];
  client.setQueryData(key, existingItems);
  const save = deferred<StoryChecklistItem>();
  jest.mocked(updateChecklistAction).mockReturnValue(save.promise);

  act(() => {
    result.current.mutate({ type: "create", title: item.title });
  });
  await waitFor(() => {
    expect(client.getQueryData<StoryChecklistItem[]>(key)).toHaveLength(3);
  });
  const optimisticItem = client.getQueryData<StoryChecklistItem[]>(key)![2];
  expect(optimisticItem).toEqual(
    expect.objectContaining({
      storyId: "story",
      title: item.title,
      completed: false,
      position: 10,
      convertedAt: null,
      convertedStory: null,
    }),
  );
  expect(existingItems.map(({ id }) => id)).not.toContain(optimisticItem.id);
  expect(result.current.isPending).toBe(true);

  const persistedItem = { ...optimisticItem, id: "persisted" };
  // A server row already in the cache must not duplicate the optimistic row.
  client.setQueryData<StoryChecklistItem[]>(key, (items = []) => [
    ...items,
    persistedItem,
  ]);
  act(() => {
    save.resolve(persistedItem);
  });
  await waitFor(() => {
    expect(result.current.isPending).toBe(false);
  });
  expect(client.getQueryData(key)).toEqual([...existingItems, persistedItem]);
  expect(toast.success).not.toHaveBeenCalled();
  expect(toast.error).not.toHaveBeenCalled();
});

it("shows the first item immediately when the checklist has no cached data", async () => {
  const { client, result } = setup();
  client.removeQueries({ queryKey: key, exact: true });
  const save = deferred<StoryChecklistItem>();
  jest.mocked(updateChecklistAction).mockReturnValue(save.promise);

  act(() => {
    result.current.mutate({ type: "create", title: "First checklist item" });
  });
  await waitFor(() => {
    expect(client.getQueryData<StoryChecklistItem[]>(key)).toHaveLength(1);
  });
  const optimisticItem = client.getQueryData<StoryChecklistItem[]>(key)![0];
  expect(optimisticItem).toEqual(
    expect.objectContaining({
      title: "First checklist item",
      completed: false,
      position: 0,
    }),
  );
  expect(result.current.isPending).toBe(true);

  const persistedItem = { ...optimisticItem, id: "persisted" };
  act(() => {
    save.resolve(persistedItem);
  });
  await waitFor(() => {
    expect(result.current.isPending).toBe(false);
  });
  expect(client.getQueryData(key)).toEqual([persistedItem]);
  expect(toast.success).not.toHaveBeenCalled();
  expect(toast.error).not.toHaveBeenCalled();
});

it.each(["populated", "not yet loaded"])(
  "removes a failed optimistic creation and restores a %s checklist",
  async (cacheState) => {
    const { client, result } = setup();
    if (cacheState === "not yet loaded") {
      client.removeQueries({ queryKey: key, exact: true });
    }
    const previousItems = client.getQueryData<StoryChecklistItem[]>(key) ?? [];
    const save = deferred<StoryChecklistItem>();
    jest.mocked(updateChecklistAction).mockReturnValue(save.promise);

    act(() => {
      result.current.mutate({ type: "create", title: "A failed creation" });
    });
    await waitFor(() => {
      expect(client.getQueryData<StoryChecklistItem[]>(key)).toHaveLength(
        previousItems.length + 1,
      );
    });
    act(() => {
      save.reject(new Error("Unable to connect"));
    });
    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
    expect(client.getQueryData(key)).toEqual(previousItems);
    expect(toast.error).toHaveBeenCalledTimes(1);
    expect(toast.error).toHaveBeenCalledWith(
      "Unable to save checklist changes",
      expect.objectContaining({
        description: expect.stringContaining("Unable to connect"),
      }),
    );
    expect(toast.success).not.toHaveBeenCalled();
  },
);

it("keeps an optimistic creation visible through realtime invalidation and guards writes until refetch finishes", async () => {
  const { client, result, wrapper } = setup();
  const existingItems = client.getQueryData<StoryChecklistItem[]>(key)!;
  const save = deferred<StoryChecklistItem>();
  const refresh = deferred<StoryChecklistItem[]>();
  jest.mocked(updateChecklistAction).mockReturnValue(save.promise);
  jest.mocked(getStoryChecklist).mockReturnValue(refresh.promise);
  const checklist = renderHook(() => useStoryChecklist("story"), { wrapper });
  const secondSurface = renderHook(() => useChecklistMutation("story"), {
    wrapper,
  });

  act(() => {
    result.current.mutate({ type: "create", title: "Keep me visible" });
  });
  await waitFor(() => {
    expect(checklist.result.current.data).toHaveLength(3);
  });
  const optimisticItem = checklist.result.current.data![2];
  let invalidation = Promise.resolve();
  act(() => {
    // Story SSE events invalidate the detail subtree, including its checklist.
    invalidation = client.invalidateQueries({
      queryKey: storyKeys.detail("acme", "story"),
    });
  });
  await waitFor(() => {
    expect(client.getQueryState(key)?.fetchStatus).toBe("fetching");
  });
  expect(getStoryChecklist).not.toHaveBeenCalled();
  expect(checklist.result.current.data).toEqual([
    ...existingItems,
    optimisticItem,
  ]);

  const persistedItem = { ...optimisticItem, id: "persisted" };
  act(() => {
    save.resolve(persistedItem);
  });
  await waitFor(() => {
    expect(getStoryChecklist).toHaveBeenCalled();
    expect(checklist.result.current.data).toEqual([
      ...existingItems,
      persistedItem,
    ]);
  });
  expect(result.current.isPending).toBe(true);
  await act(async () => {
    await expect(
      secondSurface.result.current.mutateAsync({ type: "clear" }),
    ).rejects.toThrow("Wait for the current");
  });
  expect(updateChecklistAction).toHaveBeenCalledTimes(1);
  expect(client.getQueryData(key)).toEqual([...existingItems, persistedItem]);

  act(() => {
    refresh.resolve([...existingItems, persistedItem]);
  });
  await waitFor(() => {
    expect(result.current.isPending).toBe(false);
    expect(client.getQueryState(key)?.fetchStatus).toBe("idle");
  });
  await invalidation;
  expect(checklist.result.current.data).toEqual([
    ...existingItems,
    persistedItem,
  ]);
  expect(toast.success).not.toHaveBeenCalled();
});

it("checks an item immediately and rolls back a rejected save without touching sub-stories", async () => {
  const { client, result } = setup();
  const storyKey = storyKeys.detail("acme", "story");
  const story = {
    id: "story",
    subStories: [{ id: "child", statusId: "unstarted" }],
  };
  client.setQueryData(storyKey, story);
  let reject!: (error: Error) => void;
  jest.mocked(updateChecklistAction).mockImplementation(
    () =>
      new Promise((_resolve, rejectPromise) => {
        reject = rejectPromise;
      }),
  );
  act(() => {
    result.current.mutate({
      type: "update",
      itemId: "first",
      updates: { completed: true },
    });
  });
  await waitFor(() => {
    expect(client.getQueryData<StoryChecklistItem[]>(key)?.[0].completed).toBe(
      true,
    );
  });
  expect(client.getQueryData(storyKey)).toEqual(story);
  act(() => {
    reject(new Error("Permission denied"));
  });
  await waitFor(() => {
    expect(result.current.isPending).toBe(false);
  });
  expect(client.getQueryData<StoryChecklistItem[]>(key)?.[0].completed).toBe(
    false,
  );
  expect(toast.error).toHaveBeenCalled();
  expect(client.getQueryState(key)?.isInvalidated).toBe(true);
});

it("restores the original order if a stale reorder is rejected", async () => {
  const { client, result } = setup();
  let reject!: (error: Error) => void;
  jest.mocked(updateChecklistAction).mockImplementation(
    () =>
      new Promise((_resolve, rejectPromise) => {
        reject = rejectPromise;
      }),
  );
  act(() => {
    result.current.mutate({ type: "reorder", itemIds: ["second", "first"] });
  });
  await waitFor(() => {
    expect(
      client.getQueryData<StoryChecklistItem[]>(key)?.map(({ id }) => id),
    ).toEqual(["second", "first"]);
  });
  act(() => {
    reject(new Error("Checklist changed"));
  });
  await waitFor(() => {
    expect(result.current.isPending).toBe(false);
  });
  expect(
    client.getQueryData<StoryChecklistItem[]>(key)?.map(({ id }) => id),
  ).toEqual(["first", "second"]);
});

it("retains the checklist row after conversion and refreshes story collections and relationships", async () => {
  const { client, result } = setup();
  const otherWorkspaceKey = storyKeys.checklist("other", "story");
  client.setQueryData(otherWorkspaceKey, [item]);
  client.setQueryData(storyKeys.detail("acme", "story"), {
    id: "story",
    associations: [],
  });
  client.setQueryData(storyKeys.total("acme"), 3);
  const converted = {
    ...item,
    completed: true,
    convertedAt: "2026-10-13T12:01:00Z",
    convertedStory: {
      id: "converted",
      title: item.title,
      sequenceId: 231,
      teamId: "team",
      statusId: "unstarted",
    },
  };
  jest.mocked(updateChecklistAction).mockResolvedValue(converted);
  await act(async () => {
    await result.current.mutateAsync({ type: "convert", itemId: "first" });
  });
  expect(client.getQueryData<StoryChecklistItem[]>(key)?.[0]).toEqual(
    converted,
  );
  expect(client.getQueryData<StoryChecklistItem[]>(key)).toHaveLength(2);
  expect(
    client.getQueryState(storyKeys.detail("acme", "story"))?.isInvalidated,
  ).toBe(true);
  expect(client.getQueryState(storyKeys.total("acme"))?.isInvalidated).toBe(
    true,
  );
  expect(client.getQueryState(otherWorkspaceKey)?.isInvalidated).toBe(false);
  expect(refreshWipCapacity).toHaveBeenCalledWith(
    expect.objectContaining({
      workspaceSlug: "acme",
      showCapacityWarning: false,
      createdStory: expect.objectContaining({
        id: "converted",
        parentId: null,
        statusId: "unstarted",
      }),
    }),
  );
  expect(toast.success).not.toHaveBeenCalled();
  expect(toast.error).not.toHaveBeenCalled();
});

it("prevents a second mounted surface from overwriting an in-flight change", async () => {
  const { client, result, wrapper } = setup();
  const second = renderHook(() => useChecklistMutation("story"), { wrapper });
  let persist!: (data: StoryChecklistItem) => void;
  jest.mocked(updateChecklistAction).mockImplementation(
    () =>
      new Promise((resolve) => {
        persist = resolve;
      }),
  );
  act(() => {
    result.current.mutate({
      type: "update",
      itemId: "first",
      updates: { completed: true },
    });
  });
  await waitFor(() => {
    expect(second.result.current.isPending).toBe(true);
  });
  await act(async () => {
    await expect(
      second.result.current.mutateAsync({ type: "clear" }),
    ).rejects.toThrow("Wait for the current");
  });
  expect(updateChecklistAction).toHaveBeenCalledTimes(1);
  expect(client.getQueryData<StoryChecklistItem[]>(key)).toHaveLength(2);
  act(() => {
    persist({ ...item, completed: true });
  });
  await waitFor(() => {
    expect(result.current.isPending).toBe(false);
  });
});

it("accepts exactly one write when two surfaces submit in the same tick", async () => {
  const { result, wrapper } = setup();
  const second = renderHook(() => useChecklistMutation("story"), { wrapper });
  jest
    .mocked(updateChecklistAction)
    .mockResolvedValue({ ...item, completed: true });
  let outcomes: PromiseSettledResult<unknown>[] = [];
  await act(async () => {
    outcomes = await Promise.allSettled([
      result.current.mutateAsync({
        type: "update",
        itemId: "first",
        updates: { completed: true },
      }),
      second.result.current.mutateAsync({ type: "clear" }),
    ]);
  });
  expect(outcomes.map(({ status }) => status).sort()).toEqual([
    "fulfilled",
    "rejected",
  ]);
  expect(updateChecklistAction).toHaveBeenCalledTimes(1);
  await act(async () => {
    await second.result.current.mutateAsync({
      type: "delete",
      itemId: "second",
    });
  });
  expect(updateChecklistAction).toHaveBeenCalledTimes(2);
});
