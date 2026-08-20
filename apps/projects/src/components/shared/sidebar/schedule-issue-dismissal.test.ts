/* global beforeEach, describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { CalendarScheduleIssue } from "@/lib/queries/calendar/types";
import {
  dismissScheduleIssue,
  getScheduleIssueDismissalToken,
  isScheduleIssueDismissed,
  SCHEDULE_ISSUE_DISMISSAL_TTL_MS,
} from "./schedule-issue-dismissal";

const DISMISSED_AT = Date.parse("2026-08-08T10:00:00.000Z");
const WORKSPACE_SLUG = "acme";

const createIssue = (
  overrides: Partial<CalendarScheduleIssue> = {},
): CalendarScheduleIssue => ({
  autoSchedulingStatus: "cannot_fit",
  estimatedDurationMinutes: 90,
  remainingDurationMinutes: 30,
  scheduledDurationMinutes: 60,
  storyCode: "ENG-42",
  storyId: "story-1",
  storyTitle: "Prepare the launch brief",
  teamCode: "ENG",
  teamId: "team-1",
  teamName: "Engineering",
  updatedAt: "2026-08-08T09:55:00.000Z",
  ...overrides,
});

describe("schedule issue dismissal", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("persists the current issue dismissal for 24 hours", () => {
    const issue = createIssue();
    const dismissedUntil = dismissScheduleIssue(
      issue,
      WORKSPACE_SLUG,
      DISMISSED_AT,
    );

    expect(dismissedUntil).toBe(DISMISSED_AT + SCHEDULE_ISSUE_DISMISSAL_TTL_MS);
    expect(
      isScheduleIssueDismissed(
        issue,
        WORKSPACE_SLUG,
        dismissedUntil - 1,
        new Map(),
      ),
    ).toBe(true);
    expect(
      isScheduleIssueDismissed(
        issue,
        WORKSPACE_SLUG,
        dismissedUntil,
        new Map(),
      ),
    ).toBe(false);
  });

  it("does not hide a newer version of the scheduling issue", () => {
    const issue = createIssue();
    dismissScheduleIssue(issue, WORKSPACE_SLUG, DISMISSED_AT);

    expect(
      isScheduleIssueDismissed(
        createIssue({ updatedAt: "2026-08-08T09:58:00.000Z" }),
        WORKSPACE_SLUG,
        DISMISSED_AT + 1,
        new Map(),
      ),
    ).toBe(false);
  });

  it("supports an in-memory dismissal when storage is unavailable", () => {
    const issue = createIssue();
    const dismissedUntil = DISMISSED_AT + SCHEDULE_ISSUE_DISMISSAL_TTL_MS;
    const sessionDismissals = new Map([
      [getScheduleIssueDismissalToken(issue), dismissedUntil],
    ]);

    expect(
      isScheduleIssueDismissed(
        issue,
        WORKSPACE_SLUG,
        DISMISSED_AT + 1,
        sessionDismissals,
      ),
    ).toBe(true);
  });
});
