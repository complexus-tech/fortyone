import { addMinutes } from "date-fns";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";

export const CALENDAR_DRAG_STEP_MINUTES = 5;

export type CalendarDragKind = "move" | "resize";

export const snapCalendarDeltaMinutes = (deltaY: number, hourHeight: number) =>
  Math.round(deltaY / (hourHeight / 60) / CALENDAR_DRAG_STEP_MINUTES) *
  CALENDAR_DRAG_STEP_MINUTES;

export const snapCalendarDeltaPixels = (deltaY: number, hourHeight: number) =>
  snapCalendarDeltaMinutes(deltaY, hourHeight) * (hourHeight / 60);

export const getCalendarManualChange = ({
  block,
  deltaY,
  hourHeight,
  kind,
  targetDay,
}: {
  block: CalendarScheduleBlock;
  deltaY: number;
  hourHeight: number;
  kind: CalendarDragKind;
  targetDay: Date;
}) => {
  const originalStart = new Date(block.startAt);
  const originalEnd = new Date(block.endAt);
  const deltaMinutes = snapCalendarDeltaMinutes(deltaY, hourHeight);

  if (kind === "resize") {
    const durationMinutes = Math.max(
      CALENDAR_DRAG_STEP_MINUTES,
      (originalEnd.getTime() - originalStart.getTime()) / 60_000 + deltaMinutes,
    );

    return {
      endAt: addMinutes(originalStart, durationMinutes),
      startAt: originalStart,
    };
  }

  const startAt = new Date(targetDay);
  startAt.setHours(originalStart.getHours(), originalStart.getMinutes(), 0, 0);
  const movedStartAt = addMinutes(startAt, deltaMinutes);

  return {
    endAt: addMinutes(
      movedStartAt,
      (originalEnd.getTime() - originalStart.getTime()) / 60_000,
    ),
    startAt: movedStartAt,
  };
};
