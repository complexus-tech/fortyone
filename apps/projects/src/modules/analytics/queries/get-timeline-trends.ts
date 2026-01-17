import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { TimelineTrends, AnalyticsFilters } from "../types";

export const getTimelineTrends = async (
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

  const trends = await get<ApiResponse<TimelineTrends>>(
    `analytics/timeline-trends${query}`,
    ctx,
  );
  return trends.data!;
};
