import { put } from "@/lib/http/fetch";
import type { ApiResponse } from "@/types";
import type { DetailedStory } from "@/modules/stories/types";

export const updateStoryAction = async (
  storyId: string,
  payload: Partial<DetailedStory>
) => {
  const response = await put<Partial<DetailedStory>, ApiResponse<null>>(
    `stories/${storyId}`,
    payload
  );
  return response;
};
