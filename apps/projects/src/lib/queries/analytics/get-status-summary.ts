import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, StatusSummary } from "@/types";

export const getStatusSummary = async (session: Session) => {
  const summary = await get<ApiResponse<StatusSummary[]>>(
    "analytics/status",
    session,
  );
  return summary.data!;
};
