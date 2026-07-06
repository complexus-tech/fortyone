/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getVisibleStoriesFilterButtonFields } from "./stories-filter-button-options";

describe("getVisibleStoriesFilterButtonFields", () => {
  it("removes hidden fields from the filter popover options", () => {
    expect(
      getVisibleStoriesFilterButtonFields({
        hasRouteTeam: false,
        hiddenFields: ["assigneeIds", "reporterIds", "hasNoAssignee"],
      }),
    ).not.toEqual(
      expect.arrayContaining(["assigneeIds", "reporterIds", "hasNoAssignee"]),
    );
  });

  it("hides team filters when the current route is already team-scoped", () => {
    expect(
      getVisibleStoriesFilterButtonFields({
        hasRouteTeam: true,
        hiddenFields: [],
      }),
    ).not.toContain("teamIds");
  });
});
