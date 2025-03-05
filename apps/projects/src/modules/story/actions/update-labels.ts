"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";

export const updateLabelsAction = async (storyId: string, labels: string[]) => {
  try {
    await put(`stories/${storyId}/labels`, { labels });
    return storyId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update labels");
  }
};
