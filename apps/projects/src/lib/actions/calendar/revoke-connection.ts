import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const revokeCalendarConnectionAction = async (
  workspaceSlug: string,
  connectionId: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await remove<ApiResponse<null>>(
      `integrations/calendar/${connectionId}`,
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
