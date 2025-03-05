"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const removeTeamMemberAction = async (
  teamId: string,
  memberId: string,
) => {
  try {
    await remove<ApiResponse<void>>(`teams/${teamId}/members/${memberId}`);
  } catch (error) {
    const apiError = getApiError(error);
    throw new Error(apiError.error?.message || "Failed to remove member");
  }
};
