"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";

export type UpdateComment = {
  content: string;
};

export const updateCommentAction = async (
  commentId: string,
  payload: UpdateComment,
) => {
  try {
    await put<UpdateComment, null>(`comments/${commentId}`, payload);
    return commentId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update comment");
  }
};
