"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type Payload = {
  storyIds: string[];
};

export const bulkRestoreAction = async (storyIds: string[]) => {
  try {
    const stories = await post<Payload, ApiResponse<Payload>>(
      "stories/restore",
      {
        storyIds,
      },
    );

    return stories;
  } catch (error) {
    return getApiError(error);
  }
};
