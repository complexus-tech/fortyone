/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { ApiError } from "api-client";
import { getActivities } from "@/lib/queries/activities/get-activities";
import { getGroupedStories } from "@/modules/stories/public/queries";
import { getStory } from "@/modules/story/queries/get-story";
import { useRecentStoryHistory } from "@/shared/story/use-recent-story-history";
import type { DetailedStory, StoryActivity } from "@/shared/story/types";
import type { MayaWorkTab } from "../utils/recent-work";
import { useRecentWork } from "./use-recent-work";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user-a" } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace-a" }),
}));
jest.mock("@/lib/queries/activities/get-activities", () => ({
  getActivities: jest.fn(),
}));
jest.mock("@/modules/stories/public/queries", () => ({
  getGroupedStories: jest.fn(),
}));
jest.mock("@/modules/story/queries/get-story", () => ({ getStory: jest.fn() }));
jest.mock("@/shared/story/use-recent-story-history", () => ({
  useRecentStoryHistory: jest.fn(),
}));
jest.mock("api-client", () => ({
  ApiError: class ApiError extends Error {
    constructor(
      message: string,
      public status: number,
    ) {
      super(message);
    }
  },
}));

const story = (id: string) =>
  ({
    id,
    title: `Task ${id}`,
    deletedAt: null,
    archivedAt: null,
  }) as DetailedStory;
const activity = (storyId: string) =>
  ({
    id: `activity-${storyId}`,
    storyId,
    createdAt: "2026-09-05T08:00:00Z",
    field: "title",
  }) as StoryActivity;
const renderRecentWork = (tab: MayaWorkTab = "all") => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return renderHook(({ tab }: { tab: MayaWorkTab }) => useRecentWork(tab), {
    wrapper,
    initialProps: { tab },
  });
};

