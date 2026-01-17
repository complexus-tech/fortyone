import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State } from "@/types/states";

export const getStatuses = async (ctx: WorkspaceCtx) => {
  const statuses = await get<ApiResponse<State[]>>("states", ctx);
  return statuses.data!;
};
