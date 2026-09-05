/* global describe, expect, it -- Jest globals are provided by the projects test runner. */
import { getStoryAttentionFilters } from "./attention";

describe("work attention filters", () => {
  it("keeps overdue strictly before today, including at month boundaries", () => {
    expect(
      getStoryAttentionFilters("overdue", new Date(2024, 2, 1, 0, 15)),
    ).toEqual({
      assignedToMe: true,
      showSubStories: true,
      categories: ["backlog", "unstarted", "started", "paused"],
      deadlineAfter: undefined,
      deadlineBefore: "2024-02-29",
    });
  });
  it("uses the same local calendar date for both due-today boundaries", () => {
    expect(
      getStoryAttentionFilters("today", new Date(2026, 8, 5, 23, 59)),
    ).toMatchObject({
      deadlineAfter: "2026-09-05",
      deadlineBefore: "2026-09-05",
    });
  });
});
