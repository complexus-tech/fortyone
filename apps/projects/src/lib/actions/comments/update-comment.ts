"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import { getApiError } from "@/utils";

export type UpdateComment = {
  content: string;
};

export const updateCommentAction = async (
  commentId: string,
  payload: UpdateComment,
  storyId: string,
) => {
  try {
    await put<UpdateComment, null>(`comments/${commentId}`, payload);
    revalidateTag(storyTags.comments(storyId));
    return commentId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update comment");
  }
};
