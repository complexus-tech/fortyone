import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { TimelineTrends, AnalyticsFilters } from "../types";

export const getTimelineTrends = async (
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

  const trends = await get<ApiResponse<TimelineTrends>>(
    `analytics/timeline-trends${query}`,
    session,
  );
  return trends.data!;
};
