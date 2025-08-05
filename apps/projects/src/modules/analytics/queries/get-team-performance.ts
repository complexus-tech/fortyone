import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { TeamPerformance, AnalyticsFilters } from "../types";

export const getTeamPerformance = async (
  filters?: AnalyticsFilters,
  session?: Session,
) => {
  if (!session) return null;

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
    session,
  );
  return performance.data!;
};
