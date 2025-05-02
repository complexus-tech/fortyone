"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteStoryAction = async (storyId: string) => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<null>>(`stories/${storyId}`, session!);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
