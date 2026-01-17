import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "../types";

export const getPublicTeams = async (ctx: WorkspaceCtx): Promise<Team[]> => {
  const response = await get<ApiResponse<Team[]>>("teams/public", ctx);
  return response.data!;
};
