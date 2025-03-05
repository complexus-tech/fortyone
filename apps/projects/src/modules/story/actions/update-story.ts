"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import type { DetailedStory } from "../types";

export const updateStoryAction = async (
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  try {
    await put(`stories/${storyId}`, payload);

    return storyId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update story");
  }
};
