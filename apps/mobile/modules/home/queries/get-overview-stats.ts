import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { OverviewStats } from "../types";

export const getOverviewStats = async (): Promise<OverviewStats> => {
  try {
    const response =
      await get<ApiResponse<OverviewStats>>("analytics/overview");
    return response.data!;
  } catch (error) {
    console.error(error);
    throw error;
  }
};
