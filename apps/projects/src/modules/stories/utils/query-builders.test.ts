/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  buildGroupedStoriesQuery,
  buildGroupStoriesQuery,
} from "./query-builders";

describe("buildGroupedStoriesQuery", () => {
  it("serializes the selected order direction", () => {
    expect(
      buildGroupedStoriesQuery({
        groupBy: "status",
        orderBy: "created",
        orderDirection: "asc",
      }),
    ).toContain("orderDirection=asc");
  });

  it("omits blank and empty filters from current-user story queries", () => {
    expect(
      buildGroupedStoriesQuery({
        groupBy: "status",
        assignedToMe: true,
        assigneeIds: [],
        labelIds: [],
        objectiveId: "",
        parentId: "   ",
        statusIds: ["", "   "],
        teamIds: [],
        titleContains: "\t",
      }),
    ).toBe("?groupBy=status&assignedToMe=true");
  });

  it("normalizes optional strings without removing valid false values", () => {
    expect(
      buildGroupStoriesQuery({
        groupBy: "status",
        groupKey: "  status-1  ",
        includeArchived: false,
        labelIds: ["", "  label-1  "],
        titleContains: "  launch  ",
      }),
    ).toBe(
      "?groupBy=status&groupKey=status-1&includeArchived=false&labelIds=label-1&titleContains=launch",
    );
  });
});
