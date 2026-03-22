import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteStoryGitHubLinkAction = async (
  storyId: string,
  linkId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await remove<ApiResponse<null>>(
      `stories/${storyId}/github-links/${linkId}`,
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
