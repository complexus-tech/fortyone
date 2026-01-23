import { stringify } from "qs";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Contribution } from "@/types";
import type { SummaryFilters } from "@/modules/summary/types";

export const getContributions = async (
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
  const contributions = await get<ApiResponse<Contribution[]>>(
    `analytics/contributions${query}`,
    ctx,
  );
  return contributions.data!;
};
