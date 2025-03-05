"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteCommentAction = async (commentId: string) => {
  try {
    const _ = await remove(`comments/${commentId}`);
    return commentId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete comment");
  }
};
