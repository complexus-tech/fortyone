import { get } from "@/lib/http";
import type { ApiResponse, AutomationPreferences } from "@/types";

export const getAutomationPreferences = async () => {
  const preferences = await get<ApiResponse<AutomationPreferences>>(
    "automation/preferences",
  );
  return preferences.data!;
};
