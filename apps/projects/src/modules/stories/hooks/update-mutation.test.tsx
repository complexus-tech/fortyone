import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  QueryObserver,
} from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import type { DetailedStory, Story } from "../types";
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
});
