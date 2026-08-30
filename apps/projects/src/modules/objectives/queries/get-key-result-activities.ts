import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ActivitiesResponse } from "../types";

export const getKeyResultActivities = async (
  keyResultId: string,
  page: number | undefined,
  pageSize: number | undefined,
  ctx: WorkspaceCtx,
) => {
  const resolvedPage = page === undefined ? 1 : page;
  const resolvedPageSize = pageSize === undefined ? 20 : pageSize;
  const response = await get<ApiResponse<ActivitiesResponse>>(
    `key-results/${keyResultId}/activities?page=${resolvedPage}&pageSize=${resolvedPageSize}`,
    ctx,
  );
  return response.data!;
};
