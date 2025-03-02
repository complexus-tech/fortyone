"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoryActivity } from "@/modules/stories/types";
import { getApiError } from "@/utils";

type Payload = {
  comment: string;
  parentId?: string | null;
};

export const commentStoryAction = async (storyId: string, payload: Payload) => {
  try {
    const activity = await post<Payload, ApiResponse<StoryActivity>>(
      `stories/${storyId}/comments`,
      payload,
    );
    return activity.data!;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to comment on story");
  }
};
