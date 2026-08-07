import {
  addDays,
  addMonths,
  format,
  isSameMonth,
  isSameYear,
  startOfDay,
  startOfMonth,
  startOfWeek,
} from "date-fns";

export type CalendarView = "day" | "week" | "month";

export type CalendarViewRange = {
  start: Date;
  end: Date;
};

const WEEK_STARTS_ON = 1 as const;
const MINIMUM_MONTH_VIEW_DAYS = 35;

const getWeekRange = (cursor: Date): CalendarViewRange => {
  const start = startOfWeek(cursor, { weekStartsOn: WEEK_STARTS_ON });
  return { start, end: addDays(start, 7) };
};

const getMonthRange = (cursor: Date): CalendarViewRange => {
  const monthStart = startOfMonth(cursor);
  const start = startOfWeek(monthStart, { weekStartsOn: WEEK_STARTS_ON });
  const nextMonthStart = addMonths(monthStart, 1);
  const nextMonthWeekStart = startOfWeek(nextMonthStart, {
    weekStartsOn: WEEK_STARTS_ON,
  });
  const end =
    nextMonthWeekStart < nextMonthStart
      ? addDays(nextMonthWeekStart, 7)
      : nextMonthWeekStart;
  const minimumEnd = addDays(start, MINIMUM_MONTH_VIEW_DAYS);

  return { start, end: end < minimumEnd ? minimumEnd : end };
};

export const getCalendarViewRange = (
  cursor: Date,
  view: CalendarView,
): CalendarViewRange => {
  if (view === "day") {
    const start = startOfDay(cursor);
    return { start, end: addDays(start, 1) };
  }

  if (view === "week") {
    return getWeekRange(cursor);
  }

  return getMonthRange(cursor);
};

export const getCalendarViewDays = (cursor: Date, view: CalendarView) => {
  const { start, end } = getCalendarViewRange(cursor, view);
  const days: Date[] = [];

  for (let day = start; day < end; day = addDays(day, 1)) {
    days.push(day);
  }

  return days;
};

export const getCalendarViewTitle = (cursor: Date, view: CalendarView) => {
  if (view === "day") {
    return format(cursor, "MMMM d, yyyy");
  }

  if (view === "month") {
    return format(cursor, "MMMM yyyy");
  }

  const { start } = getWeekRange(cursor);
  const end = addDays(start, 6);
  if (isSameMonth(start, end)) {
    return format(start, "MMMM yyyy");
  }
  if (isSameYear(start, end)) {
    return `${format(start, "MMMM")} – ${format(end, "MMMM yyyy")}`;
  }
  return `${format(start, "MMMM yyyy")} – ${format(end, "MMMM yyyy")}`;
};

export const moveCalendarCursor = (
  cursor: Date,
  view: CalendarView,
  amount: number,
) => {
  if (view === "day") {
    return addDays(startOfDay(cursor), amount);
  }

  if (view === "week") {
    return addDays(cursor, amount * 7);
  }

  return addMonths(startOfMonth(cursor), amount);
};
