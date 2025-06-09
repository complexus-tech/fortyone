import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SprintAnalytics } from "../types";

export const getSprintAnalytics = async (
  sprintId: string,
  session: Session,
) => {
  const analytics = await get<ApiResponse<SprintAnalytics>>(
    `sprints/${sprintId}/analytics`,
    session,
  );
  return analytics.data!;
};
