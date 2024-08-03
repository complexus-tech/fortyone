import { get } from "@/lib/http";
import { Story } from "@/modules/stories/types";

export const getMyStories = async () => {
  const stories = await get<Story[]>("/my-stories");
  return stories;
};
