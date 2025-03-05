"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { DetailedStory, NewStory } from "../types";

export const createStoryAction = async (newStory: NewStory) => {
  try {
    const story = await post<NewStory, ApiResponse<DetailedStory>>(
      "stories",
      newStory,
    );
    return story.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create story");
  }
};
