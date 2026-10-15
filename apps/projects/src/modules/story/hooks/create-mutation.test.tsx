import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  QueryObserver,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { createDesktopStory } from "@/shell/desktop/story-offline";
import { storyKeys } from "@/modules/stories/constants";
import type { GroupedStoriesResponse } from "@/modules/stories/types";
import { createStoryAction } from "../actions/create-story";
import { useCreateStoryMutation } from "./create-mutation";

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: jest.fn() }),
}));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn() },
}));
jest.mock("@/hooks", () => ({
  useAnalytics: () => ({ analytics: { track: jest.fn() } }),
  useTerminology: () => ({ getTermDisplay: () => "story" }),
  useWorkspacePath: () => ({
    workspaceSlug: "acme",
    withWorkspace: (path: string) => path,
  }),
}));
jest.mock("@/modules/stories/public/wip-capacity", () => ({
  loadWipCapacity: jest.fn(),
}));
jest.mock("@/shell/desktop/bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("@/shell/desktop/workspace-scope", () => ({
  useDesktopWorkspaceScope: () => ({
    accountId: "alice",
    workspaceId: "workspace",
    workspaceSlug: "acme",
  }),
}));
jest.mock("@/shell/desktop/story-offline", () => ({
  createDesktopStory: jest.fn(),
  isDesktopPending: jest.fn(),
}));
jest.mock("@/shell/desktop/outbox-projection", () => ({
  retainDesktopCreatedStory: jest.fn(),
}));
jest.mock("../actions/create-story", () => ({ createStoryAction: jest.fn() }));

const group = (key: string, count: number) => ({
  key,
  stories: [{ id: `existing-${key}`, title: key, sequenceId: 1 }],
  totalCount: count,
  loadedCount: 1,
  hasMore: count > 1,
  nextPage: 2,
});
const wrapperFor = (client: QueryClient) =>
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  };

beforeEach(() => {
  jest.resetAllMocks();
});

it("keeps every cached group unchanged when native enqueue fails without adding a web placeholder", async () => {
  const client = new QueryClient({
    defaultOptions: { mutations: { retry: false } },
  });
  const key = storyKeys.grouped("acme");
  const anotherWorkspace = storyKeys.grouped("another-workspace");
  const original = { groups: [group("todo", 12), group("done", 8)] };
  client.setQueryData(key, original);
  client.setQueryData(anotherWorkspace, original);
  jest.mocked(getDesktopBridge).mockReturnValue({
    enqueueOperation: jest.fn(),
  } as unknown as ReturnType<typeof getDesktopBridge>);
  let rejectEnqueue!: (reason: Error) => void;
  jest.mocked(createDesktopStory).mockImplementation(
    () =>
      new Promise((_resolve, reject) => {
        rejectEnqueue = reject;
      }),
  );
  const { result, unmount } = renderHook(() => useCreateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  act(() => {
    result.current.mutate({ title: "Local story", teamId: "team" });
  });
  await waitFor(() => {
    expect(createDesktopStory).toHaveBeenCalledTimes(1);
  });
  // The request's original mutation context decides whether rollback is needed,
  // even when bridge availability changes before the error callback runs.
  jest.mocked(getDesktopBridge).mockReturnValue(null);
  act(() => {
    rejectEnqueue(new Error("Local disk is full"));
  });
  await waitFor(() => {
    expect(result.current.isError).toBe(true);
  });
  expect(client.getQueryData(key)).toEqual(original);
  expect(client.getQueryData(anotherWorkspace)).toEqual(original);
  expect(createStoryAction).not.toHaveBeenCalled();
  expect(toast.error).toHaveBeenCalledWith(
    "Failed to create story: Local story",
    expect.objectContaining({ description: "Local disk is full" }),
  );
  unmount();
  client.clear();
});

it("still removes the web placeholder and restores its count after a browser create fails", async () => {
  const client = new QueryClient({
    defaultOptions: { mutations: { retry: false } },
  });
  const key = storyKeys.grouped("acme");
  const original = { groups: [group("todo", 12)] };
  client.setQueryData(key, original);
  const unsubscribe = new QueryObserver(client, {
    queryKey: key,
    queryFn: () => original,
    staleTime: Infinity,
  }).subscribe(() => undefined);
  jest.mocked(getDesktopBridge).mockReturnValue(null);
  let rejectCreate!: (reason: Error) => void;
  jest.mocked(createStoryAction).mockImplementation(
    () =>
      new Promise((_resolve, reject) => {
        rejectCreate = reject;
      }),
  );
  const { result, unmount } = renderHook(() => useCreateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  act(() => {
    result.current.mutate({ title: "Browser story", teamId: "team" });
  });
  await waitFor(() => {
    expect(createStoryAction).toHaveBeenCalledTimes(1);
  });
  expect(
    client.getQueryData<GroupedStoriesResponse>(key)?.groups[0].totalCount,
  ).toBe(13);
  act(() => {
    rejectCreate(new Error("Server unavailable"));
  });
  await waitFor(() => {
    expect(result.current.isError).toBe(true);
  });
  expect(client.getQueryData(key)).toEqual(original);
  expect(createDesktopStory).not.toHaveBeenCalled();
  unmount();
  unsubscribe();
  client.clear();
});
