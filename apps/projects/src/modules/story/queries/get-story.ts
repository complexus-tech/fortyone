import { get } from "@/lib/http";
import { DetailedStory } from "../types";
import { ApiResponse } from "@/types";

export const getStory = async (id: string) => {
  const story = await get<ApiResponse<DetailedStory>>(`stories/${id}`);
  return story?.data;
};
