"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  TeamStoryAutomationSettings,
  UpdateStoryAutomationSettingsInput,
} from "../types";

export const updateStoryAutomationSettingsAction = async (
  teamId: string,
  input: UpdateStoryAutomationSettingsInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const settings = await put<
      UpdateStoryAutomationSettingsInput,
      ApiResponse<TeamStoryAutomationSettings>
    >(`teams/${teamId}/settings/story-automation`, input, ctx);
    return settings;
  } catch (error) {
    return getApiError(error);
  }
};
