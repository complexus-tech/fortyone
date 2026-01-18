"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const restoreStoryAction = async (
  storyId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<null, ApiResponse<null>>(
      `stories/${storyId}/restore`,
      null,
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
