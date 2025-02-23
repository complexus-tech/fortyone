"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const removeTeamMemberAction = async (
  teamId: string,
  memberId: string,
) => {
  await remove<ApiResponse<void>>(`teams/${teamId}/members/${memberId}`);
};
