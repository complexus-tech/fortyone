/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { DetailedStory, StoryGroup } from "../types";
import {
  computeTargetKey,
  moveStoryBetweenGroups,
  patchStories,
  updateStoryInGroups,
} from "./optimistic";

const story: DetailedStory = {
  archivedAt: null,
  assigneeId: null,
  autoSchedulingEnabled: false,
  autoSchedulingLocked: false,
  autoSchedulingReason: null,
  autoSchedulingStatus: "off",
  autoSchedulingUpdatedAt: null,
  associations: [],
  collaboratorCount: 0,
  collaboratorIds: [],
  collaborators: [],
  completedAt: null,
  createdAt: "2026-07-23T08:00:00.000Z",
  deletedAt: null,
  description: "",
  descriptionHTML: "",
  endDate: null,
  epicId: null,
  estimateLabel: null,
  estimateScheme: "points",
  estimateValue: null,
  estimatedDurationMinutes: null,
  minimumFocusBlockMinutes: null,
  id: "story-1",
  keyResultId: null,
  labels: null,
  objectiveId: null,
  parentId: "",
  priority: "Medium",
  reporterId: "user-1",
  sequenceId: 41,
  sprintId: null,
  startDate: null,
  statusId: "development",
  subStories: [],
  teamCode: "ENG",
  teamId: "team-1",
  title: "Ship notifications",
  updatedAt: "2026-07-23T08:00:00.000Z",
  isWatching: false,
  watcherCount: 0,
  watchers: [],
  watchingReason: null,
  workspaceId: "workspace-1",
};

const createGroup = ({
  key,
  stories = [],
  totalCount = stories.length,
}: {
  key: string;
  stories?: DetailedStory[];
  totalCount?: number;
}): StoryGroup => ({
  hasMore: totalCount > stories.length,
  key,
  loadedCount: stories.length,
  nextPage: 2,
  stories,
  totalCount,
});

const createStory = (
  id: string,
  overrides: Partial<DetailedStory> = {},
): DetailedStory => ({
  ...story,
  id,
  title: `Story ${id}`,
  ...overrides,
});

describe("computeTargetKey", () => {
  it("maps an explicit unassignment to the API's null group", () => {
    expect(computeTargetKey("assignee", { assigneeId: null })).toBe("null");
    expect(computeTargetKey("assignee", { title: "No group change" })).toBe(
      undefined,
    );
  });
});

describe("patchStories", () => {
  it("returns the original array when the story is absent", () => {
    const stories = [story];

    const result = patchStories(stories, "missing-story", {
      title: "Unreachable update",
    });

    expect(result).toBe(stories);
    expect(result[0]).toBe(story);
  });

  it("preserves order and untouched story references", () => {
    const firstStory = createStory("story-first");
    const lastStory = createStory("story-last");
    const stories = [firstStory, story, lastStory];

    const result = patchStories(stories, story.id, {
      title: "Ship polished notifications",
    });

    expect(result).not.toBe(stories);
    expect(result.map(({ id }) => id)).toEqual([
      firstStory.id,
      story.id,
      lastStory.id,
    ]);
    expect(result[0]).toBe(firstStory);
    expect(result[1]).not.toBe(story);
    expect(result[1]?.title).toBe("Ship polished notifications");
    expect(result[2]).toBe(lastStory);
  });
});

describe("moveStoryBetweenGroups", () => {
  it("reveals an empty target group by updating its optimistic counts", () => {
    const groups = [
      createGroup({ key: "development", stories: [story] }),
      createGroup({ key: "qa" }),
    ];

    const result = moveStoryBetweenGroups(groups, story.id, "qa", {
      statusId: "qa",
    });

    expect(result).toEqual([
      expect.objectContaining({
        key: "development",
        loadedCount: 0,
        stories: [],
        totalCount: 0,
      }),
      expect.objectContaining({
        key: "qa",
        loadedCount: 1,
        stories: [expect.objectContaining({ id: story.id, statusId: "qa" })],
        totalCount: 1,
      }),
    ]);
  });

  it("keeps group counts stable when updating a story within its group", () => {
    const groups = [
      createGroup({
        key: "development",
        stories: [story],
        totalCount: 3,
      }),
    ];

    const result = moveStoryBetweenGroups(groups, story.id, "development", {
      title: "Ship polished notifications",
    });

    expect(result[0]).toEqual(
      expect.objectContaining({
        loadedCount: 1,
        stories: [
          expect.objectContaining({
            id: story.id,
            title: "Ship polished notifications",
          }),
        ],
        totalCount: 3,
      }),
    );
  });

  it("returns the original groups when the story is absent", () => {
    const developmentStories = [story];
    const qaStories = [createStory("story-2", { statusId: "qa" })];
    const groups = [
      createGroup({ key: "development", stories: developmentStories }),
      createGroup({ key: "qa", stories: qaStories }),
    ];

    const result = moveStoryBetweenGroups(groups, "missing-story", "qa", {
      statusId: "qa",
    });

    expect(result).toBe(groups);
    expect(result[0]).toBe(groups[0]);
    expect(result[0]?.stories).toBe(developmentStories);
    expect(result[1]).toBe(groups[1]);
    expect(result[1]?.stories).toBe(qaStories);
  });

  it("changes only source and target groups while preserving story order", () => {
    const sourceFirst = createStory("story-source-first");
    const sourceLast = createStory("story-source-last");
    const targetStory = createStory("story-target", { statusId: "qa" });
    const untouchedStory = createStory("story-untouched", {
      statusId: "released",
    });
    const groups = [
      createGroup({
        key: "development",
        stories: [sourceFirst, story, sourceLast],
      }),
      createGroup({ key: "qa", stories: [targetStory] }),
      createGroup({ key: "released", stories: [untouchedStory] }),
    ];

    const result = moveStoryBetweenGroups(groups, story.id, "qa", {
      statusId: "qa",
    });

    expect(result).not.toBe(groups);
    expect(result[0]).not.toBe(groups[0]);
    expect(result[0]?.stories).toEqual([sourceFirst, sourceLast]);
    expect(result[0]?.stories[0]).toBe(sourceFirst);
    expect(result[0]?.stories[1]).toBe(sourceLast);
    expect(result[1]).not.toBe(groups[1]);
    expect(result[1]?.stories.map(({ id }) => id)).toEqual([
      story.id,
      targetStory.id,
    ]);
    expect(result[1]?.stories[1]).toBe(targetStory);
    expect(result[2]).toBe(groups[2]);
    expect(result[2]?.stories).toBe(groups[2]?.stories);
  });
});

