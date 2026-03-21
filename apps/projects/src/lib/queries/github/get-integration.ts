import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { GitHubIntegration } from "@/modules/settings/workspace/integrations/github/types";

export const getGitHubIntegration = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<GitHubIntegration>>(
    "integrations/github",
    ctx,
  );
  return response.data!;
};
