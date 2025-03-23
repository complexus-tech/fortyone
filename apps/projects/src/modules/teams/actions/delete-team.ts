import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteTeamAction = async (id: string) => {
  try {
    const res = await remove<ApiResponse<void>>(`teams/${id}`);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
