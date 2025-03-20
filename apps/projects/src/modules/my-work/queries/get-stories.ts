import { get } from "@/lib/http";
import type { Story } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getMyStories = async () => {
  const stories = await get<ApiResponse<Story[]>>("my-stories");
  return stories.data!;
};
