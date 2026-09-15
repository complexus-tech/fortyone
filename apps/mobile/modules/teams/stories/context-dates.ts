import { endOfDay, format, isValid, parseISO } from "date-fns";

function calendarDate(value: string | null | undefined) {
  if (!value || !/^\d{4}-\d{2}-\d{2}(T|$)/.test(value)) return null;
  const date = parseISO(value.slice(0, 10));
  return isValid(date) ? date : null;
}

export function formatContextDates(
  start: string | null | undefined,
  end: string | null | undefined,
) {
  const startDate = calendarDate(start);
  const endDate = calendarDate(end);
  if (startDate && endDate)
    return `${format(startDate, "MMM d")} – ${format(endDate, "MMM d")}`;
  if (endDate) return `Due ${format(endDate, "MMM d")}`;
  if (startDate) return `Starts ${format(startDate, "MMM d")}`;
  return "No dates set";
}

export function getSprintTiming(
  start: string | null | undefined,
  end: string | null | undefined,
  now = new Date(),
) {
  const startDate = calendarDate(start);
  const endDate = calendarDate(end);
  if (!startDate || !endDate || endDate < startDate) return null;
  if (startDate > now) return "Upcoming";
  if (endOfDay(endDate) < now) return "Completed";
  return "In progress";
}
