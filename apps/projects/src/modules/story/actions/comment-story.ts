import { post } from "@/lib/http";
import type { ApiResponse, Comment } from "@/types";
import { getApiError } from "@/utils";

type CommentPayload = {
  comment: string;
  parentId?: string | null;
};

export const commentStoryAction = async (
  storyId: string,
  payload: CommentPayload,
) => {
  try {
    const res = await post<CommentPayload, ApiResponse<Comment>>(
      `stories/${storyId}/comments`,
      payload,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
