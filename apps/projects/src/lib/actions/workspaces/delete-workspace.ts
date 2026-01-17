"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteWorkspaceAction = async (
  workspaceSlug: string,
): Promise<void> => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    await remove<ApiResponse<void>>("", ctx);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete workspace");
  }
};
