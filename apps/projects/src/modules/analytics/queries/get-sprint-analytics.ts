import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SprintAnalytics, AnalyticsFilters } from "../types";

export const getSprintAnalytics = async (
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

  const analytics = await get<ApiResponse<SprintAnalytics>>(
    `analytics/sprint-analytics${query}`,
    session,
  );
  return analytics.data!;
};
