"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse, Comment } from "@/types";
import { getApiError } from "@/utils";

type CommentPayload = {
  comment: string;
  parentId?: string | null;
  mentions: string[];
};

export const commentStoryAction = async (
  storyId: string,
  payload: CommentPayload,
) => {
  try {
    const session = await auth();
    const res = await post<CommentPayload, ApiResponse<Comment>>(
      `stories/${storyId}/comments`,
      payload,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
