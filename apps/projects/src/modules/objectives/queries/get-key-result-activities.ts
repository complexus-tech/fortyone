
import { get, WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ActivitiesResponse } from "../types";

export const getKeyResultActivities = async (
  keyResultId: string,
  page = 1,
  pageSize = 20,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<ActivitiesResponse>>(
    `key-results/${keyResultId}/activities?page=${page}&pageSize=${pageSize}`,
    ctx,
  );
  return response.data!;
};
