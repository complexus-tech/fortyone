import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { CreateSlackInstallSessionResponse } from "@/modules/settings/workspace/integrations/slack/types";

export const createSlackInstallSessionAction = async (
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      Record<string, never>,
      ApiResponse<CreateSlackInstallSessionResponse>
    >("integrations/slack/install-session", {}, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
