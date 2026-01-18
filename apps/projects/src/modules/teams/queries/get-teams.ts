import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "../types";

export const getTeams = async (ctx: WorkspaceCtx): Promise<Team[]> => {
  const response = await get<ApiResponse<Team[]>>("teams", ctx);
  return response.data!;
};
