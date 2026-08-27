/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { DetailedStory } from "@/modules/story/types";
import { getBulkStoryCalendarImpact } from "./story-calendar-impact";

type StoryCalendarState = Pick<
  DetailedStory,
  | "autoSchedulingEnabled"
  | "autoSchedulingStatus"
  | "endDate"
  | "estimatedDurationMinutes"
>;

const calendarStory = (
  overrides: Partial<StoryCalendarState> = {},
): StoryCalendarState => ({
  autoSchedulingEnabled: true,
  autoSchedulingStatus: "planning",
  endDate: "2026-09-04",
  estimatedDurationMinutes: 60,
  ...overrides,
});

describe("getBulkStoryCalendarImpact", () => {
  it("reports planning without promising that calendar time is reserved", () => {
    const impact = getBulkStoryCalendarImpact([
      calendarStory(),
      calendarStory(),
    ]);

    expect(impact).toBe(
      "Calendar scheduling is on for all 2 stories. Maya is planning focus time for all 2 stories; no reservation is confirmed yet.",
    );
    expect(impact).not.toContain("will reserve");
  });

  it("reports that Maya-assigned stories still need a calendar owner", () => {
    const impact = getBulkStoryCalendarImpact([
      calendarStory({ autoSchedulingStatus: "needs_owner" }),
      calendarStory({ autoSchedulingStatus: "needs_owner" }),
    ]);

    expect(impact).toBe(
      "Calendar scheduling is on for all 2 stories. 2 stories still need a calendar owner before Maya can plan focus time; no reservation is confirmed for them.",
    );
    expect(impact).not.toContain("will reserve");
  });

  it("separates confirmed, planning, blocked, and disabled calendar states", () => {
    const impact = getBulkStoryCalendarImpact([
      calendarStory({ autoSchedulingStatus: "scheduled" }),
      calendarStory(),
      calendarStory({ autoSchedulingStatus: "cannot_fit" }),
      calendarStory({
        autoSchedulingEnabled: false,
        autoSchedulingStatus: "off",
      }),
    ]);

    expect(impact).toContain(
      "Calendar scheduling is on for 3 stories and off for 1 story.",
    );
    expect(impact).toContain("Focus time is reserved for 1 story.");
    expect(impact).toContain(
      "Maya is planning focus time for 1 story; no reservation is confirmed yet.",
    );
    expect(impact).toContain(
      "Maya could not fit 1 story before its delivery date.",
    );
  });
});
