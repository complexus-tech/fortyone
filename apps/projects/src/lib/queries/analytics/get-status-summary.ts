import { get } from "@/lib/http";
import type { ApiResponse, StatusSummary } from "@/types";

export const getStatusSummary = async () => {
  const summary = await get<ApiResponse<StatusSummary>>("analytics/status");
  return summary.data!;
};
