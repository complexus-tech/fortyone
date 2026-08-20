/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";
import {
  CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE,
  CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP,
  RESERVED_TIME_BLOCK_CLASS,
  getCalendarScheduleBlockSecondaryLabel,
  getCalendarStoryBlockStyle,
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
  it("keeps reserved time dashed and lined", () => {
    expect(RESERVED_TIME_BLOCK_CLASS).toContain("border-dashed");
    expect(RESERVED_TIME_BLOCK_CLASS).toContain("repeating-linear-gradient");
  });

  it("uses the story status color for scheduled work", () => {
    expect(
      getCalendarStoryBlockStyle(createBlock({ storyStatusColor: "#3c90ff" })),
    ).toEqual({
      "--calendar-story-accent": "#3c90ff",
      "--calendar-story-background": "rgba(60, 144, 255, 0.1)",
      "--calendar-story-border": "rgba(60, 144, 255, 0.2)",
      "--calendar-story-hover": "rgba(60, 144, 255, 0.15)",
    });
  });

  it("keeps conflicts, focus time, and cross-workspace blocks semantic", () => {
    for (const patch of [
      { hasConflict: true },
      { blockType: "focus" as const },
      { isCrossWorkspace: true },
      { storyStatusColor: "invalid" },
    ]) {
      expect(
        getCalendarStoryBlockStyle(
          createBlock({ storyStatusColor: "#3c90ff", ...patch }),
        ),
      ).toBeUndefined();
    }
  });

  it("keeps every Maya block managed by reconciliation", () => {
    expect(isCalendarScheduleBlockEditable(createBlock())).toBe(false);
    expect(
      isCalendarScheduleBlockEditable(createBlock({ isLocked: true })),
    ).toBe(false);
    expect(
      isCalendarScheduleBlockEditable(createBlock({ source: "user" })),
    ).toBe(true);
    expect(
      isCalendarScheduleBlockEditable(
        createBlock({ isCrossWorkspace: true, source: "other_workspace" }),
      ),
    ).toBe(false);
  });

  it("does not expose Maya provenance for another workspace", () => {
    const block = createBlock({ isCrossWorkspace: true });

    expect(getMayaCalendarBlockLabel(block)).toBeNull();
    expect(getMayaCalendarBlockReason(block)).toBeNull();
    expect(
      getCalendarScheduleBlockSecondaryLabel(
        block,
        "Another workspace",
        "9 – 10am",
      ),
    ).toBe("9 – 10am");
    expect(CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE).toBe("Scheduled elsewhere");
    expect(CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP).toBe(
      "This time is reserved by a task in another workspace. Task details are hidden here.",
    );
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
