import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoryAnalytics, AnalyticsFilters } from "../types";

export const getStoryAnalytics = async (
  ctx: WorkspaceCtx,
  filters?: AnalyticsFilters,
) => {
  const query = filters
    ? stringify(filters, {
        skipNulls: true,
        addQueryPrefix: true,
        encodeValuesOnly: true,
        arrayFormat: "comma",
      })
    : "";

  const analytics = await get<ApiResponse<StoryAnalytics>>(
    `analytics/story-analytics${query}`,
    ctx,
  );
  return analytics.data!;
};
