import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SprintAnalytics } from "../types";

export const getSprintAnalytics = async (
  sprintId: string,
  ctx: WorkspaceCtx,
) => {
  const analytics = await get<ApiResponse<SprintAnalytics>>(
    `sprints/${sprintId}/analytics`,
    ctx,
  );
  return analytics.data!;
};
