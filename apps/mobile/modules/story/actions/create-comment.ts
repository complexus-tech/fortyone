import { post } from "@/lib/http";
import type { ApiResponse, Comment } from "@/types";

export type CreateCommentPayload = {
  storyId: string;
  comment: string;
  mentions: string[];
  parentId?: string;
};

export const createComment = async ({
  storyId,
  ...payload
}: CreateCommentPayload) => {
  const response = await post<typeof payload, ApiResponse<Comment>>(
    `stories/${storyId}/comments`,
    payload,
  );
  if (response.error || !response.data)
    throw new Error(response.error?.message ?? "Your comment was not saved.");
  return response.data;
};
