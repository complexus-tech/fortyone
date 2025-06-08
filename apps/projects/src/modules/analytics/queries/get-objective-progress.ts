import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ObjectiveProgress, AnalyticsFilters } from "../types";

export const getObjectiveProgress = async (
  filters?: AnalyticsFilters,
  session?: Session,
) => {
  if (!session) return null;

  const query = filters
    ? stringify(filters, {
        skipNulls: true,
        addQueryPrefix: true,
        encodeValuesOnly: true,
      })
    : "";

  const progress = await get<ApiResponse<ObjectiveProgress>>(
    `analytics/objective-progress${query}`,
    session,
  );
  return progress.data!;
};
