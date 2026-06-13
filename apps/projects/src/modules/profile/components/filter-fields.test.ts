/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import {
  hasActiveProfileStoriesFilters,
  PROFILE_HIDDEN_FILTER_FIELDS,
} from "./filter-fields";

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

describe("profile filter fields", () => {
  it("hides redundant person filters while leaving story filters available", () => {
    expect(PROFILE_HIDDEN_FILTER_FIELDS).toEqual([
      "assigneeIds",
      "reporterIds",
      "assignedToMe",
      "createdByMe",
      "hasNoAssignee",
    ]);
  });

  it("does not treat hidden person filters as active profile toolbar filters", () => {
    expect(
      hasActiveProfileStoriesFilters({
        ...baseFilters,
        assigneeIds: ["user-1"],
      }),
    ).toBe(false);
  });

  it("treats visible story filters as active profile toolbar filters", () => {
    expect(
      hasActiveProfileStoriesFilters({
        ...baseFilters,
        priorities: ["High"],
      }),
    ).toBe(true);
  });
});
