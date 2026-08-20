import { getAutoSchedulingLabel } from "@/lib/auto-scheduling";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";

export const CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE = "Scheduled elsewhere";
export const CROSS_WORKSPACE_CALENDAR_BLOCK_CONTEXT = "Another workspace";
export const CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP =
  "This time is reserved by a task in another workspace. Task details are hidden here.";

export const isCalendarScheduleBlockEditable = (block: CalendarScheduleBlock) =>
  !block.isCrossWorkspace && block.source !== "maya";

export const getMayaCalendarBlockLabel = (
  block: CalendarScheduleBlock,
): string | null => {
  if (block.isCrossWorkspace || block.source !== "maya") return null;
  if (block.autoSchedulingStatus) {
    return `Maya · ${getAutoSchedulingLabel(block.autoSchedulingStatus)}`;
  }
  return block.isLocked ? "Maya · Locked" : "Maya";
};

export const getMayaCalendarBlockReason = (
  block: CalendarScheduleBlock,
): string | null => {
  if (block.isCrossWorkspace || block.source !== "maya") return null;
  return (
    block.autoSchedulingReason?.trim() ||
    (block.isLocked
      ? "This block stays fixed until the story schedule is unlocked."
      : "Maya manages this block as availability changes.")
  );
};
