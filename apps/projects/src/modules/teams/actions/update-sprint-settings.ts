"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { TeamSprintSettings, UpdateSprintSettingsInput } from "../types";

export const updateSprintSettingsAction = async (
  teamId: string,
  input: UpdateSprintSettingsInput,
) => {
  try {
    const session = await auth();
    const settings = await put<
      UpdateSprintSettingsInput,
      ApiResponse<TeamSprintSettings>
    >(`teams/${teamId}/settings/sprints`, input, session!);
    return settings;
  } catch (error) {
    return getApiError(error);
  }
};
