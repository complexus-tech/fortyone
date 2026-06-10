import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getTeamObjectives = async (
  teamId: string,
  ctx: WorkspaceCtx,
  search = "",
) => {
  const params = new URLSearchParams({ teamId });
  if (search.trim()) {
    params.set("search", search.trim());
  }
  const objectives = await get<ApiResponse<Objective[]>>(
    `objectives?${params.toString()}`,
    ctx,
  );
  return objectives.data ?? [];
};
