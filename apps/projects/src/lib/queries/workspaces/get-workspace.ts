import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Workspace } from "@/types";

export const getWorkspace = async (ctx: WorkspaceCtx) => {
  const workspace = await get<ApiResponse<Workspace>>("", ctx);
  return workspace.data!;
};
