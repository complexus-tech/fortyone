"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const addTeamMemberAction = async (teamId: string, memberId: string) => {
  try {
    const session = await auth();
    const res = await post<{ userId: string }, ApiResponse<{ teamId: string }>>(
      `teams/${teamId}/members`,
      {
        userId: memberId,
      },
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
