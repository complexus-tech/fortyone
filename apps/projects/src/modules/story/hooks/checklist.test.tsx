import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import { refreshWipCapacity } from "@/shared/story/wip-capacity";
import { updateChecklistAction } from "../actions/checklist";
import type { StoryChecklistItem } from "../types/checklist";
import { useChecklistMutation } from "./checklist";

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

beforeEach(() => jest.clearAllMocks());

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
      createdStory: expect.objectContaining({
        id: "converted",
        parentId: null,
        statusId: "unstarted",
      }),
    }),
  );
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
