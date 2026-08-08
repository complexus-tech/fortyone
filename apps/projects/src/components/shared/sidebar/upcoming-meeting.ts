import type { CalendarEventSummary } from "@/lib/queries/calendar/types";

const UPCOMING_MEETING_WINDOW_MS = 15 * 60 * 1000;

export type UpcomingMeeting = {
  event: CalendarEventSummary;
  meetingUrl: string;
  minutesUntilStart: number;
  status: "in-progress" | "upcoming";
};

const getSafeMeetingUrl = (value?: string) => {
  if (!value) return null;

  try {
    const url = new URL(value);
    return url.protocol === "https:" ? url.toString() : null;
  } catch {
    return null;
  }
};

export const getUpcomingMeeting = (
  events: CalendarEventSummary[],
  now: number,
): UpcomingMeeting | null => {
  let inProgress: UpcomingMeeting | null = null;
  let upcoming: UpcomingMeeting | null = null;
  const windowEnd = now + UPCOMING_MEETING_WINDOW_MS;

  for (const event of events) {
    if (event.isAllDay || event.isPrivate) continue;

    const meetingUrl = getSafeMeetingUrl(event.meetingUrl);
    const startAt = Date.parse(event.startAt);
    const endAt = Date.parse(event.endAt);

    if (
      !meetingUrl ||
      !Number.isFinite(startAt) ||
      !Number.isFinite(endAt) ||
      endAt <= startAt ||
      endAt <= now ||
      startAt > windowEnd
    ) {
      continue;
    }

    if (startAt <= now) {
      if (!inProgress || startAt < Date.parse(inProgress.event.startAt)) {
        inProgress = {
          event,
          meetingUrl,
          minutesUntilStart: 0,
          status: "in-progress",
        };
      }
      continue;
    }

    if (!upcoming || startAt < Date.parse(upcoming.event.startAt)) {
      upcoming = {
        event,
        meetingUrl,
        minutesUntilStart: Math.ceil((startAt - now) / (60 * 1000)),
        status: "upcoming",
      };
    }
  }

  return inProgress ?? upcoming;
};
