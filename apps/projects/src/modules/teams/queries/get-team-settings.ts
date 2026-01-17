import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { TeamSettings } from "../types";

export const getTeamSettings = async (
  teamId: string,
  ctx: WorkspaceCtx,
): Promise<TeamSettings> => {
  const settings = await get<ApiResponse<TeamSettings>>(
    `teams/${teamId}/settings`,
    ctx,
  );
  return settings.data!;
};
