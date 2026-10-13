import type { ReactNode } from "react";
import type { InfiniteData } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  QueryObserver,
} from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import type {
  DetailedStory,
  GroupedStoriesResponse,
  GroupStoriesResponse,
  Story,
} from "../types";
import { storyKeys } from "../constants";
import { bulkUpdateAction } from "../actions/bulk-update-stories";
import { useBulkUpdateStoriesMutation } from "./update-mutation";

jest.mock("next/navigation", () => ({
  useParams: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
  },
}));

jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(),
  useTerminology: jest.fn(),
  useWorkspacePath: jest.fn(),
}));

jest.mock("../actions/bulk-update-stories", () => ({
  bulkUpdateAction: jest.fn(),
}));

const WORKSPACE_SLUG = "forty-one";
const PARENT_STORY_ID = "parent-story";
const analyticsTrack = jest.fn();
const mockBulkUpdateAction = jest.mocked(bulkUpdateAction);
const mockUseAnalytics = jest.mocked(useAnalytics);
const mockUseParams = jest.mocked(useParams);
const mockUseTerminology = jest.mocked(useTerminology);
const mockUseWorkspacePath = jest.mocked(useWorkspacePath);

const createStory = (id: string): Story =>
  ({
    id,
    statusId: "status-backlog",
  }) as Story;

const createParentStory = (): DetailedStory =>
  ({
    id: PARENT_STORY_ID,
    subStories: [createStory("story-1"), createStory("story-2")],
  }) as DetailedStory;

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
      queries: { retry: false },
    },
  });

const wrapperFor = (queryClient: QueryClient) =>
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };

const getSubStoryStatuses = (queryClient: QueryClient) =>
  queryClient
    .getQueryData<DetailedStory>(
      storyKeys.detail(WORKSPACE_SLUG, PARENT_STORY_ID),
    )
    ?.subStories.map(({ statusId }) => statusId);

