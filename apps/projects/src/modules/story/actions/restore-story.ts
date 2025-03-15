"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const restoreStoryAction = async (storyId: string) => {
  try {
    const res = await post<null, ApiResponse<null>>(
      `stories/${storyId}/restore`,
      null,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
