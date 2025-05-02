"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const removeTeamMemberAction = async (
  teamId: string,
  memberId: string,
) => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<void>>(
      `teams/${teamId}/members/${memberId}`,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
