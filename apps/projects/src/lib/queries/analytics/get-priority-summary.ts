import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, PrioritySummary } from "@/types";

export const getPrioritySummary = async (ctx: WorkspaceCtx) => {
  const summary = await get<ApiResponse<PrioritySummary[]>>(
    "analytics/priority",
    ctx,
  );
  return summary.data!;
};
