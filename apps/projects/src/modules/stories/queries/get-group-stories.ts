import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { GroupStoriesResponse, GroupStoryParams } from "../types";
import { buildGroupStoriesQuery, getStoriesUrl } from "../utils/query-builders";

export const getGroupStories = async (
  session: Session,
  params: GroupStoryParams,
) => {
  const query = buildGroupStoriesQuery(params);
  const url = getStoriesUrl("group");

  const response = await get<ApiResponse<GroupStoriesResponse>>(
    `${url}${query}`,
    session,
  );

  return response.data!;
};
