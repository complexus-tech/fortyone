"use server";

import { put } from "@/lib/http";
import type { ApiResponse, Workspace } from "@/types";
import { getApiError } from "@/utils";

export type UpdateWorkspaceInput = {
  name: string;
};

export const updateWorkspaceAction = async (
  input: UpdateWorkspaceInput,
): Promise<Workspace> => {
  try {
    const workspace = await put<UpdateWorkspaceInput, ApiResponse<Workspace>>(
      "",
      input,
    );
    return workspace.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update workspace");
  }
};
