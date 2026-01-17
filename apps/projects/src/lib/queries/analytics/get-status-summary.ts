import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, StatusSummary } from "@/types";

export const getStatusSummary = async (ctx: WorkspaceCtx) => {
  const summary = await get<ApiResponse<StatusSummary[]>>(
    "analytics/status",
    ctx,
  );
  return summary.data!;
};
