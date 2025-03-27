"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export type UpdateComment = {
  content: string;
};

export const updateCommentAction = async (
  commentId: string,
  payload: UpdateComment,
) => {
  try {
    const res = await put<UpdateComment, ApiResponse<null>>(
      `comments/${commentId}`,
      payload,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
