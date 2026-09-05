/* global describe, expect, it -- Jest globals are provided by the projects test runner. */
import type { StoryActivity } from "@/shared/story/types";
import { getRecentStories } from "./recent-work";

const activity = (storyId: string, createdAt: string) =>
  ({ storyId, createdAt }) as StoryActivity;

describe("recent work ordering", () => {
  it("merges opened and worked-on tasks using the latest personal interaction", () => {
    const visits = [
      { storyId: "a", visitedAt: "2026-09-05T08:00:00Z" },
      { storyId: "b", visitedAt: "2026-09-05T10:00:00Z" },
    ];
    const activities = [
      activity("a", "2026-09-05T09:00:00Z"),
      activity("a", "2026-09-05T07:00:00Z"),
      activity("c", "2026-09-05T06:00:00Z"),
    ];
    expect(getRecentStories(visits, activities, 2)).toEqual([
      { storyId: "b", timestamp: "2026-09-05T10:00:00Z", action: "Opened" },
      { storyId: "a", timestamp: "2026-09-05T09:00:00Z", action: "Worked on" },
    ]);
    expect(visits[0].storyId).toBe("a");
    expect(activities).toHaveLength(3);
  });

  it("uses activity history before any browser visits exist", () => {
    expect(
      getRecentStories([], [activity("a", "2026-09-05T09:00:00Z")])[0].action,
    ).toBe("Worked on");
    expect(getRecentStories([], [])).toEqual([]);
  });

  it("shows the three most recent distinct tasks by default", () => {
    const activities = ["a", "a", "b", "c", "d", "e"].map((id, index) =>
      activity(id, `2026-09-05T09:0${5 - index}:00Z`),
    );
    expect(
      getRecentStories([], activities).map(({ storyId }) => storyId),
    ).toEqual(["a", "b", "c"]);
  });
});
