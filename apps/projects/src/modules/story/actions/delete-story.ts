"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteStoryAction = async (storyId: string) => {
  try {
    await remove(`stories/${storyId}`);
    return storyId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete story");
  }
};
