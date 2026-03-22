import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type PostGitHubCommentPayload = {
  body: string;
};

export const postGitHubCommentAction = async (
  storyId: string,
  payload: PostGitHubCommentPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<PostGitHubCommentPayload, ApiResponse<null>>(
      `stories/${storyId}/github-comments`,
      payload,
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
