import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Sprint } from "../types";

export const getSprints = async (ctx: WorkspaceCtx) => {
  const sprints = await get<ApiResponse<Sprint[]>>("sprints", ctx);
  return sprints.data!;
};
