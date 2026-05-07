import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  SlackWorkspaceSettings,
  UpdateSlackWorkspaceSettingsInput,
} from "@/modules/settings/workspace/integrations/slack/types";

export const updateSlackWorkspaceSettingsAction = async (
  input: UpdateSlackWorkspaceSettingsInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<
      UpdateSlackWorkspaceSettingsInput,
      ApiResponse<SlackWorkspaceSettings>
    >("integrations/slack/settings", input, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
