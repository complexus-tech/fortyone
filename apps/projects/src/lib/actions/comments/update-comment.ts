"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type UpdateComment = {
  content: string;
  mentions: string[];
};

export const updateCommentAction = async (
  commentId: string,
  payload: UpdateComment,
) => {
  try {
    const session = await auth();
    const res = await put<UpdateComment, ApiResponse<null>>(
      `comments/${commentId}`,
      payload,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
