import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { DetailedStory } from "../types";

export const getStory = async (id: string) => {
  const story = await get<ApiResponse<DetailedStory>>(`stories/${id}`);
  return story.data;
};
