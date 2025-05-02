"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const updateLabelsAction = async (storyId: string, labels: string[]) => {
  try {
    const session = await auth();
    const res = await put<{ labels: string[] }, ApiResponse<null>>(
      `stories/${storyId}/labels`,
      { labels },
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
