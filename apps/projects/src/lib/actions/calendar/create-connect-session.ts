import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { CreateCalendarConnectSessionResponse } from "@/modules/settings/workspace/integrations/calendar/types";

export const createCalendarConnectSessionAction = async (
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      Record<string, never>,
      ApiResponse<CreateCalendarConnectSessionResponse>
    >("integrations/calendar/google/connect-session", {}, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
