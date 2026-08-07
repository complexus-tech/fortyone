/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  buildCalendarEventLayouts,
  deriveCalendarVisibleHours,
  getDisplayBusyWindows,
  parseCalendarDate,
} from "./calendar-layout";

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

  it.each([new Date(2026, 2, 8), new Date(2026, 10, 1)])(
    "keeps the visible grid on wall-clock time across DST (%s)",
    (day) => {
      const eventStart = new Date(day);
      eventStart.setHours(8, 0, 0, 0);
      const eventEnd = new Date(day);
      eventEnd.setHours(9, 0, 0, 0);

      const [layout] = buildCalendarEventLayouts({
        day,
        events: [
          {
            id: "event-1",
            startAt: eventStart.toISOString(),
            endAt: eventEnd.toISOString(),
          },
        ],
        hourHeight: 80,
        visibleEndHour: 18,
        visibleStartHour: 8,
      });

      expect(layout).toMatchObject({ height: 80, top: 0 });
    },
  );

  it.each([
    {
      endAt: "2026-03-08T03:30:00-04:00",
      startAt: "2026-03-08T01:30:00-05:00",
    },
    {
      endAt: "2026-11-01T01:30:00-05:00",
      startAt: "2026-11-01T01:30:00-04:00",
    },
  ])("uses elapsed duration across a DST transition", ({ endAt, startAt }) => {
    const localStart = new Date(startAt);
    const day = new Date(
      localStart.getFullYear(),
      localStart.getMonth(),
      localStart.getDate(),
    );
    const [layout] = buildCalendarEventLayouts({
      day,
      events: [{ id: "event-1", startAt, endAt }],
      hourHeight: 60,
      visibleEndHour: 24,
      visibleStartHour: 0,
    });

    expect(layout).toMatchObject({ height: 60 });
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

  it("expands the visible hours for early and evening meetings", () => {
    expect(
      deriveCalendarVisibleHours({
        defaultEndHour: 18,
        defaultStartHour: 8,
        events: [
          {
            startAt: new Date(2026, 5, 15, 6, 30).toISOString(),
            endAt: new Date(2026, 5, 15, 7, 30).toISOString(),
          },
          {
            startAt: new Date(2026, 5, 16, 19).toISOString(),
            endAt: new Date(2026, 5, 16, 20, 30).toISOString(),
          },
        ],
      }),
    ).toEqual({ visibleStartHour: 6, visibleEndHour: 21 });
  });

  it("does not expand timed hours for all-day events", () => {
    expect(
      deriveCalendarVisibleHours({
        defaultEndHour: 18,
        defaultStartHour: 8,
        events: [
          {
            startAt: new Date(2026, 5, 15).toISOString(),
            endAt: new Date(2026, 5, 16).toISOString(),
            isAllDay: true,
          },
        ],
      }),
    ).toEqual({ visibleStartHour: 8, visibleEndHour: 18 });
  });

  it("uses busy windows only for availability-only snapshots", () => {
    const busyWindows = [{ id: "busy-window" }];

    expect(
      getDisplayBusyWindows({ busyWindows, events: [{ id: "event" }] }),
    ).toEqual([]);
    expect(getDisplayBusyWindows({ busyWindows, events: [] })).toEqual(
      busyWindows,
    );
  });

  it("parses all-day values as local calendar dates", () => {
    const date = parseCalendarDate("2026-08-07");

    expect(date).not.toBeNull();
    expect(date?.getFullYear()).toBe(2026);
    expect(date?.getMonth()).toBe(7);
    expect(date?.getDate()).toBe(7);
    expect(parseCalendarDate("2026-02-30")).toBeNull();
  });
});