describe("useBulkUpdateStoriesMutation", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseParams.mockReturnValue({ storyId: PARENT_STORY_ID });
    mockUseWorkspacePath.mockReturnValue({
      workspaceSlug: WORKSPACE_SLUG,
    } as ReturnType<typeof useWorkspacePath>);
    mockUseAnalytics.mockReturnValue({
      analytics: { track: analyticsTrack },
    } as unknown as ReturnType<typeof useAnalytics>);
    mockUseTerminology.mockReturnValue({
      getTermDisplay: (_term, options) =>
        options?.variant === "singular" ? "story" : "stories",
    } as ReturnType<typeof useTerminology>);
  });

  it.each(["success", "error"] as const)(
    "updates selected nested stories in all active cache shapes before a bulk %s",
    async (outcome) => {
      mockUseParams.mockReturnValue({});
      const queryClient = createQueryClient();
      const parent = {
        ...createStory(PARENT_STORY_ID),
        subStories: [createStory("story-1"), createStory("story-2")],
      };
      const unrelatedStory = createStory("unrelated-story");
      const stories = [parent, unrelatedStory];
      const groupedKey = storyKeys.teamGrouped(WORKSPACE_SLUG, "team-1", {
        groupBy: "status",
      });
      const infiniteKey = storyKeys.groupStories(
        WORKSPACE_SLUG,
        "status-backlog",
        { groupBy: "status" },
      );
      const listKey = storyKeys.list(WORKSPACE_SLUG, "filtered");
      const detailKey = storyKeys.detail(WORKSPACE_SLUG, PARENT_STORY_ID);
      const recentKey = [...detailKey, "maya-recent", "user-1"];
      queryClient.setQueryData(groupedKey, {
        groups: [{ key: "status-backlog", stories }],
        meta: {
          groupBy: "status",
          filters: {},
          orderBy: "created",
          orderDirection: "desc",
          totalGroups: 1,
        },
      });
      queryClient.setQueryData(infiniteKey, {
        pages: [{ stories }],
        pageParams: [1],
      });
      queryClient.setQueryData(listKey, stories);
      queryClient.setQueryData(detailKey, parent);
      queryClient.setQueryData(recentKey, parent);
      const unsubscribes = [
        groupedKey,
        infiniteKey,
        listKey,
        detailKey,
        recentKey,
      ].map((queryKey) =>
        new QueryObserver(queryClient, {
          queryKey,
          queryFn: () =>
            new Promise<unknown>(() => {
              // Keep invalidation from replacing the mutation's cache data.
            }),
          staleTime: Infinity,
        }).subscribe(() => undefined),
      );
      const readParents = () => [
        queryClient.getQueryData<GroupedStoriesResponse>(groupedKey)?.groups[0]
          .stories[0],
        queryClient.getQueryData<InfiniteData<GroupStoriesResponse>>(
          infiniteKey,
        )?.pages[0].stories[0],
        queryClient.getQueryData<Story[]>(listKey)?.[0],
        queryClient.getQueryData<DetailedStory>(detailKey),
        queryClient.getQueryData<DetailedStory>(recentKey),
      ];
      let finishUpdate:
        | ((response: Awaited<ReturnType<typeof bulkUpdateAction>>) => void)
        | undefined;
      mockBulkUpdateAction.mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finishUpdate = resolve;
          }),
      );
      const { result } = renderHook(() => useBulkUpdateStoriesMutation(), {
        wrapper: wrapperFor(queryClient),
      });

      act(() => {
        result.current.mutate({
          storyIds: ["story-1"],
          payload: { statusId: "status-started" },
        });
      });
      await waitFor(() => {
        expect(mockBulkUpdateAction).toHaveBeenCalledTimes(1);
      });

      expect(result.current.isPending).toBe(true);
      for (const cachedParent of readParents()) {
        expect(cachedParent?.statusId).toBe("status-backlog");
        expect(cachedParent?.subStories[0].statusId).toBe("status-started");
        expect(cachedParent?.subStories[1]).toBe(parent.subStories[1]);
      }
      expect(queryClient.getQueryData<Story[]>(listKey)?.[1]).toBe(
        unrelatedStory,
      );

      act(() => {
        finishUpdate?.(
          outcome === "success"
            ? {
                data: {
                  totalCount: 1,
                  succeededCount: 1,
                  failedCount: 0,
                  partial: false,
                  items: [{ storyId: "story-1", success: true }],
                },
              }
            : { data: null, error: { message: "Status update rejected" } },
        );
      });
      await waitFor(() => {
        expect(result.current.status).toBe(outcome);
      });

      for (const cachedParent of readParents()) {
        expect(cachedParent?.statusId).toBe("status-backlog");
        expect(
          cachedParent?.subStories.map(({ statusId }) => statusId),
        ).toEqual([
          outcome === "success" ? "status-started" : "status-backlog",
          "status-backlog",
        ]);
      }
      for (const unsubscribe of unsubscribes) unsubscribe();
    },
  );

  it("updates selected children even when their parent is selected too", async () => {
    const queryClient = createQueryClient();
    const parentKey = storyKeys.detail(WORKSPACE_SLUG, PARENT_STORY_ID);
    queryClient.setQueryData(parentKey, createParentStory());
    mockBulkUpdateAction.mockImplementationOnce(
      () =>
        new Promise(() => {
          // Observe both optimistic updates before the request completes.
        }),
    );
    const { result } = renderHook(() => useBulkUpdateStoriesMutation(), {
      wrapper: wrapperFor(queryClient),
    });

    act(() => {
      result.current.mutate({
        storyIds: [PARENT_STORY_ID, "story-1"],
        payload: { priority: "High" },
      });
    });
    await waitFor(() => {
      expect(mockBulkUpdateAction).toHaveBeenCalledTimes(1);
    });

    const parent = queryClient.getQueryData<DetailedStory>(parentKey);
    expect(parent?.priority).toBe("High");
    expect(parent?.subStories[0].priority).toBe("High");
    expect(parent?.subStories[1].priority).toBeUndefined();
  });

  it("turns HTTP-200 item failures into a rollback and item-aware error", async () => {
    const queryClient = createQueryClient();
    const parentStoryKey = storyKeys.detail(WORKSPACE_SLUG, PARENT_STORY_ID);
    queryClient.setQueryData(parentStoryKey, createParentStory());
    const activeParentObserver = new QueryObserver(queryClient, {
      queryKey: parentStoryKey,
      queryFn: () =>
        new Promise<DetailedStory>(() => {
          // Keep invalidation from refetching the original value so this test
          // observes the mutation rollback itself.
        }),
      staleTime: Infinity,
    });
    const unsubscribeParentObserver = activeParentObserver.subscribe(
      () => undefined,
    );
    mockBulkUpdateAction.mockResolvedValueOnce({
      data: {
        totalCount: 2,
        succeededCount: 1,
        failedCount: 1,
        partial: true,
        items: [
          { storyId: "story-1", success: true },
          {
            storyId: "story-2",
            success: false,
            error: "Status does not belong to this story's team.",
          },
        ],
      },
    });

    const { result } = renderHook(() => useBulkUpdateStoriesMutation(), {
      wrapper: wrapperFor(queryClient),
    });

    act(() => {
      result.current.mutate({
        storyIds: ["story-1", "story-2"],
        payload: { statusId: "status-started" },
      });
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(getSubStoryStatuses(queryClient)).toEqual([
      "status-backlog",
      "status-backlog",
    ]);
    expect(analyticsTrack).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(
      "Failed to update 1 of 2 stories",
      expect.objectContaining({
        description: "Status does not belong to this story's team.",
      }),
    );

    mockBulkUpdateAction.mockResolvedValueOnce({
      data: {
        totalCount: 1,
        succeededCount: 1,
        failedCount: 0,
        partial: false,
        items: [{ storyId: "story-2", success: true }],
      },
    });
    const toastOptions = jest.mocked(toast.error).mock.calls[0]?.[1];
    const retryAction = toastOptions?.action as
      | { onClick?: () => void }
      | undefined;

    act(() => {
      retryAction?.onClick?.();
    });

    await waitFor(() => {
      expect(mockBulkUpdateAction).toHaveBeenCalledTimes(2);
    });
    expect(mockBulkUpdateAction).toHaveBeenLastCalledWith(
      {
        storyIds: ["story-2"],
        updates: { statusId: "status-started" },
      },
      WORKSPACE_SLUG,
    );
    unsubscribeParentObserver();
  });

  it("preserves the top-level API error toast and full retry", async () => {
    const queryClient = createQueryClient();
    queryClient.setQueryData(
      storyKeys.detail(WORKSPACE_SLUG, PARENT_STORY_ID),
      createParentStory(),
    );
    mockBulkUpdateAction.mockResolvedValueOnce({
      data: null,
      error: { message: "The status update was rejected" },
    });

    const { result } = renderHook(() => useBulkUpdateStoriesMutation(), {
      wrapper: wrapperFor(queryClient),
    });

    act(() => {
      result.current.mutate({
        storyIds: ["story-1", "story-2"],
        payload: { statusId: "status-started" },
      });
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(getSubStoryStatuses(queryClient)).toEqual([
      "status-backlog",
      "status-backlog",
    ]);
    expect(toast.error).toHaveBeenCalledWith("Failed to update stories", {
      description: "The status update was rejected",
      action: expect.objectContaining({ label: "Retry" }),
    });
  });

  it("keeps the optimistic update and records analytics after full success", async () => {
    const queryClient = createQueryClient();
    queryClient.setQueryData(
      storyKeys.detail(WORKSPACE_SLUG, PARENT_STORY_ID),
      createParentStory(),
    );
    mockBulkUpdateAction.mockResolvedValueOnce({
      data: {
        totalCount: 2,
        succeededCount: 2,
        failedCount: 0,
        partial: false,
        items: [
          { storyId: "story-1", success: true },
          { storyId: "story-2", success: true },
        ],
      },
    });

    const { result } = renderHook(() => useBulkUpdateStoriesMutation(), {
      wrapper: wrapperFor(queryClient),
    });

    act(() => {
      result.current.mutate({
        storyIds: ["story-1", "story-2"],
        payload: { statusId: "status-started" },
      });
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(getSubStoryStatuses(queryClient)).toEqual([
      "status-started",
      "status-started",
    ]);
    expect(analyticsTrack).toHaveBeenCalledWith("stories_bulk_updated", {
      storyIds: ["story-1", "story-2"],
      count: 2,
      statusId: "status-started",
    });
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("updates selected Maya stories before the bulk request completes", async () => {
    const queryClient = createQueryClient();
    const recentKey = [
      ...storyKeys.detail(WORKSPACE_SLUG, "story-1"),
      "maya-recent",
      "user-1",
    ] as const;
    queryClient.setQueryData(recentKey, createStory("story-1"));
    const recentObserver = new QueryObserver(queryClient, {
      queryKey: recentKey,
      queryFn: () =>
        new Promise<Story>(() => {
          // Keep the optimistic value visible until this test completes.
        }),
      staleTime: Infinity,
    });
    const unsubscribe = recentObserver.subscribe(() => undefined);
    let completeUpdate: (() => void) | undefined;
    mockBulkUpdateAction.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          completeUpdate = () => {
            resolve({
              data: {
                totalCount: 1,
                succeededCount: 1,
                failedCount: 0,
                partial: false,
                items: [{ storyId: "story-1", success: true }],
              },
            });
          };
        }),
    );
    const { result } = renderHook(() => useBulkUpdateStoriesMutation(), {
      wrapper: wrapperFor(queryClient),
    });

    act(() => {
      result.current.mutate({
        storyIds: ["story-1"],
        payload: { statusId: "status-started" },
      });
    });
    await waitFor(() => {
      expect(queryClient.getQueryData<Story>(recentKey)?.statusId).toBe(
        "status-started",
      );
    });
    act(() => {
      completeUpdate?.();
    });
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
    unsubscribe();
  });
});
