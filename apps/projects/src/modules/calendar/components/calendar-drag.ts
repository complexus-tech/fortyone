import { addDays, addMinutes, startOfDay } from "date-fns";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";

export const CALENDAR_MOVE_STEP_MINUTES = 1;
export const CALENDAR_RESIZE_STEP_MINUTES = 5;

export type CalendarDragKind = "move" | "resize";

export const activateCalendarResize = <
  TEvent extends { stopPropagation: () => void },
>(
  event: TEvent,
  listener?: (event: TEvent) => void,
) => {
  event.stopPropagation();
  listener?.(event);
};

const getCalendarDragStepMinutes = (kind: CalendarDragKind) =>
  kind === "move" ? CALENDAR_MOVE_STEP_MINUTES : CALENDAR_RESIZE_STEP_MINUTES;

export const snapCalendarDeltaMinutes = (
  deltaY: number,
  hourHeight: number,
  kind: CalendarDragKind,
) => {
  const stepMinutes = getCalendarDragStepMinutes(kind);
  return Math.round(deltaY / (hourHeight / 60) / stepMinutes) * stepMinutes;
};

export const snapCalendarDeltaPixels = (
  deltaY: number,
  hourHeight: number,
  kind: CalendarDragKind,
) => snapCalendarDeltaMinutes(deltaY, hourHeight, kind) * (hourHeight / 60);

export const resizeCalendarBlockByMinutes = (
  block: CalendarScheduleBlock,
  deltaMinutes: number,
) => {
  const startAt = new Date(block.startAt);
  const endAt = new Date(block.endAt);
  const durationMinutes = Math.max(
    CALENDAR_RESIZE_STEP_MINUTES,
    (endAt.getTime() - startAt.getTime()) / 60_000 + deltaMinutes,
  );

  return {
    endAt: addMinutes(startAt, durationMinutes),
    startAt,
  };
};

export const isCalendarBlockResizeTerminalDay = (
  block: Pick<CalendarScheduleBlock, "endAt">,
  day: Date,
) => {
  const dayStart = startOfDay(day);
  const endAt = new Date(block.endAt);
  return endAt > dayStart && endAt <= addDays(dayStart, 1);
};

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
  const deltaMinutes = snapCalendarDeltaMinutes(deltaY, hourHeight, kind);

  if (kind === "resize") {
    return resizeCalendarBlockByMinutes(block, deltaMinutes);
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
