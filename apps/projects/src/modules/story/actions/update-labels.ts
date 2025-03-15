"use server";

import { patch } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const updateLabelsAction = async (storyId: string, labels: string[]) => {
  try {
    const res = await patch<{ labels: string[] }, ApiResponse<null>>(
      `stories/${storyId}/labels`,
      { labels },
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
