import type { CalendarEventSummary } from "@/lib/queries/calendar/types";

const UPCOMING_MEETING_WINDOW_MS = 15 * 60 * 1000;
const MEETING_DISMISSAL_COOKIE_PREFIX = "fortyone_meeting_dismissed_";

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
): UpcomingMeeting | null => getUpcomingMeetings(events, now).at(0) ?? null;

export const getUpcomingMeetings = (
  events: CalendarEventSummary[],
  now: number,
): UpcomingMeeting[] => {
  const meetings: UpcomingMeeting[] = [];
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
      meetings.push({
        event,
        meetingUrl,
        minutesUntilStart: 0,
        status: "in-progress",
      });
      continue;
    }

    meetings.push({
      event,
      meetingUrl,
      minutesUntilStart: Math.ceil((startAt - now) / (60 * 1000)),
      status: "upcoming",
    });
  }

  return meetings.sort((left, right) => {
    if (left.status !== right.status) {
      return left.status === "in-progress" ? -1 : 1;
    }
    return Date.parse(left.event.startAt) - Date.parse(right.event.startAt);
  });
};

const getMeetingDismissalCookieName = (meetingId: string) =>
  `${MEETING_DISMISSAL_COOKIE_PREFIX}${meetingId}`;

export const isMeetingDismissed = (
  meetingId: string,
  cookieHeader = typeof document === "undefined" ? "" : document.cookie,
) => {
  const cookieName = `${getMeetingDismissalCookieName(meetingId)}=`;
  return cookieHeader
    .split(";")
    .some((cookie) => cookie.trim().startsWith(cookieName));
};

export const dismissMeetingUntilEnd = (
  meeting: UpcomingMeeting,
  options?: { isSecure?: boolean },
) => {
  if (typeof document === "undefined") return;

  const endAt = new Date(meeting.event.endAt);
  if (!Number.isFinite(endAt.getTime())) return;

  const secure =
    options?.isSecure ??
    (typeof window !== "undefined" && window.location.protocol === "https:");
  const maxAgeSeconds = Math.max(
    0,
    Math.ceil((endAt.getTime() - Date.now()) / 1000),
  );
  document.cookie = [
    `${getMeetingDismissalCookieName(meeting.event.id)}=1`,
    "Path=/",
    `Expires=${endAt.toUTCString()}`,
    `Max-Age=${maxAgeSeconds}`,
    "SameSite=Lax",
    secure ? "Secure" : "",
  ]
    .filter(Boolean)
    .join("; ");
};
