import { get } from "@/lib/http";
import { Story } from "@/modules/stories/types";
import { ApiResponse } from "@/types";
import { Options } from "ky";
import qs from "qs";

export const getStories = async (
  params: {
    reporterId?: string;
    teamId?: string;
    sprintId?: string;
    objectiveId?: string;
    epicId?: string;
    assigneeId?: string;
  } = {},
  options?: Options,
) => {
  const query = qs.stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const stories = await get<ApiResponse<Story[]>>(`stories${query}`, options);
  return stories?.data;
};
