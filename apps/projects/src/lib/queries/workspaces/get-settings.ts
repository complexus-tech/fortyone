"use server";

import { get } from "@/lib/http";
import type { ApiResponse, WorkspaceSettings } from "@/types";

export const getWorkspaceSettings = async () => {
  const settings = await get<ApiResponse<WorkspaceSettings>>("settings");
  return settings.data!;
};
