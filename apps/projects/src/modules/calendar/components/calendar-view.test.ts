/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getCalendarViewDays,
  getCalendarViewRange,
  getCalendarViewTitle,
  moveCalendarCursor,
  normalizeCalendarView,
} from "./calendar-view";

const expectLocalDate = (
  date: Date,
  year: number,
  month: number,
  day: number,
) => {
  expect(date.getFullYear()).toBe(year);
  expect(date.getMonth()).toBe(month);
  expect(date.getDate()).toBe(day);
  expect(date.getHours()).toBe(0);
  expect(date.getMinutes()).toBe(0);
};

describe("calendar view", () => {
  const augustSeventh = new Date(2026, 7, 7, 16, 45);

  it("normalizes persisted calendar views", () => {
    expect(normalizeCalendarView("day")).toBe("day");
    expect(normalizeCalendarView("week")).toBe("week");
    expect(normalizeCalendarView("month")).toBe("month");
    expect(normalizeCalendarView("agenda")).toBe("week");
    expect(normalizeCalendarView(null)).toBe("week");
  });

  it("builds a half-open local day range", () => {
    const range = getCalendarViewRange(augustSeventh, "day");
    const days = getCalendarViewDays(augustSeventh, "day");

    expectLocalDate(range.start, 2026, 7, 7);
    expectLocalDate(range.end, 2026, 7, 8);
    expect(days).toHaveLength(1);
    expectLocalDate(days[0], 2026, 7, 7);
  });

  it("builds a Monday-start, half-open week range", () => {
    const range = getCalendarViewRange(augustSeventh, "week");
    const days = getCalendarViewDays(augustSeventh, "week");

    expectLocalDate(range.start, 2026, 7, 3);
    expectLocalDate(range.end, 2026, 7, 10);
    expect(days).toHaveLength(7);
    expectLocalDate(days[0], 2026, 7, 3);
    expectLocalDate(days[6], 2026, 7, 9);
  });

  it("includes six Monday-start weeks in the August 2026 month view", () => {
    const range = getCalendarViewRange(augustSeventh, "month");
    const days = getCalendarViewDays(augustSeventh, "month");

    expectLocalDate(range.start, 2026, 6, 27);
    expectLocalDate(range.end, 2026, 8, 7);
    expect(days).toHaveLength(42);
    expectLocalDate(days[0], 2026, 6, 27);
    expectLocalDate(days[41], 2026, 8, 6);
  });

  it("formats day, week, and month headings", () => {
    expect(getCalendarViewTitle(augustSeventh, "day")).toBe("August 7, 2026");
    expect(getCalendarViewTitle(augustSeventh, "week")).toBe("August 2026");
    expect(getCalendarViewTitle(augustSeventh, "month")).toBe("August 2026");
    expect(getCalendarViewTitle(new Date(2026, 7, 31), "week")).toBe(
      "August – September 2026",
    );
    expect(getCalendarViewTitle(new Date(2026, 11, 31), "week")).toBe(
      "December 2026 – January 2027",
    );
  });

  it("moves month views across year boundaries without end-of-month drift", () => {
    const february = moveCalendarCursor(new Date(2026, 0, 31, 14), "month", 1);
    const march = moveCalendarCursor(february, "month", 1);
    const january = moveCalendarCursor(new Date(2026, 11, 31, 14), "month", 1);
    const december = moveCalendarCursor(new Date(2026, 0, 31, 14), "month", -1);

    expectLocalDate(february, 2026, 1, 1);
    expectLocalDate(march, 2026, 2, 1);
    expectLocalDate(january, 2027, 0, 1);
    expectLocalDate(december, 2025, 11, 1);
  });

  it("keeps the selected weekday when moving between weeks", () => {
    const nextWeek = moveCalendarCursor(augustSeventh, "week", 1);
    const previousWeek = moveCalendarCursor(augustSeventh, "week", -1);

    expect(nextWeek.getFullYear()).toBe(2026);
    expect(nextWeek.getMonth()).toBe(7);
    expect(nextWeek.getDate()).toBe(14);
    expect(previousWeek.getFullYear()).toBe(2026);
    expect(previousWeek.getMonth()).toBe(6);
    expect(previousWeek.getDate()).toBe(31);
  });

  it.each([
    [new Date(2026, 2, 8, 15), 2026, 2, 9],
    [new Date(2026, 10, 1, 15), 2026, 10, 2],
  ])(
    "moves local dates by calendar day around DST boundaries (%s)",
    (cursor, year, month, day) => {
      expectLocalDate(moveCalendarCursor(cursor, "day", 1), year, month, day);
    },
  );
});
