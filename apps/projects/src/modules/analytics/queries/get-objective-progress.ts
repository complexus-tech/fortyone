import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ObjectiveProgress, AnalyticsFilters } from "../types";

export const getObjectiveProgress = async (
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

  const progress = await get<ApiResponse<ObjectiveProgress>>(
    `analytics/objective-progress${query}`,
    ctx,
  );
  return progress.data!;
};
