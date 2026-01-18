import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State } from "@/types/states";

export const getTeamStatuses = async (teamId: string, ctx: WorkspaceCtx) => {
  const statuses = await get<ApiResponse<State[]>>(
    `states?teamId=${teamId}`,
    ctx,
  );
  return statuses.data!;
};
