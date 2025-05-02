"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type Payload = {
  storyIds: string[];
};

export const bulkDeleteAction = async (storyIds: string[]) => {
  try {
    const session = await auth();
    const stories = await remove<ApiResponse<Payload>>("stories", session!, {
      json: { storyIds },
    });
    return stories;
  } catch (error) {
    return getApiError(error);
  }
};
