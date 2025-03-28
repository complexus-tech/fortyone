import { get } from "@/lib/http";
import type { ApiResponse, Contribution } from "@/types";

export const getContributions = async () => {
  const contributions = await get<ApiResponse<Contribution[]>>(
    "analytics/contributions",
  );
  return contributions.data!;
};
