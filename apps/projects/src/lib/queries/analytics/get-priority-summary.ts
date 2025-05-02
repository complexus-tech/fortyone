import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, PrioritySummary } from "@/types";

export const getPrioritySummary = async (session: Session) => {
  const summary = await get<ApiResponse<PrioritySummary[]>>(
    "analytics/priority",
    session,
  );
  return summary.data!;
};
