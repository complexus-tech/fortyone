import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const leaveTeamAction = async (
  teamId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Authentication required to leave teams" },
      } satisfies ApiResponse<null>;
    }

    const ctx = { session, workspaceSlug };
    return await remove<ApiResponse<void>>(`teams/${teamId}/membership`, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
