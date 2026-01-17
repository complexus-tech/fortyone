import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { TeamPerformance, AnalyticsFilters } from "../types";

export const getTeamPerformance = async (
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

  const performance = await get<ApiResponse<TeamPerformance>>(
    `analytics/team-performance${query}`,
    ctx,
  );
  return performance.data!;
};
