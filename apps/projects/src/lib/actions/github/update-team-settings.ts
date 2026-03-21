import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  GitHubTeamSettings,
  UpdateGitHubTeamSettingsInput,
} from "@/modules/integrations/github/types";

export const updateGitHubTeamSettingsAction = async (
  teamId: string,
  input: UpdateGitHubTeamSettingsInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<
      UpdateGitHubTeamSettingsInput,
      ApiResponse<GitHubTeamSettings>
    >(`teams/${teamId}/settings/github`, input, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
