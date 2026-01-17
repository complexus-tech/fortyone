import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ActivitiesResponse } from "../types";

export const getObjectiveActivities = async (
  objectiveId: string,
  ctx: WorkspaceCtx,
  page = 1,
  pageSize = 20,
) => {
  const response = await get<ApiResponse<ActivitiesResponse>>(
    `objectives/${objectiveId}/activities?page=${page}&pageSize=${pageSize}`,
    ctx,
  );
  return response.data!;
};
