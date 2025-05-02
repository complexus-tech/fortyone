"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { DetailedStory, NewStory } from "../types";

export const createStoryAction = async (newStory: NewStory) => {
  try {
    const session = await auth();
    const story = await post<NewStory, ApiResponse<DetailedStory>>(
      "stories",
      newStory,
      session!,
    );
    return story;
  } catch (error) {
    return getApiError(error);
  }
};
