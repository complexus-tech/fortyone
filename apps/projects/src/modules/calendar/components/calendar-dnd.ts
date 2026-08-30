import type { CollisionDetection, DragEndEvent, Modifier } from "@dnd-kit/core";
import { pointerWithin, rectIntersection } from "@dnd-kit/core";
import { snapCalendarDeltaPixels } from "./calendar-drag";
import { HOUR_HEIGHT } from "./calendar-presentation";
import type { CalendarDragData } from "./calendar-types";

export const getCalendarDragData = (active: DragEndEvent["active"]) =>
  (active.data.current?.calendarDrag as CalendarDragData | undefined) ?? null;

export const calendarCollisionDetection: CollisionDetection = (args) => {
  const pointerCollisions = pointerWithin(args);
  return pointerCollisions.length > 0
    ? pointerCollisions
    : rectIntersection(args);
};

const snapCalendarDragModifier: Modifier = ({ active, transform }) => {
  if (!active?.data.current?.calendarDrag) return transform;
  const drag = active.data.current.calendarDrag as CalendarDragData;
  return {
    ...transform,
    x: drag.kind === "resize" ? 0 : transform.x,
    y: snapCalendarDeltaPixels(transform.y, HOUR_HEIGHT, drag.kind),
  };
};

export const calendarDragModifiers = [snapCalendarDragModifier];
