"use server";

import { get } from "@/lib/http";
import type { StoryActivity } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getActivities = async () => {
  const activities = await get<ApiResponse<StoryActivity[]>>("activities");
  return activities.data!;
};
