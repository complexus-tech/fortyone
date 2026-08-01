/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  calculateGanttPosition,
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
