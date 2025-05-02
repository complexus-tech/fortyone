"use server";

import { put } from "@/lib/http";
import type { ApiResponse, Workspace } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type UpdateWorkspaceInput = {
  name: string;
};

export const updateWorkspaceAction = async (input: UpdateWorkspaceInput) => {
  try {
    const session = await auth();
    const workspace = await put<UpdateWorkspaceInput, ApiResponse<Workspace>>(
      "",
      input,
      session!,
    );
    return workspace;
  } catch (error) {
    return getApiError(error);
  }
};
