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
    return story;
  } catch (error) {
    return getApiError(error);
  }
};
