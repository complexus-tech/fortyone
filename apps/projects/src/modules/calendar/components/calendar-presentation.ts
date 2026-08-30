import { addDays, format, isSameDay, startOfDay } from "date-fns";
import type {
  CalendarBusyWindow,
  CalendarEventSummary,
} from "@/lib/queries/calendar/types";
import { parseCalendarDate } from "./calendar-layout";
import type { CalendarItem, CalendarStoryOption } from "./calendar-types";

export const DEFAULT_VISIBLE_START_HOUR = 8;
export const DEFAULT_VISIBLE_END_HOUR = 24;
export const HOUR_HEIGHT = 52;
export const TIMED_BLOCK_VERTICAL_GAP = 6;
export const TIMED_BLOCK_VERTICAL_INSET = TIMED_BLOCK_VERTICAL_GAP / 2;
export const TWO_LINE_TITLE_MINIMUM_HEIGHT = HOUR_HEIGHT * 1.5;
export const TIME_RAIL_WIDTH = 8;
export const CALENDAR_HISTORY_DAYS = 7;
export const CALENDAR_LOOKAHEAD_DAYS = 90;
export const SCHEDULED_TASK_BACKGROUND_CLASS =
  "bg-surface-muted dark:bg-surface-prominent/65";
export const SCHEDULED_TASK_HOVER_BACKGROUND_CLASS =
  "hover:bg-accent dark:hover:bg-surface-prominent/70";
export const SCHEDULED_STORY_STATUS_CLASS =
  "border-[var(--calendar-story-border)] bg-[var(--calendar-story-background)]";
export const SCHEDULED_STORY_STATUS_HOVER_CLASS =
  "hover:bg-[var(--calendar-story-hover)]";
export const COMPLETED_CALENDAR_BLOCK_PATTERN =
  "repeating-linear-gradient(135deg, transparent 0, transparent 5px, rgba(100, 116, 139, 0.12) 5px, rgba(100, 116, 139, 0.12) 8px)";

export const toDateTimeInputValue = (value: Date | string) =>
  format(new Date(value), "yyyy-MM-dd'T'HH:mm");

export const toClockLabel = (value: Date, includePeriod: boolean) => {
  const timePattern = value.getMinutes() === 0 ? "h" : "h:mm";
  return format(
    value,
    includePeriod ? `${timePattern}a` : timePattern,
  ).toLowerCase();
};

export const toTimeLabel = (startAt: string, endAt: string) => {
  const start = new Date(startAt);
  const end = new Date(endAt);
  const isSamePeriod = format(start, "a") === format(end, "a");
  return `${toClockLabel(start, !isSamePeriod)} – ${toClockLabel(end, true)}`;
};

export const toResizeEndLabel = (startAt: Date, endAt: Date) =>
  isSameDay(startAt, endAt)
    ? toClockLabel(endAt, true)
    : `${format(endAt, "EEE")} ${toClockLabel(endAt, true)}`;

export const toMoveTimeLabel = (startAt: Date, endAt: Date) =>
  isSameDay(startAt, endAt)
    ? toTimeLabel(startAt.toISOString(), endAt.toISOString())
    : `${format(startAt, "EEE")} ${toClockLabel(startAt, true)} – ${format(endAt, "EEE")} ${toClockLabel(endAt, true)}`;

export const getUtcOffsetLabel = (date: Date) => {
  const offsetMinutes = -date.getTimezoneOffset();
  const sign = offsetMinutes >= 0 ? "+" : "-";
  const absoluteOffset = Math.abs(offsetMinutes);
  const hours = Math.floor(absoluteOffset / 60)
    .toString()
    .padStart(2, "0");
  const minutes = absoluteOffset % 60;
  return `GMT${sign}${hours}${minutes ? `:${minutes.toString().padStart(2, "0")}` : ""}`;
};

export const getLocalTimeZoneName = () => {
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  const location = timeZone.split("/").at(-1);
  return location?.replaceAll("_", " ") || "Local time";
};

export const roundToNextHalfHour = (date: Date) => {
  const next = new Date(date);
  const minutes = next.getMinutes();
  next.setMinutes(minutes < 30 ? 30 : 60, 0, 0);
  return next;
};

export const overlapsDay = (
  item: Pick<CalendarItem, "startAt" | "endAt">,
  day: Date,
) => {
  const dayStart = startOfDay(day);
  const dayEnd = addDays(dayStart, 1);
  return new Date(item.startAt) < dayEnd && new Date(item.endAt) > dayStart;
};

export const calendarEventOverlapsDay = (
  event: CalendarEventSummary,
  day: Date,
) => {
  if (event.isAllDay) {
    const startDate = parseCalendarDate(event.startDate);
    const endDate = parseCalendarDate(event.endDate);
    if (startDate && endDate) {
      const dayStart = startOfDay(day);
      const dayEnd = addDays(dayStart, 1);
      return startDate < dayEnd && endDate > dayStart;
    }
  }
  return overlapsDay(event, day);
};

export const getStoryCode = (story: CalendarStoryOption) =>
  story.team?.code
    ? `${story.team.code}-${story.sequenceId}`
    : `#${story.sequenceId}`;

export const getBusyWindowTitle = (window: CalendarBusyWindow) => {
  if (window.isPrivate) {
    return "Busy";
  }
  return window.title?.trim() || "Busy";
};
