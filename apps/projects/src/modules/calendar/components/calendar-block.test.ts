/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";
import {
  getMayaCalendarBlockLabel,
  getMayaCalendarBlockReason,
  isCalendarScheduleBlockEditable,
} from "./calendar-block";

const createBlock = (
  patch: Partial<CalendarScheduleBlock> = {},
): CalendarScheduleBlock => ({
  blockType: "work",
  createdAt: "2026-08-15T08:00:00Z",
  endAt: "2026-08-15T10:00:00Z",
  hasConflict: false,
  id: "block-1",
  isLocked: false,
  source: "maya",
  startAt: "2026-08-15T09:00:00Z",
  title: "Implement calendar controls",
  updatedAt: "2026-08-15T08:00:00Z",
  ...patch,
});

describe("calendar block presentation", () => {
  it("keeps every Maya block managed by reconciliation", () => {
    expect(isCalendarScheduleBlockEditable(createBlock())).toBe(false);
    expect(
      isCalendarScheduleBlockEditable(createBlock({ isLocked: true })),
    ).toBe(false);
    expect(
      isCalendarScheduleBlockEditable(createBlock({ source: "user" })),
    ).toBe(true);
  });

  it("shows Maya provenance and enriched scheduling state when available", () => {
    const block = createBlock({
      autoSchedulingReason: "A meeting moved the earlier focus window.",
      autoSchedulingStatus: "at_risk",
    });

    expect(getMayaCalendarBlockLabel(block)).toBe("Maya · At risk");
    expect(getMayaCalendarBlockReason(block)).toBe(
      "A meeting moved the earlier focus window.",
    );
  });

  it("falls back to lock-aware Maya copy for older responses", () => {
    const block = createBlock({ isLocked: true });

    expect(getMayaCalendarBlockLabel(block)).toBe("Maya · Locked");
    expect(getMayaCalendarBlockReason(block)).toContain("stays fixed");
  });
});
