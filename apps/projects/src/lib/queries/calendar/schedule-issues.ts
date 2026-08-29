import { post } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const retryCalendarScheduleIssue = async (
  ctx: WorkspaceCtx,
  storyId: string,
) =>
  post<Record<string, never>, ApiResponse<null>>(
    `maya/schedule-issues/${storyId}/retry`,
    {},
    ctx,
  );

export const overrideCalendarScheduleIssue = async (
  ctx: WorkspaceCtx,
  storyId: string,
  input: { startAt: string; timezone: string },
) =>
  post<typeof input, ApiResponse<null>>(
    `maya/schedule-issues/${storyId}/override`,
    input,
    ctx,
  );
