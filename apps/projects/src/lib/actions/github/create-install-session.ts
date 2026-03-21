import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { CreateGitHubInstallSessionResponse } from "@/modules/integrations/github/types";

export const createGitHubInstallSessionAction = async (
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      Record<string, never>,
      ApiResponse<CreateGitHubInstallSessionResponse>
    >("integrations/github/install-session", {}, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
