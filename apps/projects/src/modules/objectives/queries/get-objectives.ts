import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getObjectives = async (ctx: WorkspaceCtx) => {
  const objectives = await get<ApiResponse<Objective[]>>("objectives", ctx);
  return objectives.data ?? [];
};
