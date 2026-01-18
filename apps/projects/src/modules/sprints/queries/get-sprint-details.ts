import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { SprintDetails } from "../types";

export const getSprint = async (sprintId: string, ctx: WorkspaceCtx) => {
  const sprint = await get<ApiResponse<SprintDetails>>(`sprints/${sprintId}`, ctx);
  return sprint.data;
};
