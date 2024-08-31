import { get } from "@/lib/http";
import { DetailedStory } from "../types";

export const getStory = async (id: string) => {
  const story = await get<DetailedStory>(`stories/${id}`);
  return story;
};
