/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";
import {
  CALENDAR_DRAG_STEP_MINUTES,
  getCalendarManualChange,
  snapCalendarDeltaMinutes,
} from "./calendar-drag";

const createBlock = (
  patch: Partial<CalendarScheduleBlock> = {},
): CalendarScheduleBlock => ({
  blockType: "work",
  createdAt: "2026-08-20T08:00:00.000Z",
  endAt: new Date(2026, 7, 20, 11).toISOString(),
  hasConflict: false,
  id: "block-1",
  isLocked: false,
  source: "user",
  startAt: new Date(2026, 7, 20, 10).toISOString(),
  storyId: "story-1",
  title: "Prepare launch notes",
  updatedAt: "2026-08-20T08:00:00.000Z",
  ...patch,
});

describe("calendar drag calculations", () => {
  it("snaps vertical movement to five-minute increments", () => {
    const fiveMinutePixels = (52 / 60) * CALENDAR_DRAG_STEP_MINUTES;

    expect(snapCalendarDeltaMinutes(fiveMinutePixels, 52)).toBe(5);
    expect(snapCalendarDeltaMinutes(fiveMinutePixels * 1.6, 52)).toBe(10);
    expect(snapCalendarDeltaMinutes(-fiveMinutePixels, 52)).toBe(-5);
  });

  it("moves a block across dates while preserving its duration", () => {
    const block = createBlock();
    const result = getCalendarManualChange({
      block,
      deltaY: (52 / 60) * 5,
      hourHeight: 52,
      kind: "move",
      targetDay: new Date(2026, 7, 21),
    });

    expect(result.startAt.getFullYear()).toBe(2026);
    expect(result.startAt.getMonth()).toBe(7);
    expect(result.startAt.getDate()).toBe(21);
    expect(result.startAt.getHours()).toBe(10);
    expect(result.startAt.getMinutes()).toBe(5);
    expect(result.endAt.getTime() - result.startAt.getTime()).toBe(60 * 60_000);
  });

  it("resizes in five-minute increments and keeps a five-minute minimum", () => {
    const block = createBlock({
      endAt: new Date(2026, 7, 20, 10, 10).toISOString(),
    });
    const result = getCalendarManualChange({
      block,
      deltaY: -52,
      hourHeight: 52,
      kind: "resize",
      targetDay: new Date(2026, 7, 20),
    });

    expect(result.endAt.getTime() - result.startAt.getTime()).toBe(5 * 60_000);
  });
});
