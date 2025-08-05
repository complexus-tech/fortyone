import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoryAnalytics, AnalyticsFilters } from "../types";

export const getStoryAnalytics = async (
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

  const analytics = await get<ApiResponse<StoryAnalytics>>(
    `analytics/story-analytics${query}`,
    session,
  );
  return analytics.data!;
};
