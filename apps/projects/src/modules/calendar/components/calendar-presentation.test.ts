/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { CalendarEventSummary } from "@/lib/queries/calendar/types";
import {
  calendarEventOverlapsDay,
  getBusyWindowTitle,
  overlapsDay,
} from "./calendar-presentation";

const createEvent = (
  patch: Partial<CalendarEventSummary> = {},
): CalendarEventSummary => ({
  endAt: "2026-08-17T00:00:00.000Z",
  id: "event-1",
  isAllDay: true,
  isPrivate: false,
  provider: "google",
  startAt: "2026-08-15T00:00:00.000Z",
  ...patch,
});

describe("calendar presentation helpers", () => {
  it("treats all-day event end dates as exclusive when filtering a calendar day", () => {
    const event = createEvent({
      endDate: "2026-08-17",
      startDate: "2026-08-15",
    });

    expect(calendarEventOverlapsDay(event, new Date(2026, 7, 15))).toBe(true);
    expect(calendarEventOverlapsDay(event, new Date(2026, 7, 16))).toBe(true);
    expect(calendarEventOverlapsDay(event, new Date(2026, 7, 17))).toBe(false);
  });

  it("keeps timed items that cross midnight visible in each overlapping day", () => {
    const item = {
      endAt: new Date(2026, 7, 16, 1).toISOString(),
      startAt: new Date(2026, 7, 15, 23).toISOString(),
    };

    expect(overlapsDay(item, new Date(2026, 7, 15))).toBe(true);
    expect(overlapsDay(item, new Date(2026, 7, 16))).toBe(true);
  });

  it("keeps private availability opaque while retaining public busy titles", () => {
    expect(
      getBusyWindowTitle({
        createdAt: "2026-08-15T00:00:00.000Z",
        endAt: "2026-08-15T10:00:00.000Z",
        id: "private",
        isPrivate: true,
        provider: "google",
        startAt: "2026-08-15T09:00:00.000Z",
        status: "busy",
        title: "Personal appointment",
        updatedAt: "2026-08-15T00:00:00.000Z",
      }),
    ).toBe("Busy");
    expect(
      getBusyWindowTitle({
        createdAt: "2026-08-15T00:00:00.000Z",
        endAt: "2026-08-15T10:00:00.000Z",
        id: "public",
        isPrivate: false,
        provider: "google",
        startAt: "2026-08-15T09:00:00.000Z",
        status: "busy",
        title: "  Planning  ",
        updatedAt: "2026-08-15T00:00:00.000Z",
      }),
    ).toBe("Planning");
  });
});
