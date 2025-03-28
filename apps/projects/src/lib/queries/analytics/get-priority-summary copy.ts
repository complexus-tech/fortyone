import { get } from "@/lib/http";
import type { ApiResponse, PrioritySummary } from "@/types";

export const getPrioritySummary = async () => {
  const summary = await get<ApiResponse<PrioritySummary>>("analytics/priority");
  return summary.data!;
};
