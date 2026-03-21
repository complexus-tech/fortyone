import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { GitHubTeamSettings } from "@/modules/settings/workspace/integrations/github/types";

export const getGitHubTeamSettings = async (
  teamId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<GitHubTeamSettings>>(
    `teams/${teamId}/settings/github`,
    ctx,
  );
  return response.data!;
};
