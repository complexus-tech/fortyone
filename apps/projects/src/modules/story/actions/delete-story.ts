import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteStoryAction = async (storyId: string) => {
  try {
    const res = await remove<ApiResponse<null>>(`stories/${storyId}`);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
