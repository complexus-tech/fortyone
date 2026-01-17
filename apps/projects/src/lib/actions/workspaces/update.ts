"use server";

import { put } from "@/lib/http";
import type { ApiResponse, WorkspaceSettings } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type UpdateWorkspaceSettings = Partial<WorkspaceSettings>;

export const updateWorkspaceSettingsAction = async (
  payload: UpdateWorkspaceSettings,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await put<
      UpdateWorkspaceSettings,
      ApiResponse<WorkspaceSettings>
    >("settings", payload, ctx);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
