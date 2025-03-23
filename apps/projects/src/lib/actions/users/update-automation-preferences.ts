import { put } from "@/lib/http";
import type {
  ApiResponse,
  UpdateAutomationPreferences,
  AutomationPreferences,
} from "@/types";
import { getApiError } from "@/utils";

export const updateAutomationPreferencesAction = async (
  payload: UpdateAutomationPreferences,
) => {
  try {
    const res = await put<
      UpdateAutomationPreferences,
      ApiResponse<AutomationPreferences>
    >("automation/preferences", payload);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
