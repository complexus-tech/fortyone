"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type Payload = {
  storyIds: string[];
};

export const bulkDeleteAction = async (storyIds: string[]) => {
  try {
    const stories = await remove<ApiResponse<Payload>>("stories", {
      json: { storyIds },
    });
    return stories.data;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete stories");
  }
};
