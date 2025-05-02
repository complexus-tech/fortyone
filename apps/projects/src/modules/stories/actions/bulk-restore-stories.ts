"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

type Payload = {
  storyIds: string[];
};

export const bulkRestoreAction = async (storyIds: string[]) => {
  try {
    const session = await auth();
    const stories = await post<Payload, ApiResponse<Payload>>(
      "stories/restore",
      {
        storyIds,
      },
      session!,
    );

    return stories;
  } catch (error) {
    return getApiError(error);
  }
};
