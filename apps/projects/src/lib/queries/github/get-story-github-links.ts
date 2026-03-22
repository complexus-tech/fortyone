import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoryGitHubLink } from "@/modules/settings/workspace/integrations/github/types";

export const getStoryGitHubLinks = async (
  storyId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<StoryGitHubLink[]>>(
    `stories/${storyId}/github-links`,
    ctx,
  );
  return response.data ?? [];
};
