import { get } from "@/lib/http";
import type { StoryActivity } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getStoryActivities = async (id: string) => {
  const story = await get<ApiResponse<StoryActivity[]>>(
    `stories/${id}/activities`,
  );
  return story.data;
};
