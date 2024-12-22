"use server";

import { post } from "@/lib/http";
import { ApiResponse } from "@/types";
import { StoryActivity } from "@/modules/stories/types";

type Payload = {
  comment: string;
  parentId?: string | null;
};

export const commentStoryAction = async (storyId: string, payload: Payload) => {
  const activity = await post<Payload, ApiResponse<StoryActivity>>(
    `stories/${storyId}/comments`,
    payload,
  );
  return activity.data!;
};
