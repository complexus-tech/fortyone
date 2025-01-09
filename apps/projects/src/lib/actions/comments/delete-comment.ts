"use server";

import { remove } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import { revalidateTag } from "next/cache";

export const deleteCommentAction = async (
  commentId: string,
  storyId: string,
) => {
  const _ = await remove(`comments/${commentId}`);
  revalidateTag(storyTags.comments(storyId));
  return commentId;
};
