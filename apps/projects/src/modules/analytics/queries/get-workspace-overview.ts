import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { WorkspaceOverview, AnalyticsFilters } from "../types";

export const getWorkspaceOverview = async (
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

  const overview = await get<ApiResponse<WorkspaceOverview>>(
    `analytics/overview${query}`,
    ctx,
  );
  return overview.data!;
};
