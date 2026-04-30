import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type PostGitHubCommentPayload = {
  body: string;
};

export const postRequestGitHubCommentAction = async (
  requestId: string,
  payload: PostGitHubCommentPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<PostGitHubCommentPayload, ApiResponse<null>>(
      `integration-requests/${requestId}/github-comments`,
      payload,
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
