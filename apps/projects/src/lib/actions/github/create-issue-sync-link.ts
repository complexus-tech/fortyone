import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  CreateGitHubIssueSyncLinkInput,
  GitHubIssueSyncLink,
} from "@/modules/settings/workspace/integrations/github/types";

export const createGitHubIssueSyncLinkAction = async (
  input: CreateGitHubIssueSyncLinkInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      CreateGitHubIssueSyncLinkInput,
      ApiResponse<GitHubIssueSyncLink>
    >("integrations/github/issue-sync-links", input, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
