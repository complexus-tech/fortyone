import { format, isSameDay } from "date-fns";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";

export const getCalendarScheduleBlockTimeLabel = (
  block: Pick<CalendarScheduleBlock, "startAt" | "endAt">,
) => {
  const start = new Date(block.startAt);
  const end = new Date(block.endAt);
  if (isSameDay(start, end)) {
    return `${format(start, "EEEE, MMMM d")} · ${format(start, "h:mm a")} – ${format(end, "h:mm a")}`;
  }
  return `${format(start, "MMM d, h:mm a")} – ${format(end, "MMM d, h:mm a")}`;
};
