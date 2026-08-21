import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { CalendarConnection } from "@/modules/settings/workspace/integrations/calendar/types";

export const setPrimaryCalendarConnectionAction = async (
  workspaceSlug: string,
  connectionId: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<Record<string, never>, ApiResponse<CalendarConnection>>(
      `integrations/calendar/${connectionId}/primary`,
      {},
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
