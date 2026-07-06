/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getStoriesFilterStorageKey,
  mergeStoriesFilterDefaults,
} from "./stories-filter-storage";

describe("stories filter storage", () => {
  it("scopes filter storage by pathname", () => {
    expect(getStoriesFilterStorageKey("/workspace/teams/team-1/stories")).toBe(
      "stories:filters:/workspace/teams/team-1/stories",
    );
  });

  it("preserves stored filters while backfilling defaults", () => {
    expect(
      mergeStoriesFilterDefaults({
        estimateValues: [1, 5],
        labelIds: ["label-1"],
      }),
    ).toMatchObject({
      estimateValues: [1, 5],
      labelIds: ["label-1"],
      assignedToMe: false,
      createdByMe: false,
      statusIds: null,
    });
  });
});
