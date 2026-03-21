import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  GitHubWorkspaceSettings,
  UpdateGitHubWorkspaceSettingsInput,
} from "@/modules/integrations/github/types";

export const updateGitHubWorkspaceSettingsAction = async (
  input: UpdateGitHubWorkspaceSettingsInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<
      UpdateGitHubWorkspaceSettingsInput,
      ApiResponse<GitHubWorkspaceSettings>
    >("integrations/github/settings", input, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
