import type { CSSProperties } from "react";
import { hexToRgba } from "@/utils";
import { getAutoSchedulingLabel } from "@/lib/auto-scheduling";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";

export const CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE = "Scheduled elsewhere";
export const CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP =
  "This time is reserved by a task in another workspace. Task details are hidden here.";
export const RESERVED_TIME_BLOCK_CLASS =
  "border-border-strong/50 bg-surface-muted/55 text-text-muted dark:bg-surface-elevated/65 border-dashed bg-[repeating-linear-gradient(135deg,transparent_0,transparent_5px,rgba(100,116,139,0.1)_5px,rgba(100,116,139,0.1)_8px)]";

type CalendarStoryBlockStyle = CSSProperties & {
  "--calendar-story-accent": string;
  "--calendar-story-background": string;
  "--calendar-story-border": string;
  "--calendar-story-hover": string;
};

export const getCalendarStoryBlockStyle = (
  block: CalendarScheduleBlock,
): CalendarStoryBlockStyle | undefined => {
  const color = block.storyStatusColor?.trim();
  if (
    block.blockType !== "work" ||
    block.hasConflict ||
    block.isCrossWorkspace ||
    !color
  ) {
    return undefined;
  }

  try {
    return {
      "--calendar-story-accent": color,
      "--calendar-story-background": hexToRgba(color, 0.1),
      "--calendar-story-border": hexToRgba(color, 0.2),
      "--calendar-story-hover": hexToRgba(color, 0.15),
    };
  } catch {
    return undefined;
  }
};

export const isCalendarScheduleBlockEditable = (block: CalendarScheduleBlock) =>
  !block.isCrossWorkspace && block.source !== "maya";

export const getCalendarScheduleBlockSecondaryLabel = (
  block: CalendarScheduleBlock,
  statusLabel: string,
  timeLabel: string,
) =>
  block.isCrossWorkspace || block.source === "maya"
    ? timeLabel
    : `${statusLabel} · ${timeLabel}`;

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
