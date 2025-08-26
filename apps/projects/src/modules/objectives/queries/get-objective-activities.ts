import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ActivitiesResponse } from "../types";

export const getObjectiveActivities = async (
  objectiveId: string,
  session: Session,
  page = 1,
  pageSize = 20,
) => {
  const response = await get<ApiResponse<ActivitiesResponse>>(
    `objectives/${objectiveId}/activities?page=${page}&pageSize=${pageSize}`,
    session,
  );
  return response.data!;
};
