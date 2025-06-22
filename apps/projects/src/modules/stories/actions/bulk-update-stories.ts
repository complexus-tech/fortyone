"use server";

import { auth } from "@/auth";
import { put } from "@/lib/http";
import type { DetailedStory } from "@/modules/story/types";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type Payload = {
  storyIds: string[];
  updates: Partial<DetailedStory>;
};

export const bulkUpdateAction = async (updates: Payload) => {
  try {
    const session = await auth();
    const stories = await put<Payload, ApiResponse<null>>(
      "stories",
      updates,
      session!,
    );
    return stories;
  } catch (error) {
    return getApiError(error);
  }
};
