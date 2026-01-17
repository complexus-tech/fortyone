import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";
import { getApiError } from "@/utils";

export const getTeam = async (id: string, ctx: WorkspaceCtx) => {
  try {
    const data = await get<ApiResponse<Team>>(`teams/${id}`, ctx);
    return data;
  } catch (error) {
    return getApiError(error);
  }
};
