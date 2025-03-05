"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const addTeamMemberAction = async (teamId: string, memberId: string) => {
  try {
    await post<{ userId: string }, ApiResponse<{ teamId: string }>>(
      `teams/${teamId}/members`,
      {
        userId: memberId,
      },
    );
  } catch (error) {
    const apiError = getApiError(error);
    throw new Error(apiError.error?.message || "Failed to add member");
  }
};
