import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getTotalStories = async () => {
  try {
    const stories = await get<ApiResponse<{ count: number }>>("stories/count");
    return stories.data!.count;
  } catch {
    return 0;
  }
};
