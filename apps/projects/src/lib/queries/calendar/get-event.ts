import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { CalendarEventDetail } from "./types";

export const getCalendarEvent = async (ctx: WorkspaceCtx, eventId: string) => {
  const response = await get<ApiResponse<CalendarEventDetail>>(
    `calendar/events/${encodeURIComponent(eventId)}`,
    ctx,
  );
  return response.data!;
};
