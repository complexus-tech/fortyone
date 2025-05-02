import type { Options } from "ky";
import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { Story } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getStories = async (
  session: Session,
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
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });

  const stories = await get<ApiResponse<Story[]>>(
    `stories${query}`,
    session,
    options,
  );
  return stories.data!;
};
