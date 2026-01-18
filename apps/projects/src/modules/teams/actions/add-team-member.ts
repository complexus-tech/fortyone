"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const addTeamMemberAction = async (
  teamId: string,
  memberId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<{ userId: string }, ApiResponse<{ teamId: string }>>(
      `teams/${teamId}/members`,
      {
        userId: memberId,
      },
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
