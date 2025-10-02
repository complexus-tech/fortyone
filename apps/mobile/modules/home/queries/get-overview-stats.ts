import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoriesSummary } from "../types";

export const getOverviewStats = async () => {
  const response = await get<ApiResponse<StoriesSummary>>("analytics/summary");
  return response.data!;
};
