/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  calculateGanttPosition,
  calculateTimelineDateFromPosition,
  calculateTimelineDatePosition,
  getGanttVirtualRows,
  getGanttOffscreenDirection,
} from "./base-gantt-utils";
import {
  getGanttBarDatesForDrag,
  getGanttBarDragPosition,
} from "./base-gantt-bar-utils";

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

describe("BaseGantt vertical windowing", () => {
  it("uses the same fixed row coordinates at the top and middle", () => {
    const top = getGanttVirtualRows({
      itemCount: 100,
      overscan: 1,
      rowHeight: 56,
      scrollTop: 0,
      viewportHeight: 176,
    });
    const middle = getGanttVirtualRows({
      itemCount: 100,
      overscan: 1,
      rowHeight: 56,
      scrollTop: 64 + 56 * 20,
      viewportHeight: 168,
    });

    expect(top.rows.map(({ index, start }) => [index, start])).toEqual([
      [0, 0],
      [1, 56],
      [2, 112],
    ]);
    expect(middle.rows.map(({ index, start }) => [index, start])).toEqual([
      [19, 1_064],
      [20, 1_120],
      [21, 1_176],
      [22, 1_232],
      [23, 1_288],
    ]);
    expect(middle.totalSize).toBe(5_600);
  });

  it("pins an active drag row without mounting the intervening rows", () => {
    const layout = getGanttVirtualRows({
      itemCount: 100,
      overscan: 0,
      pinnedIndices: [80],
      rowHeight: 56,
      scrollTop: 64,
      viewportHeight: 112,
    });

    expect(layout.rows.map(({ index }) => index)).toEqual([0, 1, 80]);
  });

  it("clamps a stale vertical anchor after filtering", () => {
    const layout = getGanttVirtualRows({
      itemCount: 3,
      overscan: 0,
      rowHeight: 56,
      scrollTop: 10_000,
      viewportHeight: 300,
    });

    expect(layout.rows.map(({ index }) => index)).toEqual([2]);
    expect(layout.totalSize).toBe(168);
  });
});

describe("BaseGantt bar dragging", () => {
  const dateRange = {
    start: new Date("2026-08-01T00:00:00"),
    end: new Date("2026-08-31T00:00:00"),
  };

  it("keeps visual and persisted dates aligned after a weekly move", () => {
    const dragStart = {
      originalEndDate: new Date("2026-08-04T00:00:00"),
      originalLeft: 64,
      originalStartDate: new Date("2026-08-02T00:00:00"),
      originalWidth: 128,
      type: "move" as const,
      x: 100,
    };
    const position = getGanttBarDragPosition({
      dateRange,
      dragStart,
      pixelOffsetX: 64,
      zoomLevel: "weeks",
    });
    const dates = getGanttBarDatesForDrag({
      dateRange,
      dragStart,
      pixelOffsetX: 64,
      zoomLevel: "weeks",
    });

    expect(position).toEqual({ leftPosition: 128, width: 128 });
    expect(dates).toEqual({
      endDate: new Date("2026-08-05T00:00:00"),
      startDate: new Date("2026-08-03T00:00:00"),
    });
  });

  it("never yields a zero-day duration while resizing", () => {
    const dates = getGanttBarDatesForDrag({
      dateRange,
      dragStart: {
        originalEndDate: new Date("2026-08-04T00:00:00"),
        originalLeft: 64,
        originalStartDate: new Date("2026-08-02T00:00:00"),
        originalWidth: 128,
        type: "resize-end",
        x: 100,
      },
      pixelOffsetX: -10_000,
      zoomLevel: "weeks",
    });

    expect(dates.endDate).toEqual(new Date("2026-08-03T00:00:00"));
    expect(dates.startDate).toEqual(new Date("2026-08-02T00:00:00"));
  });
});
