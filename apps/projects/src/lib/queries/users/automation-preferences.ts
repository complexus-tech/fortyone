import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, AutomationPreferences } from "@/types";

export const getAutomationPreferences = async (ctx: WorkspaceCtx) => {
  const preferences = await get<ApiResponse<AutomationPreferences>>(
    "automation/preferences",
    ctx,
  );
  return preferences.data!;
};
