"use server";

import { post } from "@/lib/http";
import { getApiError } from "@/utils";

export const restoreStoryAction = async (storyId: string) => {
  try {
    await post(`stories/${storyId}/restore`, {});
    return storyId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to restore story");
  }
};
