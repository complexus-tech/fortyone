"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { DetailedStory } from "../types";

export const updateStoryAction = async (
  storyId: string,
  payload: Partial<DetailedStory>,
) => {
  try {
    const session = await auth();
    const res = await put<Partial<DetailedStory>, ApiResponse<null>>(
      `stories/${storyId}`,
      payload,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
