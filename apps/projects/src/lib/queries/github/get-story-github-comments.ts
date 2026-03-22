import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { GitHubComment } from "@/modules/settings/workspace/integrations/github/types";

export const getStoryGitHubComments = async (
  storyId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<GitHubComment[]>>(
    `stories/${storyId}/github-comments`,
    ctx,
  );
  return response.data ?? [];
};
