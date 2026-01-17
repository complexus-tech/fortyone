import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SprintAnalytics, AnalyticsFilters } from "../types";

export const getSprintAnalytics = async (
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

  const analytics = await get<ApiResponse<SprintAnalytics>>(
    `analytics/sprint-analytics${query}`,
    ctx,
  );
  return analytics.data!;
};
