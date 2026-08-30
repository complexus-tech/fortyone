import { stringify } from "qs";
import type { RequestOptions } from "api-client";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { Story } from "@/modules/stories/types";
import type { ApiResponse } from "@/types";

export const getStories = async (
  ctx: WorkspaceCtx,
  params: {
    reporterId?: string;
    teamId?: string;
    sprintId?: string;
    objectiveId?: string;
    keyResultId?: string;
    epicId?: string;
    assigneeId?: string;
    showSubStories?: boolean;
  } = {},
  options?: RequestOptions,
) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });

  const stories = await get<ApiResponse<Story[]>>(
    `stories${query}`,
    ctx,
    options,
  );
  return stories.data!;
};
