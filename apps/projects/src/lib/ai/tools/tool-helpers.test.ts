/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  filterActivityTimeline,
  paginateRecords,
  requireToolConfirmation,
  resolvePaginationInput,
} from "./tool-helpers";

const activities = [
  {
    id: "activity-1",
    userId: "user-1",
    field: "priority",
    createdAt: "2026-06-10T10:00:00.000Z",
  },
  {
    id: "activity-2",
    userId: "user-2",
    field: "estimate",
    createdAt: "2026-06-12T10:00:00.000Z",
  },
  {
    id: "activity-3",
    userId: "user-1",
    field: "status",
    createdAt: "2026-06-13T10:00:00.000Z",
  },
];

describe("AI tool helpers", () => {
  it("clamps page and page size to safe defaults", () => {
    expect(resolvePaginationInput({ page: 0, pageSize: 500 })).toEqual({
      page: 1,
      pageSize: 100,
    });
    expect(resolvePaginationInput({ page: 3, pageSize: 15 })).toEqual({
      page: 3,
      pageSize: 15,
    });
  });

  it("paginates records with total and hasMore metadata", () => {
    expect(paginateRecords(["a", "b", "c"], { page: 2, pageSize: 2 })).toEqual({
      records: ["c"],
      pagination: {
        page: 2,
        pageSize: 2,
        totalCount: 3,
        totalPages: 2,
        hasMore: false,
        nextPage: null,
      },
    });
  });

  it("returns a consistent confirmation response", () => {
    expect(requireToolConfirmation("delete all pending requests")).toEqual({
      success: false,
      needsConfirmation: true,
      message: "Please confirm before I delete all pending requests.",
    });
  });

  it("filters activities by user, field, and date", () => {
    expect(
      filterActivityTimeline(activities, {
        userId: "user-1",
        fields: ["status", "priority"],
        since: "2026-06-11T00:00:00.000Z",
      }),
    ).toEqual([activities[2]]);
  });
});
