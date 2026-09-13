import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoriesSummary } from "../types";

export const getOverviewStats = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<StoriesSummary>>("analytics/summary", {
    signal,
  });
  return response.data!;
};
