"use server";

import { put } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";
import { revalidateTag } from "next/cache";

export type UpdateComment = {
  content: string;
};

export const updateCommentAction = async (
  commentId: string,
  payload: UpdateComment,
  storyId: string,
) => {
  const _ = await put<UpdateComment, any>(`comments/${commentId}`, payload);
  revalidateTag(storyTags.comments(storyId));
  return commentId;
};
