import { post, put, remove } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  CalendarScheduleBlock,
  CalendarScheduleBlockInput,
} from "./types";

export const createCalendarScheduleBlock = async (
  ctx: WorkspaceCtx,
  input: CalendarScheduleBlockInput,
) => {
  const response = await post<
    CalendarScheduleBlockInput,
    ApiResponse<CalendarScheduleBlock>
  >("calendar/schedule-blocks", input, ctx);
  return response.data!;
};

export const updateCalendarScheduleBlock = async (
  ctx: WorkspaceCtx,
  blockId: string,
  input: CalendarScheduleBlockInput,
) => {
  const response = await put<
    CalendarScheduleBlockInput,
    ApiResponse<CalendarScheduleBlock>
  >(`calendar/schedule-blocks/${blockId}`, input, ctx);
  return response.data!;
};

export const deleteCalendarScheduleBlock = async (
  ctx: WorkspaceCtx,
  blockId: string,
) => remove<ApiResponse<null>>(`calendar/schedule-blocks/${blockId}`, ctx);
