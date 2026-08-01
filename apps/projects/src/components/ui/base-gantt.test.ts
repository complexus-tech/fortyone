/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  calculateGanttPosition,
  calculateTimelineDateFromPosition,
  calculateTimelineDatePosition,
  getGanttOffscreenDirection,
} from "./base-gantt-utils";

describe("BaseGantt timeline positioning", () => {
  it("places a weekly item using day-sized columns", () => {
    const position = calculateGanttPosition({
      start: new Date("2026-08-05T00:00:00"),
      end: new Date("2026-08-08T00:00:00"),
      dateRange: {
        start: new Date("2026-08-01T00:00:00"),
        end: new Date("2026-08-31T00:00:00"),
      },
      zoomLevel: "weeks",
    });

    expect(position).toEqual({
      leftPosition: 256,
      width: 192,
    });
  });

  it("places the current date within its month column", () => {
    const position = calculateTimelineDatePosition({
      date: new Date("2026-08-16T00:00:00"),
      dateRange: {
        start: new Date("2026-07-01T00:00:00"),
        end: new Date("2026-09-30T00:00:00"),
      },
      zoomLevel: "months",
    });

    expect(position).toBeCloseTo(120 + (15 / 31) * 120);
  });

  it("places the current date within its quarter column", () => {
    const position = calculateTimelineDatePosition({
      date: new Date("2026-05-16T00:00:00"),
      dateRange: {
        start: new Date("2026-01-01T00:00:00"),
        end: new Date("2026-12-31T00:00:00"),
      },
      zoomLevel: "quarters",
    });

    expect(position).toBeCloseTo(180 + (45 / 91) * 180);
  });

  it("returns the hovered day in week zoom", () => {
    const date = calculateTimelineDateFromPosition({
      position: 64 * 5,
      dateRange: {
        start: new Date("2026-08-01T00:00:00"),
        end: new Date("2026-08-31T00:00:00"),
      },
      zoomLevel: "weeks",
    });

    expect(date).toEqual(new Date("2026-08-06T00:00:00"));
  });

  it("returns the hovered day within a fixed-width month", () => {
    const date = calculateTimelineDateFromPosition({
      position: 60,
      dateRange: {
        start: new Date("2026-08-01T00:00:00"),
        end: new Date("2026-09-30T00:00:00"),
      },
      zoomLevel: "months",
    });

    expect(date).toEqual(new Date("2026-08-16T00:00:00"));
  });

  it.each([
    [{ leftPosition: 40, width: 80 }, "left"],
    [{ leftPosition: 900, width: 80 }, "right"],
    [{ leftPosition: 300, width: 180 }, null],
  ] as const)(
    "detects whether an item sits outside the visible timeline",
    (position, expectedDirection) => {
      expect(
        getGanttOffscreenDirection(position, {
          scrollLeft: 240,
          visibleWidth: 520,
        }),
      ).toBe(expectedDirection);
    },
  );
});
