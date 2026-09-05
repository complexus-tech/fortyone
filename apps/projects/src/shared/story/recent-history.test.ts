/* global describe, expect, it -- Jest globals are provided by the projects test runner. */
import {
  addStoryVisit,
  getRecentStoryHistoryKey,
  parseStoryVisits,
  RECENT_STORY_HISTORY_LIMIT,
} from "./recent-history";

const storyId = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
const visitedAt = "2026-09-05T08:00:00.000Z";

describe("recent story history", () => {
  it("isolates history by account and workspace", () => {
    const key = getRecentStoryHistoryKey("alice", "design");
    expect(key).not.toBe(getRecentStoryHistoryKey("bob", "design"));
    expect(key).not.toBe(getRecentStoryHistoryKey("alice", "engineering"));
  });

  it("recovers from corrupt storage and validates saved identifiers and dates", () => {
    expect(parseStoryVisits("invalid")).toEqual([]);
    expect(parseStoryVisits('{"items":[]}')).toEqual([]);
    expect(
      parseStoryVisits(
        JSON.stringify([
          { storyId, visitedAt, title: "Do not persist task content" },
          { storyId: "../other-path", visitedAt },
          { storyId, visitedAt: "invalid-date" },
          null,
        ]),
      ),
    ).toEqual([{ storyId, visitedAt }]);
  });

  it("moves a reopened task to the front without duplicating it or exceeding the bound", () => {
    const visits = Array.from(
      { length: RECENT_STORY_HISTORY_LIMIT },
      (_, index) => ({
        storyId: `task-${index}`,
        visitedAt,
      }),
    );
    const reopened = {
      storyId: "task-10",
      visitedAt: "2026-09-05T09:00:00.000Z",
    };
    const result = addStoryVisit(visits, reopened);
    expect(result[0]).toEqual(reopened);
    expect(
      result.filter((visit) => visit.storyId === reopened.storyId),
    ).toHaveLength(1);
    expect(addStoryVisit(result, { storyId, visitedAt })).toHaveLength(
      RECENT_STORY_HISTORY_LIMIT,
    );
  });
});
