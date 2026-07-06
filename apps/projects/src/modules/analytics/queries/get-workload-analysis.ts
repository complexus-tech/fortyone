import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AnalyticsFilters, WorkloadAnalysis } from "../types";

export const getWorkloadAnalysis = async (
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

  const analysis = await get<ApiResponse<WorkloadAnalysis>>(
    `analytics/workload-analysis${query}`,
    ctx,
  );
  return analysis.data!;
};
