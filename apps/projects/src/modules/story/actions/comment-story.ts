import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse, Comment } from "@/types";
import { getApiError } from "@/utils";
import type { StoryCommentPayload } from "@/shared/story/comment-payload";

export const commentStoryAction = async (
  storyId: string,
  payload: StoryCommentPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<StoryCommentPayload, ApiResponse<Comment>>(
      `stories/${storyId}/comments`,
      payload,
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
