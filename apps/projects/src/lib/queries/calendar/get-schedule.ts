import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { CalendarSchedule } from "./types";

export const getCalendarSchedule = async (
  ctx: WorkspaceCtx,
  params: { startAt: string; endAt: string },
) => {
  const query = new URLSearchParams({
    start: params.startAt,
    end: params.endAt,
  });
  const response = await get<ApiResponse<CalendarSchedule>>(
    `calendar/schedule?${query.toString()}`,
    ctx,
  );
  return response.data!;
};
