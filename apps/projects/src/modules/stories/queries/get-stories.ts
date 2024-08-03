import { get } from "@/lib/http";
import { Story } from "@/modules/stories/types";
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
) => {
  const query = qs.stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
  });
  const stories = await get<Story[]>(`/stories${query}`);
  return stories;
};
