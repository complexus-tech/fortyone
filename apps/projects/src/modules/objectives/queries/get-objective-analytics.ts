import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { ObjectiveAnalytics } from "../types";

export const getObjectiveAnalytics = async (
  objectiveId: string,
  ctx: WorkspaceCtx,
) => {
  const analytics = await get<ApiResponse<ObjectiveAnalytics>>(
    `objectives/${objectiveId}/analytics`,
    ctx,
  );
  return analytics.data!;
};
