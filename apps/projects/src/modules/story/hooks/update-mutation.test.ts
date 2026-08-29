/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { InfiniteData } from "@tanstack/react-query";
import { QueryClient } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import type {
  GroupStoriesResponse,
  GroupStoryParams,
  Story,
} from "@/modules/stories/types";
import { updateInfiniteQuery } from "./update-mutation";

jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(),
  useTerminology: jest.fn(),
  useWorkspacePath: jest.fn(),
}));

jest.mock("../actions/update-story", () => ({
  updateStoryAction: jest.fn(),
}));

const createStory = (
  id: string,
  statusId: string,
  overrides: Partial<Story> = {},
): Story => ({
  archivedAt: null,
  assignee: null,
  assigneeId: null,
  autoSchedulingEnabled: false,
  autoSchedulingLocked: false,
  autoSchedulingReason: null,
  autoSchedulingStatus: "off",
  autoSchedulingUpdatedAt: null,
  collaboratorCount: 0,
  completedAt: null,
  createdAt: "2026-08-22T08:00:00.000Z",
  deletedAt: null,
  endDate: null,
  epicId: null,
  estimateLabel: null,
  estimateScheme: "points",
  estimateValue: null,
  estimatedDurationMinutes: null,
  id,
  keyResultId: null,
  labels: null,
  minimumFocusBlockMinutes: null,
  objective: null,
  objectiveId: null,
  priority: "Medium",
  reporterId: "user-1",
  sequenceId: 1,
  sprint: null,
  sprintId: null,
  startDate: null,
  statusId,
  subStories: [],
  team: { code: "ENG", id: "team-1", name: "Engineering" },
  teamId: "team-1",
  title: `Story ${id}`,
  updatedAt: "2026-08-22T08:00:00.000Z",
  workspaceId: "workspace-1",
  ...overrides,
});

const createPage = (
  groupKey: string,
  stories: Story[],
  page = 1,
): GroupStoriesResponse => ({
  filters: {},
  groupKey,
  orderBy: "created",
  orderDirection: "desc",
  pagination: {
    hasMore: false,
    nextPage: 0,
    page,
    pageSize: 15,
  },
  stories,
});

const createInfiniteData = (
  pages: GroupStoriesResponse[],
): InfiniteData<GroupStoriesResponse> => ({
  pageParams: pages.map(({ pagination }) => pagination.page),
  pages,
});

const createParams = (groupKey: string): GroupStoryParams => ({
  groupBy: "status",
  groupKey,
  orderBy: "created",
  orderDirection: "desc",
});

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: { structuralSharing: false },
    },
  });

