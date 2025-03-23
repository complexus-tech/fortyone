import { put } from "@/lib/http";
import type { ApiResponse, WorkspaceSettings } from "@/types";
import { getApiError } from "@/utils";

export type UpdateWorkspaceSettings = Partial<WorkspaceSettings>;

export const updateWorkspaceSettingsAction = async (
  payload: UpdateWorkspaceSettings,
) => {
  try {
    const res = await put<
      UpdateWorkspaceSettings,
      ApiResponse<WorkspaceSettings>
    >("settings", payload);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
