import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const removeTeamMemberAction = async (
  teamId: string,
  memberId: string,
) => {
  try {
    const res = await remove<ApiResponse<void>>(
      `teams/${teamId}/members/${memberId}`,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
