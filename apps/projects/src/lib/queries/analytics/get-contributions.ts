import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, Contribution } from "@/types";

export const getContributions = async (session: Session) => {
  const contributions = await get<ApiResponse<Contribution[]>>(
    "analytics/contributions",
    session,
  );
  return contributions.data!;
};
