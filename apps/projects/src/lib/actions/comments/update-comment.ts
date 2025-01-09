"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { storyTags } from "@/modules/stories/constants";

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