describe("Maya recent work", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(useRecentStoryHistory).mockReturnValue([]);
    jest.mocked(getActivities).mockResolvedValue([]);
    jest
      .mocked(getGroupedStories)
      .mockResolvedValue({ groups: [] } as unknown as Awaited<
        ReturnType<typeof getGroupedStories>
      >);
  });

  it("loads a task once when it appears in both visits and activity", async () => {
    jest
      .mocked(useRecentStoryHistory)
      .mockReturnValue([{ storyId: "a", visitedAt: "2026-09-05T09:00:00Z" }]);
    jest.mocked(getActivities).mockResolvedValue([activity("a")]);
    jest.mocked(getStory).mockResolvedValue(story("a"));
    const { result } = renderRecentWork();
    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
    expect(result.current.stories[0].title).toBe("Task a");
    expect(getGroupedStories).not.toHaveBeenCalled();
    expect(getStory).toHaveBeenCalledTimes(1);
  });

  it.each(["deleted", "archived", "missing", "private"])(
    "omits %s tasks from the list",
    async (unavailableId) => {
      jest
        .mocked(getActivities)
        .mockResolvedValue([unavailableId, "available"].map(activity));
      jest.mocked(getStory).mockImplementation(async (id) => {
        if (id === "missing") return null;
        if (id === "private") throw new ApiError("Forbidden", 403, null);
        return {
          ...story(id),
          deletedAt: id === "deleted" ? "2026-09-05" : null,
          archivedAt: id === "archived" ? "2026-09-05" : null,
        };
      });
      const { result } = renderRecentWork();
      await waitFor(() => {
        expect(result.current.isPending).toBe(false);
      });
      expect(result.current.stories.map((story) => story.id)).toEqual([
        "available",
      ]);
      expect(result.current.isError).toBe(false);
    },
  );

  it("surfaces fetch failures without discarding successfully loaded tasks", async () => {
    jest
      .mocked(getActivities)
      .mockResolvedValue([activity("available"), activity("failed")]);
    jest.mocked(getStory).mockImplementation(async (id) => {
      if (id === "failed") throw new Error("Service unavailable");
      return story(id);
    });
    const { result } = renderRecentWork();
    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
    expect(result.current.isError).toBe(true);
    expect(result.current.stories).toHaveLength(1);
  });

  it("shows assigned open tasks when there is no recent history", async () => {
    jest.mocked(getGroupedStories).mockResolvedValue({
      groups: [{ stories: [story("assigned")] }],
    } as unknown as Awaited<ReturnType<typeof getGroupedStories>>);
    const { result } = renderRecentWork();
    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
    expect(result.current.stories.map(({ id }) => id)).toEqual(["assigned"]);
    expect(result.current.isAssignedFallback).toBe(true);
    expect(getGroupedStories).toHaveBeenCalledWith(
      expect.objectContaining({ workspaceSlug: "workspace-a" }),
      expect.objectContaining({
        assignedToMe: true,
        categories: ["backlog", "unstarted", "started", "paused"],
        storiesPerGroup: 3,
      }),
    );
  });

  it("uses assigned tasks after all saved tasks become unavailable", async () => {
    jest
      .mocked(useRecentStoryHistory)
      .mockReturnValue([
        { storyId: "missing", visitedAt: "2026-09-05T09:00:00Z" },
      ]);
    jest.mocked(getStory).mockResolvedValue(null);
    const { result } = renderRecentWork();
    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
    expect(getGroupedStories).toHaveBeenCalledTimes(1);
    expect(result.current.stories).toEqual([]);
    expect(result.current.isError).toBe(false);
  });

  it("does not treat failed history requests as empty history", async () => {
    jest
      .mocked(getActivities)
      .mockRejectedValue(new Error("Service unavailable"));
    const { result } = renderRecentWork();
    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
    expect(result.current.isError).toBe(true);
    expect(getGroupedStories).not.toHaveBeenCalled();
  });

  it("surfaces an assigned-task failure and allows retrying it", async () => {
    jest
      .mocked(getGroupedStories)
      .mockRejectedValueOnce(new Error("Service unavailable"));
    const { result } = renderRecentWork();
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    result.current.retry();
    await waitFor(() => {
      expect(result.current.isError).toBe(false);
    });
    expect(result.current.stories).toEqual([]);
    expect(getGroupedStories).toHaveBeenCalledTimes(2);
  });
  it.each([
    ["assigned", "assignedToMe"],
    ["created", "createdByMe"],
  ] as const)(
    "loads the latest three matching tasks for %s",
    async (tab, filter) => {
      jest.mocked(getGroupedStories).mockResolvedValue({
        groups: [{ stories: [story("a"), story("b"), story("c"), story("d")] }],
      } as unknown as Awaited<ReturnType<typeof getGroupedStories>>);
      const { result } = renderRecentWork(tab);
      await waitFor(() => {
        expect(result.current.isPending).toBe(false);
      });
      expect(result.current.stories.map(({ id }) => id)).toEqual([
        "a",
        "b",
        "c",
      ]);
      expect(getGroupedStories).toHaveBeenCalledWith(
        expect.objectContaining({ workspaceSlug: "workspace-a" }),
        expect.objectContaining({
          [filter]: true,
          storiesPerGroup: 3,
          orderBy: "updated",
        }),
      );
      expect(getActivities).not.toHaveBeenCalled();
      expect(getStory).not.toHaveBeenCalled();
    },
  );

  it("loads separate results when switching between assigned and created", async () => {
    jest.mocked(getGroupedStories).mockImplementation(
      async (_, params) =>
        ({
          groups: [
            { stories: [story(params.createdByMe ? "created" : "assigned")] },
          ],
        }) as unknown as Awaited<ReturnType<typeof getGroupedStories>>,
    );
    const { result, rerender } = renderRecentWork("assigned");
    await waitFor(() => {
      expect(result.current.stories[0].id).toBe("assigned");
    });
    rerender({ tab: "created" });
    await waitFor(() => {
      expect(result.current.stories[0].id).toBe("created");
    });
  });
});
