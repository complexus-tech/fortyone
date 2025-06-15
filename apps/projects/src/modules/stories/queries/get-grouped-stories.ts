import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { GroupedStoriesResponse, GroupedStoryParams } from "../types";
import {
  buildGroupedStoriesQuery,
  getStoriesUrl,
} from "../utils/query-builders";

export const getGroupedStories = async (
  session: Session,
  params: GroupedStoryParams,
) => {
  const query = buildGroupedStoriesQuery(params);
  const url = getStoriesUrl("grouped");

  const response = await get<ApiResponse<GroupedStoriesResponse>>(
    `${url}${query}`,
    session,
  );

  return response.data!;
};