describe("updateInfiniteQuery", () => {
  it("keeps an unrelated infinite query and its pages referentially stable", () => {
    const queryClient = createQueryClient();
    const params = createParams("development");
    const queryKey = storyKeys.groupStories(
      "workspace-1",
      params.groupKey,
      params,
    );
    const firstPage = createPage("development", [
      createStory("story-1", "development"),
    ]);
    const secondPage = createPage(
      "development",
      [createStory("story-2", "development")],
      2,
    );
    const data = createInfiniteData([firstPage, secondPage]);
    queryClient.setQueryData(queryKey, data);

    updateInfiniteQuery(queryClient, queryKey, "missing-story", {
      statusId: "qa",
    });

    const result = queryClient.getQueryData(queryKey);
    expect(result).toBe(data);
    expect((result as InfiniteData<GroupStoriesResponse>).pages[0]).toBe(
      firstPage,
    );
    expect((result as InfiniteData<GroupStoriesResponse>).pages[1]).toBe(
      secondPage,
    );
  });

  it("changes only the source page and target first page when moving a story", () => {
    const queryClient = createQueryClient();
    const sourceParams = createParams("development");
    const targetParams = createParams("qa");
    const sourceKey = storyKeys.groupStories(
      "workspace-1",
      sourceParams.groupKey,
      sourceParams,
    );
    const targetKey = storyKeys.groupStories(
      "workspace-1",
      targetParams.groupKey,
      targetParams,
    );
    const sourceFirst = createStory("source-first", "development");
    const sourceBefore = createStory("source-before", "development");
    const movedStory = createStory("story-moving", "development");
    const sourceAfter = createStory("source-after", "development");
    const targetFirst = createStory("target-first", "qa");
    const targetLast = createStory("target-last", "qa");
    const sourceFirstPage = createPage("development", [sourceFirst]);
    const sourceSecondPage = createPage(
      "development",
      [sourceBefore, movedStory, sourceAfter],
      2,
    );
    const targetFirstPage = createPage("qa", [targetFirst]);
    const targetSecondPage = createPage("qa", [targetLast], 2);
    const sourceData = createInfiniteData([sourceFirstPage, sourceSecondPage]);
    const targetData = createInfiniteData([targetFirstPage, targetSecondPage]);
    queryClient.setQueryData(sourceKey, sourceData);
    queryClient.setQueryData(targetKey, targetData);

    updateInfiniteQuery(queryClient, sourceKey, movedStory.id, {
      statusId: "qa",
    });

    const nextSource =
      queryClient.getQueryData<InfiniteData<GroupStoriesResponse>>(sourceKey)!;
    const nextTarget =
      queryClient.getQueryData<InfiniteData<GroupStoriesResponse>>(targetKey)!;
    expect(nextSource).not.toBe(sourceData);
    expect(nextSource.pages[0]).toBe(sourceFirstPage);
    expect(nextSource.pages[1]).not.toBe(sourceSecondPage);
    expect(nextSource.pages[1]?.stories).toEqual([sourceBefore, sourceAfter]);
    expect(nextSource.pages[1]?.stories[0]).toBe(sourceBefore);
    expect(nextSource.pages[1]?.stories[1]).toBe(sourceAfter);
    expect(nextTarget).not.toBe(targetData);
    expect(nextTarget.pages[0]).not.toBe(targetFirstPage);
    expect(nextTarget.pages[0]?.stories.map(({ id }) => id)).toEqual([
      movedStory.id,
      targetFirst.id,
    ]);
    expect(nextTarget.pages[0]?.stories[0]).toEqual(
      expect.objectContaining({ id: movedStory.id, statusId: "qa" }),
    );
    expect(nextTarget.pages[0]?.stories[1]).toBe(targetFirst);
    expect(nextTarget.pages[1]).toBe(targetSecondPage);
  });

  it("preserves page and story order for an in-group patch", () => {
    const queryClient = createQueryClient();
    const params = createParams("development");
    const queryKey = storyKeys.groupStories(
      "workspace-1",
      params.groupKey,
      params,
    );
    const firstStory = createStory("story-first", "development");
    const patchedStory = createStory("story-patched", "development");
    const lastStory = createStory("story-last", "development");
    const firstPage = createPage("development", [firstStory]);
    const secondPage = createPage("development", [patchedStory, lastStory], 2);
    const data = createInfiniteData([firstPage, secondPage]);
    queryClient.setQueryData(queryKey, data);

    updateInfiniteQuery(queryClient, queryKey, patchedStory.id, {
      title: "Patched title",
    });

    const result =
      queryClient.getQueryData<InfiniteData<GroupStoriesResponse>>(queryKey)!;
    expect(result).not.toBe(data);
    expect(result.pages[0]).toBe(firstPage);
    expect(result.pages[1]).not.toBe(secondPage);
    expect(result.pages[1]?.stories.map(({ id }) => id)).toEqual([
      patchedStory.id,
      lastStory.id,
    ]);
    expect(result.pages[1]?.stories[0]).toEqual(
      expect.objectContaining({ id: patchedStory.id, title: "Patched title" }),
    );
    expect(result.pages[1]?.stories[1]).toBe(lastStory);
  });
});
