import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, AutomationPreferences } from "@/types";

export const getAutomationPreferences = async (session: Session) => {
  const preferences = await get<ApiResponse<AutomationPreferences>>(
    "automation/preferences",
    session,
  );
  return preferences.data!;
};
