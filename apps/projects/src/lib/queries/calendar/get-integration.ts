import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { CalendarIntegration } from "@/modules/settings/workspace/integrations/calendar/types";

export const getCalendarIntegration = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<CalendarIntegration>>(
    "integrations/calendar",
    ctx,
  );
  return response.data!;
};
