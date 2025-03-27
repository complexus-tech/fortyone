"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const addTeamMemberAction = async (teamId: string, memberId: string) => {
  try {
    const res = await post<{ userId: string }, ApiResponse<{ teamId: string }>>(
      `teams/${teamId}/members`,
      {
        userId: memberId,
      },
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
