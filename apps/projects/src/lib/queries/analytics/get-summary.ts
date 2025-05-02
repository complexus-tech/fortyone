import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, StoriesSummary } from "@/types";

export const getSummary = async (session: Session) => {
  const summary = await get<ApiResponse<StoriesSummary>>(
    "analytics/summary",
    session,
  );
  return summary.data!;
};
