import { get } from "@/lib/http";
import type { ApiResponse, StoriesSummary } from "@/types";

export const getSummary = async () => {
  const summary = await get<ApiResponse<StoriesSummary>>("analytics/summary");
  return summary.data!;
};
