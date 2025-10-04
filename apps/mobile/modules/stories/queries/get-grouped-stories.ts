import { stringify } from "qs";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { GroupedStoriesResponse, GroupedStoryParams } from "../types";

const buildGroupedStoriesQuery = (params: GroupedStoryParams) => {
  return stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
    arrayFormat: "comma",
  });
};

export const getGroupedStories = async (params: GroupedStoryParams) => {
  const query = buildGroupedStoriesQuery(params);
  const response = await get<ApiResponse<GroupedStoriesResponse>>(
    `stories/grouped${query}`
  );
  return response.data;
};
