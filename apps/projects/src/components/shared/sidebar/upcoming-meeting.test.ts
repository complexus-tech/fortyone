/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { CalendarEventSummary } from "@/lib/queries/calendar/types";
import {
  getUpcomingMeeting,
  getUpcomingMeetings,
  isMeetingDismissed,
} from "./upcoming-meeting";

const now = Date.parse("2026-08-08T10:00:00.000Z");

const createEvent = (
  overrides: Partial<CalendarEventSummary> = {},
): CalendarEventSummary => ({
  id: "meeting-1",
  provider: "google",
  title: "Weekly planning",
  meetingUrl: "https://meet.google.com/abc-defg-hij",
  startAt: "2026-08-08T10:15:00.000Z",
  endAt: "2026-08-08T11:00:00.000Z",
  isAllDay: false,
  isPrivate: false,
  ...overrides,
});

describe("getUpcomingMeeting", () => {
  it("returns a meeting starting within the next 15 minutes", () => {
    expect(getUpcomingMeeting([createEvent()], now)).toMatchObject({
      minutesUntilStart: 15,
      status: "upcoming",
      meetingUrl: "https://meet.google.com/abc-defg-hij",
    });
  });

  it("prioritizes an in-progress meeting over the next meeting", () => {
    const inProgress = createEvent({
      id: "meeting-in-progress",
      startAt: "2026-08-08T09:45:00.000Z",
      endAt: "2026-08-08T10:30:00.000Z",
    });

    expect(getUpcomingMeeting([createEvent(), inProgress], now)).toMatchObject({
      event: { id: "meeting-in-progress" },
      minutesUntilStart: 0,
      status: "in-progress",
    });
  });

  it("selects the earliest upcoming meeting", () => {
    const laterMeeting = createEvent({
      id: "meeting-later",
      startAt: "2026-08-08T10:14:00.000Z",
    });
    const earlierMeeting = createEvent({
      id: "meeting-earlier",
      startAt: "2026-08-08T10:05:00.000Z",
    });

    expect(
      getUpcomingMeeting([laterMeeting, earlierMeeting], now)?.event.id,
    ).toBe("meeting-earlier");
  });

  it("returns simultaneous meetings so the sidebar can page through them", () => {
    const meetings = getUpcomingMeetings(
      [
        createEvent({
          id: "meeting-later",
          startAt: "2026-08-08T10:12:00.000Z",
        }),
        createEvent({ id: "meeting-now", startAt: "2026-08-08T09:55:00.000Z" }),
        createEvent({
          id: "meeting-next",
          startAt: "2026-08-08T10:05:00.000Z",
        }),
      ],
      now,
    );

    expect(meetings.map(({ event }) => event.id)).toEqual([
      "meeting-now",
      "meeting-next",
      "meeting-later",
    ]);
  });

  it("matches only the dismissed meeting cookie", () => {
    expect(
      isMeetingDismissed(
        "meeting-1",
        "theme=dark; fortyone_meeting_dismissed_meeting-1=1",
      ),
    ).toBe(true);
    expect(
      isMeetingDismissed(
        "meeting-2",
        "theme=dark; fortyone_meeting_dismissed_meeting-1=1",
      ),
    ).toBe(false);
  });

  it.each([
    createEvent({ startAt: "2026-08-08T10:16:00.000Z" }),
    createEvent({
      startAt: "2026-08-08T09:00:00.000Z",
      endAt: "2026-08-08T10:00:00.000Z",
    }),
    createEvent({ isAllDay: true }),
    createEvent({ isPrivate: true }),
    createEvent({ meetingUrl: undefined }),
    createEvent({ meetingUrl: "http://meet.example.com/unsafe" }),
  ])("ignores events that should not trigger a reminder", (event) => {
    expect(getUpcomingMeeting([event], now)).toBeNull();
  });
});