describe("updateStoryInGroups", () => {
  it("patches time needed without removing the story from its active group", () => {
    const groups = [
      createGroup({
        key: "development",
        stories: [story],
        totalCount: 3,
      }),
    ];

    const result = updateStoryInGroups(groups, story.id, "status", {
      estimatedDurationMinutes: 90,
      minimumFocusBlockMinutes: 30,
    });

    expect(result[0]).toEqual(
      expect.objectContaining({
        loadedCount: 1,
        stories: [
          expect.objectContaining({
            id: story.id,
            estimatedDurationMinutes: 90,
            minimumFocusBlockMinutes: 30,
          }),
        ],
        totalCount: 3,
      }),
    );
  });

  it("patches auto-scheduling state without removing the story from its active group", () => {
    const groups = [
      createGroup({
        key: "development",
        stories: [story],
        totalCount: 3,
      }),
    ];

    const result = updateStoryInGroups(groups, story.id, "status", {
      autoSchedulingEnabled: true,
      autoSchedulingStatus: "needs_owner",
    });

    expect(result[0]).toEqual(
      expect.objectContaining({
        loadedCount: 1,
        stories: [
          expect.objectContaining({
            id: story.id,
            autoSchedulingEnabled: true,
            autoSchedulingStatus: "needs_owner",
          }),
        ],
        totalCount: 3,
      }),
    );
  });

  it("still removes an explicitly unassigned story from an assignee group", () => {
    const assignedStory = { ...story, assigneeId: "user-2" };
    const groups = [createGroup({ key: "user-2", stories: [assignedStory] })];

    const result = updateStoryInGroups(groups, story.id, "assignee", {
      assigneeId: null,
    });

    expect(result[0]).toEqual(
      expect.objectContaining({ loadedCount: 0, stories: [], totalCount: 0 }),
    );
  });

  it("moves an explicitly unassigned story into the null assignee group", () => {
    const assignedStory = { ...story, assigneeId: "user-2" };
    const groups = [
      createGroup({ key: "user-2", stories: [assignedStory] }),
      createGroup({ key: "null" }),
    ];

    const result = updateStoryInGroups(groups, story.id, "assignee", {
      assigneeId: null,
    });

    expect(result[0]?.stories).toEqual([]);
    expect(result[1]?.stories).toEqual([
      expect.objectContaining({ id: story.id, assigneeId: null }),
    ]);
  });

  it("returns the original groups when a non-grouping patch misses", () => {
    const developmentStories = [story];
    const groups = [
      createGroup({ key: "development", stories: developmentStories }),
    ];

    const result = updateStoryInGroups(groups, "missing-story", "status", {
      estimatedDurationMinutes: 90,
    });

    expect(result).toBe(groups);
    expect(result[0]).toBe(groups[0]);
    expect(result[0]?.stories).toBe(developmentStories);
  });

  it("preserves untouched group and story references when patching", () => {
    const neighbor = createStory("story-neighbor");
    const qaStory = createStory("story-qa", { statusId: "qa" });
    const groups = [
      createGroup({ key: "development", stories: [story, neighbor] }),
      createGroup({ key: "qa", stories: [qaStory] }),
    ];

    const result = updateStoryInGroups(groups, story.id, "status", {
      estimatedDurationMinutes: 90,
    });

    expect(result).not.toBe(groups);
    expect(result[0]).not.toBe(groups[0]);
    expect(result[0]?.stories).not.toBe(groups[0]?.stories);
    expect(result[0]?.stories[1]).toBe(neighbor);
    expect(result[1]).toBe(groups[1]);
    expect(result[1]?.stories).toBe(groups[1]?.stories);
  });
});
