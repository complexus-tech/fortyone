import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getTeamObjectives = async (teamId: string, ctx: WorkspaceCtx) => {
  const objectives = await get<ApiResponse<Objective[]>>(
    `objectives?teamId=${teamId}`,
    ctx,
  );
  return objectives.data ?? [];
};
