import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { WorkspaceSettings } from "@/types/workspace";

export const getWorkspaceSettings = async () => {
  const response = await get<ApiResponse<WorkspaceSettings>>("settings");
  return response.data!;
};
