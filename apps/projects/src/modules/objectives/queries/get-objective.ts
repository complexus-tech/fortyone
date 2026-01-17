import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getObjective = async (objectiveId: string, ctx: WorkspaceCtx) => {
  const objective = await get<ApiResponse<Objective>>(
    `objectives/${objectiveId}`,
    ctx,
  );
  return objective.data;
};
