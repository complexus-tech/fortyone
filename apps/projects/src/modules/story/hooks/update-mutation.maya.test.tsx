/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  QueryObserver,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { GroupedStoriesResponse } from "@/modules/stories/types";
import type { DetailedStory } from "../types";
import { updateStoryAction } from "../actions/update-story";
import { useUpdateStoryMutation } from "./update-mutation";

jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(),
  useTerminology: jest.fn(),
  useWorkspacePath: jest.fn(),
}));
jest.mock("../actions/update-story", () => ({ updateStoryAction: jest.fn() }));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const WORKSPACE_SLUG = "workspace-a";
const STORY_ID = "story-a";
const initialStory = {
  id: STORY_ID,
  statusId: "backlog",
  priority: "Medium",
  subStories: [],
} as unknown as DetailedStory;
const recentKey = [
  ...storyKeys.detail(WORKSPACE_SLUG, STORY_ID),
  "maya-recent",
  "user-a",
] as const;
const workKey = [
  ...storyKeys.mineGrouped(WORKSPACE_SLUG, { assignedToMe: true }),
  "maya-work",
  "user-a",
] as const;

const grouped = (story: DetailedStory) =>
  ({
    groups: [{ key: "none", stories: [story] }],
    meta: { groupBy: "none" },
  }) as unknown as GroupedStoriesResponse;

const createClient = () =>
  new QueryClient({ defaultOptions: { queries: { retry: false } } });

const wrapperFor = (client: QueryClient) =>
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  };

describe("Maya story updates", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(useWorkspacePath).mockReturnValue({
      workspaceSlug: WORKSPACE_SLUG,
    } as ReturnType<typeof useWorkspacePath>);
    jest.mocked(useAnalytics).mockReturnValue({
      analytics: { track: jest.fn() },
    } as unknown as ReturnType<typeof useAnalytics>);
    jest.mocked(useTerminology).mockReturnValue({
      getTermDisplay: () => "story",
    } as ReturnType<typeof useTerminology>);
  });

  it.each([
    ["statusId", "started"],
    ["priority", "High"],
  ] as const)(
    "updates %s immediately and refreshes Maya queries",
    async (field, value) => {
      const client = createClient();
      const serverStory = { ...initialStory, [field]: value };
      client.setQueryData(recentKey, initialStory);
      client.setQueryData(workKey, grouped(initialStory));
      const recentFetch = jest.fn(async () => serverStory);
      const workFetch = jest.fn(async () => grouped(serverStory));
      const recentObserver = new QueryObserver(client, {
        queryKey: recentKey,
        queryFn: recentFetch,
        staleTime: Infinity,
      });
      const workObserver = new QueryObserver(client, {
        queryKey: workKey,
        queryFn: workFetch,
        staleTime: Infinity,
      });
      const unsubscribeRecent = recentObserver.subscribe(() => undefined);
      const unsubscribeWork = workObserver.subscribe(() => undefined);
      let completeUpdate: (() => void) | undefined;
      jest.mocked(updateStoryAction).mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            completeUpdate = () => {
              resolve({ data: null });
            };
          }),
      );
      const { result } = renderHook(() => useUpdateStoryMutation(), {
        wrapper: wrapperFor(client),
      });

      act(() => {
        result.current.mutate({
          storyId: STORY_ID,
          payload: { [field]: value },
        });
      });

      await waitFor(() => {
        expect(client.getQueryData<DetailedStory>(recentKey)?.[field]).toBe(
          value,
        );
        expect(
          client.getQueryData<GroupedStoriesResponse>(workKey)?.groups[0]
            ?.stories[0]?.[field],
        ).toBe(value);
      });

    act(() => {
      completeUpdate?.();
    });
      await waitFor(() => {
        expect(recentFetch).toHaveBeenCalledTimes(1);
        expect(workFetch).toHaveBeenCalledTimes(1);
        expect(result.current.isSuccess).toBe(true);
      });
      unsubscribeRecent();
      unsubscribeWork();
    },
  );

  it("rolls back the Maya row when the update fails", async () => {
    const client = createClient();
    client.setQueryData(recentKey, initialStory);
    const observer = new QueryObserver(client, {
      queryKey: recentKey,
      queryFn: async () => initialStory,
      staleTime: Infinity,
    });
    const unsubscribe = observer.subscribe(() => undefined);
    let failUpdate: (() => void) | undefined;
    jest.mocked(updateStoryAction).mockImplementationOnce(
      () =>
        new Promise((_, reject) => {
          failUpdate = () => {
            reject(new Error("Update failed"));
          };
        }),
    );
    const { result } = renderHook(() => useUpdateStoryMutation(), {
      wrapper: wrapperFor(client),
    });

    act(() => {
      result.current.mutate({
        storyId: STORY_ID,
        payload: { statusId: "started" },
      });
    });
    await waitFor(() => {
      expect(client.getQueryData<DetailedStory>(recentKey)?.statusId).toBe(
        "started",
      );
    });
    act(() => {
      failUpdate?.();
    });
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
      expect(client.getQueryData<DetailedStory>(recentKey)?.statusId).toBe(
        "backlog",
      );
    });
    expect(toast.error).toHaveBeenCalled();
    unsubscribe();
  });
});
