/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { StoriesFilter } from "./stories-filter-types";
import {
  getGroupedStoryFilterParams,
  getScopedStoriesFilterTeamId,
} from "./stories-filter-query";
import { getActiveStoriesFilterCount } from "./stories-filter-utils";

const baseFilters = {
  statusIds: null,
  assigneeIds: null,
  reporterIds: null,
  priorities: null,
  teamIds: null,
  sprintIds: null,
  labelIds: null,
  estimateValues: null,
  parentId: null,
  objectiveId: null,
  epicId: null,
  keyResultId: null,
  contentContains: null,
  startDate: null,
  endDate: null,
  hasNoAssignee: null,
  assignedToMe: false,
  createdByMe: false,
} as unknown as StoriesFilter;

describe("story filter query mapping", () => {
  it("maps content and labels into grouped story query params", () => {
    const filters = {
      ...baseFilters,
      contentContains: "checkout copy",
      labelIds: ["label-1", "label-2"],
    } as unknown as StoriesFilter;

    expect(getGroupedStoryFilterParams(filters)).toMatchObject({
      titleContains: "checkout copy",
      labelIds: ["label-1", "label-2"],
    });
  });

  it("maps estimate values into grouped story query params", () => {
    const filters = {
      ...baseFilters,
      estimateValues: [1, 5, 8],
    } as unknown as StoriesFilter;

    expect(getGroupedStoryFilterParams(filters)).toMatchObject({
      estimateValues: [1, 5, 8],
    });
  });

  it("maps negated content and collection operators to exclusion params", () => {
    const filters = {
      ...baseFilters,
      contentContains: "deprecated",
      statusIds: ["status-1"],
      assigneeIds: ["user-1"],
      operators: {
        contentContains: "doesNotContain",
        statusIds: "isNotAnyOf",
        assigneeIds: "isNotAnyOf",
      },
    } as StoriesFilter;

    expect(getGroupedStoryFilterParams(filters)).toMatchObject({
      titleNotContains: "deprecated",
      excludedStatusIds: ["status-1"],
      excludedAssigneeIds: ["user-1"],
    });
    expect(getGroupedStoryFilterParams(filters)).toMatchObject({
      titleContains: undefined,
      statusIds: undefined,
      assigneeIds: undefined,
    });
  });

  it("maps date comparison, objective, and assignee presence operators", () => {
    const filters = {
      ...baseFilters,
      startDate: "2026-08-01",
      endDate: "2026-08-31",
      objectiveId: "objective-1",
      hasNoAssignee: true,
      operators: {
        startDate: "isOnOrAfter",
        endDate: "isNot",
        objectiveId: "isNot",
        hasNoAssignee: "isNotEmpty",
      },
    } as StoriesFilter;

    expect(getGroupedStoryFilterParams(filters)).toMatchObject({
      startDateAfter: "2026-08-01",
      startDateBefore: undefined,
      deadlineAfter: undefined,
      deadlineBefore: undefined,
      deadlineNot: "2026-08-31",
      excludedObjectiveId: "objective-1",
      hasNoAssignee: undefined,
      hasAssignee: true,
    });
  });

  it("counts content and labels as active filters", () => {
    const filters = {
      ...baseFilters,
      contentContains: "analytics",
      labelIds: ["label-1"],
    } as unknown as StoriesFilter;

    expect(getActiveStoriesFilterCount(filters)).toBe(2);
  });

  it("counts estimate values as an active filter", () => {
    const filters = {
      ...baseFilters,
      estimateValues: [3],
    } as unknown as StoriesFilter;

    expect(getActiveStoriesFilterCount(filters)).toBe(1);
  });
});

describe("getScopedStoriesFilterTeamId", () => {
  it("uses the route team before selected team filters", () => {
    expect(getScopedStoriesFilterTeamId("team-route", ["team-filter"])).toBe(
      "team-route",
    );
  });

  it("uses the selected team filter when exactly one team is selected", () => {
    expect(getScopedStoriesFilterTeamId(undefined, ["team-filter"])).toBe(
      "team-filter",
    );
  });

  it("does not infer a team when the filter has no selected team or multiple selected teams", () => {
    expect(getScopedStoriesFilterTeamId(undefined, null)).toBeUndefined();
    expect(
      getScopedStoriesFilterTeamId(undefined, ["team-a", "team-b"]),
    ).toBeUndefined();
  });

  it("does not scope dependent filters to an excluded team", () => {
    expect(
      getScopedStoriesFilterTeamId(undefined, ["team-a"], "isNotAnyOf"),
    ).toBeUndefined();
  });
});
