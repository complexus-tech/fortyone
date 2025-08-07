import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { StoryActivity } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

type ActivitiesResponse = {
  activities: StoryActivity[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
  };
};

export const getStoryActivities = async (
  id: string,
  session: Session,
  page: number = 1,
) => {
  const response = await get<ApiResponse<ActivitiesResponse>>(
    `stories/${id}/activities?page=${page}`,
    session,
  );
  return response.data!;
};
