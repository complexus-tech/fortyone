"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteWorkspaceAction = async (): Promise<void> => {
  try {
    const session = await auth();
    await remove<ApiResponse<void>>("", session!);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete workspace");
  }
};
