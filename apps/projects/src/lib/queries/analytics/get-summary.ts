import { stringify } from "qs";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, StoriesSummary } from "@/types";
import type { SummaryFilters } from "@/modules/summary/types";

export const getSummary = async (
  ctx: WorkspaceCtx,
  filters?: SummaryFilters,
) => {
  const query = filters
    ? stringify(filters, {
        skipNulls: true,
        addQueryPrefix: true,
        encodeValuesOnly: true,
        arrayFormat: "comma",
      })
    : "";
  const summary = await get<ApiResponse<StoriesSummary>>(
    `analytics/summary${query}`,
    ctx,
  );
  return summary.data!;
};
