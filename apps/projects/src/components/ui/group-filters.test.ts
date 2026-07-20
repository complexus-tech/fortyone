/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { GroupedStoriesResponse } from "@/modules/stories/types";
import { groupFilters } from "./group-filters";

describe("groupFilters", () => {
  it("preserves content, exclusion, and date filters for group pagination", () => {
    const meta = {
      filters: {
        titleContains: "checkout copy",
        excludedStatusIds: ["status-1"],
        startDateAfter: "2026-07-20",
        deadlineNot: "2026-07-31",
        excludedObjectiveId: "objective-1",
        hasAssignee: true,
        hasBlockedBy: true,
      },
      groupBy: "status",
      orderBy: "created",
      orderDirection: "asc",
      totalGroups: 1,
    } satisfies GroupedStoriesResponse["meta"];

    expect(groupFilters(meta)).toMatchObject({
      titleContains: "checkout copy",
      excludedStatusIds: ["status-1"],
      startDateAfter: "2026-07-20",
      deadlineNot: "2026-07-31",
      excludedObjectiveId: "objective-1",
      hasAssignee: true,
      hasBlockedBy: true,
      orderDirection: "asc",
    });
  });
});
