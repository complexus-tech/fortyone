import { stringify } from "qs";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { StoryActivity } from "@/modules/stories/types";
import type { SummaryFilters } from "@/modules/summary/types";
import type { ApiResponse } from "@/types";

export const getActivities = async (
  ctx: WorkspaceCtx,
  filters?: SummaryFilters,
) => {
  const query = filters
    ? stringify(filters, {
        skipNulls: true,
        addQueryPrefix: true,
        encodeValuesOnly: true,
        arrayFormat: "comma",
      })
    : "";
  const activities = await get<ApiResponse<StoryActivity[]>>(
    `activities${query}`,
    ctx,
  );
  return activities.data!;
};
