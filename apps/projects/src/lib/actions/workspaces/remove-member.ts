import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const removeMemberAction = async (memberId: string) => {
  try {
    await remove<ApiResponse<void>>(`/members/${memberId}`);
  } catch (error) {
    const apiError = getApiError(error);
    throw new Error(apiError.error?.message || "Failed to remove member");
  }
};
