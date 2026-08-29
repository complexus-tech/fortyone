import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const joinPublicTeamAction = async (
  teamId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Authentication required to join teams" },
      } satisfies ApiResponse<null>;
    }

    const ctx = { session, workspaceSlug };
    return await post<Record<string, never>, ApiResponse<{ teamId: string }>>(
      `teams/${teamId}/join`,
      {},
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
