"use server";

import { put } from "@/lib/http";
import type { ApiResponse, Workspace } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export type UpdateWorkspaceInput = {
  name?: string;
  avatarUrl?: string;
};

export const updateWorkspaceAction = async (
  input: UpdateWorkspaceInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const workspace = await put<UpdateWorkspaceInput, ApiResponse<Workspace>>(
      "",
      input,
      ctx,
    );
    return workspace;
  } catch (error) {
    return getApiError(error);
  }
};
