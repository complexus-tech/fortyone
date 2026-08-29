import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  CalendarProvider,
  CreateCalendarConnectSessionResponse,
} from "@/modules/settings/workspace/integrations/calendar/types";

export const createCalendarConnectSessionAction = async (
  workspaceSlug: string,
  provider: CalendarProvider,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      Record<string, never>,
      ApiResponse<CreateCalendarConnectSessionResponse>
    >(`integrations/calendar/${provider}/connect-session`, {}, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
