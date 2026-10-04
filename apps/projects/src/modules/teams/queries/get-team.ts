import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";
import { getApiError } from "@/utils";

/** A throwing read preserves authorization/not-found status for resource resolvers. */
export const getTeamDetails = async (
  id: string,
  ctx: WorkspaceCtx,
  signal?: AbortSignal,
) => {
  const response = await get<ApiResponse<Team>>(
    `teams/${id}`,
    ctx,
    signal ? { signal } : undefined,
  );
  return response.data ?? null;
};

export const getTeam = async (id: string, ctx: WorkspaceCtx) => {
  try {
    const data = await get<ApiResponse<Team>>(`teams/${id}`, ctx);
    return data;
  } catch (error) {
    return getApiError(error);
  }
};
