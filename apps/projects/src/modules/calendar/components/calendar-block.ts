import { getAutoSchedulingLabel } from "@/lib/auto-scheduling";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";

export const isCalendarScheduleBlockEditable = (block: CalendarScheduleBlock) =>
  block.source !== "maya";

export const getMayaCalendarBlockLabel = (
  block: CalendarScheduleBlock,
): string | null => {
  if (block.source !== "maya") return null;
  if (block.autoSchedulingStatus) {
    return `Maya · ${getAutoSchedulingLabel(block.autoSchedulingStatus)}`;
  }
  return block.isLocked ? "Maya · Locked" : "Maya";
};

export const getMayaCalendarBlockReason = (
  block: CalendarScheduleBlock,
): string | null => {
  if (block.source !== "maya") return null;
  return (
    block.autoSchedulingReason?.trim() ||
    (block.isLocked
      ? "This block stays fixed until the story schedule is unlocked."
      : "Maya manages this block as availability changes.")
  );
};
