"use server";

import { put } from "@/lib/http";
import type {
  ApiResponse,
  UpdateAutomationPreferences,
  AutomationPreferences,
} from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const updateAutomationPreferencesAction = async (
  payload: UpdateAutomationPreferences,
) => {
  try {
    const session = await auth();
    const res = await put<
      UpdateAutomationPreferences,
      ApiResponse<AutomationPreferences>
    >("automation/preferences", payload, session!);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
