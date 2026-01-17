import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, StoriesSummary } from "@/types";

export const getSummary = async (ctx: WorkspaceCtx) => {
  const summary = await get<ApiResponse<StoriesSummary>>(
    "analytics/summary",
    ctx,
  );
  return summary.data!;
};
