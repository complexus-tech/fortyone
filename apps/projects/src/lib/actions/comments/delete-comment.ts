"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteCommentAction = async (commentId: string) => {
  try {
    const res = await remove<ApiResponse<null>>(`comments/${commentId}`);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
