import { auth } from "@/auth";
import { get } from "@/lib/http";
import { TAGS } from "@/constants/tags";
import { DURATION_FROM_SECONDS } from "@/constants/time";
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
  config?: RequestInit,
) => {
  const query = qs.stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
  });
  const stories = await get<Story[]>(`/stories${query}`, {
    next: {
      // revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [TAGS.stories, `${TAGS.stories}:${query}`],
    },
    ...config,
  });
  return stories;
};
