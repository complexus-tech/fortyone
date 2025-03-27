"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { DetailedStory } from "../types";

export const duplicateStoryAction = async (storyId: string) => {
  try {
    const story = await post<Record<string, never>, ApiResponse<DetailedStory>>(
      `stories/${storyId}/duplicate`,
      {},
    );
    return story;
  } catch (error) {
    return getApiError(error);
  }
};
