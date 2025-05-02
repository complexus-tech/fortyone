import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { StoryActivity } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getActivities = async (session: Session) => {
  const activities = await get<ApiResponse<StoryActivity[]>>(
    "activities",
    session,
  );
  return activities.data!;
};
