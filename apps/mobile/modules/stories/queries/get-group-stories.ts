import { stringify } from "qs";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { GroupStoriesResponse, GroupStoryParams } from "../types";

const buildGroupStoriesQuery = (params: GroupStoryParams): string => {
  return stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
    arrayFormat: "comma",
  });
};

export const getGroupStories = async (params: GroupStoryParams) => {
  const query = buildGroupStoriesQuery(params);
  const response = await get<ApiResponse<GroupStoriesResponse>>(
    `stories/group${query}`
  );
  return response.data;
};
