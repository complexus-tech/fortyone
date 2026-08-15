/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { DetailedStory } from "../../story/types";
import type { StoryGroup } from "../types";
import { moveStoryBetweenGroups, updateStoryInGroups } from "./optimistic";

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
});
