import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { WorkspaceSettings } from "@/types/workspace";

export const getWorkspaceSettings = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<WorkspaceSettings>>("settings", {
    signal,
  });
  return response.data!;
};
