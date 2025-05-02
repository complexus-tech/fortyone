import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { StoryActivity } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getStoryActivities = async (id: string, session: Session) => {
  const story = await get<ApiResponse<StoryActivity[]>>(
    `stories/${id}/activities`,
    session,
  );
  return story.data!;
};
