import type { StoriesFilter } from "./stories-filter-types";
import { getGroupedStoryFilterParams } from "./stories-filter-query";
import { getActiveStoriesFilterCount } from "./stories-filter-utils";

const baseFilters = {
  statusIds: null,
  assigneeIds: null,
  reporterIds: null,
  priorities: null,
  teamIds: null,
  sprintIds: null,
  labelIds: null,
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

  it("counts content and labels as active filters", () => {
    const filters = {
      ...baseFilters,
      contentContains: "analytics",
      labelIds: ["label-1"],
    } as unknown as StoriesFilter;

    expect(getActiveStoriesFilterCount(filters)).toBe(2);
  });
});
