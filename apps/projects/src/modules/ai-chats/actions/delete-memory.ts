"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteMemoryAction = async (
  id: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const result = await remove<ApiResponse<null>>(`users/memory/${id}`, ctx);
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
