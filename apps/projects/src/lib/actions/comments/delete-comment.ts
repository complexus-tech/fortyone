"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import { getApiError } from "@/utils";

export const deleteCommentAction = async (
  commentId: string,
  storyId: string,
) => {
  try {
    const _ = await remove(`comments/${commentId}`);
    revalidateTag(storyTags.comments(storyId));
    return commentId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete comment");
  }
};
