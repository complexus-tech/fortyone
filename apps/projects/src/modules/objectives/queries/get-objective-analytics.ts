import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ObjectiveAnalytics } from "../types";

export const getObjectiveAnalytics = async (
  objectiveId: string,
  session: Session,
) => {
  const analytics = await get<ApiResponse<ObjectiveAnalytics>>(
    `objectives/${objectiveId}/analytics`,
    session,
  );
  return analytics.data!;
};
