import type { ReactNode } from "react";
import type { InfiniteData } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  QueryObserver,
} from "@tanstack/react-query";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type {
  DetailedStory,
  GroupedStoriesResponse,
  GroupStoriesResponse,
  Story,
} from "@/modules/stories/types";
import { updateStoryAction } from "../actions/update-story";
import { useUpdateStoryMutation } from "./update-mutation";

jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(),
  useTerminology: jest.fn(),
  useWorkspacePath: jest.fn(),
}));
jest.mock("../actions/update-story", () => ({ updateStoryAction: jest.fn() }));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

const WORKSPACE_SLUG = "forty-one";
const mockUpdateStoryAction = jest.mocked(updateStoryAction);

const createStory = (id: string): Story =>
  ({ id, statusId: "started", subStories: [] }) as unknown as Story;

const wrapperFor = (queryClient: QueryClient) =>
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };

describe("nested story updates", () => {
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

  it.each(["success", "error"] as const)(
    "updates nested statuses before the request completes and handles %s",
    async (outcome) => {
      const queryClient = new QueryClient({
        defaultOptions: {
          queries: { retry: false },
          mutations: { retry: false },
        },
      });
      // The parent may already have moved when completion is confirmed for
      // its children. A child patch must leave that parent's group intact.
      const parent = {
        ...createStory("parent"),
        statusId: "completed",
        subStories: [createStory("child"), createStory("sibling")],
      };
      const unrelatedStory = createStory("unrelated");
      const stories = [parent, unrelatedStory];
      const groupedKey = storyKeys.teamGrouped(WORKSPACE_SLUG, "team-1", {
        groupBy: "status",
      });
      const infiniteKey = storyKeys.groupStories(WORKSPACE_SLUG, "completed", {
        groupBy: "status",
        groupKey: "completed",
      });
      const listKey = storyKeys.list(WORKSPACE_SLUG, "filtered");
      const detailKey = storyKeys.detail(WORKSPACE_SLUG, parent.id);
      queryClient.setQueryData(groupedKey, {
        groups: [{ key: "completed", stories, loadedCount: 2, totalCount: 5 }],
        meta: { groupBy: "status" },
      });
      queryClient.setQueryData(infiniteKey, {
        pages: [{ groupKey: "completed", stories }],
        pageParams: [1],
      });
      queryClient.setQueryData(listKey, stories);
      queryClient.setQueryData(detailKey, parent);
      const unsubscribes = [groupedKey, infiniteKey, listKey, detailKey].map(
        (queryKey) =>
          new QueryObserver(queryClient, {
            queryKey,
            queryFn: () =>
              new Promise<unknown>(() => {
                // Keep invalidation from replacing the optimistic cache state.
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
      ];
      let finishUpdate:
        | ((response: Awaited<ReturnType<typeof updateStoryAction>>) => void)
        | undefined;
      mockUpdateStoryAction.mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finishUpdate = resolve;
          }),
      );
      const { result } = renderHook(() => useUpdateStoryMutation(), {
        wrapper: wrapperFor(queryClient),
      });

      act(() => {
        result.current.mutate({
          storyId: "child",
          payload: { statusId: "completed" },
        });
      });
      await waitFor(() => {
        expect(mockUpdateStoryAction).toHaveBeenCalledTimes(1);
      });

      expect(result.current.isPending).toBe(true);
      for (const cachedParent of readParents()) {
        expect(cachedParent?.statusId).toBe("completed");
        expect(cachedParent?.subStories[0].statusId).toBe("completed");
        expect(cachedParent?.subStories[1]).toBe(parent.subStories[1]);
      }
      expect(
        queryClient.getQueryData<GroupedStoriesResponse>(groupedKey)?.groups[0],
      ).toEqual(
        expect.objectContaining({
          key: "completed",
          loadedCount: 2,
          totalCount: 5,
        }),
      );
      expect(queryClient.getQueryData<Story[]>(listKey)?.[1]).toBe(
        unrelatedStory,
      );

      act(() => {
        finishUpdate?.(
          outcome === "success"
            ? { data: null }
            : { data: null, error: { message: "Status update rejected" } },
        );
      });
      await waitFor(() => {
        expect(result.current.status).toBe(outcome);
      });
      for (const cachedParent of readParents()) {
        expect(cachedParent?.statusId).toBe("completed");
        expect(
          cachedParent?.subStories.map(({ statusId }) => statusId),
        ).toEqual([outcome === "success" ? "completed" : "started", "started"]);
      }
      for (const unsubscribe of unsubscribes) unsubscribe();
    },
  );
});
