import { auth } from "@/auth";
import type { SlackAgentSettings } from "@/modules/settings/workspace/integrations/slack/types";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const updateSlackAgentSettingsAction = async (
  workspaceSlug: string,
  input: SlackAgentSettings,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<SlackAgentSettings, ApiResponse<SlackAgentSettings>>(
      "integrations/slack/agent-settings",
      input,
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
