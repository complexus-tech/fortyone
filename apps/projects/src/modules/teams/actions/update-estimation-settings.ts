"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  TeamEstimationSettings,
  UpdateEstimationSettingsInput,
} from "../types";

export const updateEstimationSettingsAction = async (
  teamId: string,
  input: UpdateEstimationSettingsInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const settings = await put<
      UpdateEstimationSettingsInput,
      ApiResponse<TeamEstimationSettings>
    >(`teams/${teamId}/settings/estimation`, input, ctx);
    return settings;
  } catch (error) {
    return getApiError(error);
  }
};
