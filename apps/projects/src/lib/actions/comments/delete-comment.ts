"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteCommentAction = async (commentId: string) => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<null>>(
      `comments/${commentId}`,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
