/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { buildCalendarEventLayouts } from "./calendar-layout";

describe("calendar layout", () => {
  it("positions events by time inside the visible day window", () => {
    const day = new Date(2026, 5, 15);

    const [layout] = buildCalendarEventLayouts({
      day,
      events: [
        {
          id: "event-1",
          startAt: new Date(2026, 5, 15, 10, 30).toISOString(),
          endAt: new Date(2026, 5, 15, 12).toISOString(),
        },
      ],
      hourHeight: 80,
      visibleEndHour: 18,
      visibleStartHour: 8,
    });

    expect(layout).toMatchObject({
      id: "event-1",
      height: 120,
      lane: 0,
      laneCount: 1,
      top: 200,
    });
  });

  it("assigns overlapping events to separate lanes", () => {
    const day = new Date(2026, 5, 15);

    const layouts = buildCalendarEventLayouts({
      day,
      events: [
        {
          id: "event-1",
          startAt: new Date(2026, 5, 15, 10).toISOString(),
          endAt: new Date(2026, 5, 15, 11).toISOString(),
        },
        {
          id: "event-2",
          startAt: new Date(2026, 5, 15, 10, 30).toISOString(),
          endAt: new Date(2026, 5, 15, 11, 30).toISOString(),
        },
      ],
      hourHeight: 80,
      visibleEndHour: 18,
      visibleStartHour: 8,
    });

    expect(layouts).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ id: "event-1", lane: 0, laneCount: 2 }),
        expect.objectContaining({ id: "event-2", lane: 1, laneCount: 2 }),
      ]),
    );
  });
});
