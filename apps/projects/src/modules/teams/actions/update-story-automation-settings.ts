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
) => {
  try {
    const session = await auth();
    const settings = await put<
      UpdateStoryAutomationSettingsInput,
      ApiResponse<TeamStoryAutomationSettings>
    >(`teams/${teamId}/settings/story-automation`, input, session!);
    return settings;
  } catch (error) {
    return getApiError(error);
  }
};
